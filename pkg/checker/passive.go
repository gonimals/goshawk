package checker

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"sync"
	"time"

	"github.com/gonimals/goshawk/pkg/config"
	"github.com/gonimals/goshawk/pkg/notifier"
)

type PassiveChecker struct {
	baseDaemon
	exitWG *sync.WaitGroup
}

func NewPassiveChecker(config *config.Config, notifier notifier.Notifier) *PassiveChecker {
	output := &PassiveChecker{
		baseDaemon: baseDaemon{
			wg:           &sync.WaitGroup{},
			shutdownChan: make(chan bool),
			config:       config,
			notifier:     notifier,
		},
		exitWG: &sync.WaitGroup{},
	}
	go output.run()
	output.wg.Add(1)
	return output
}

func (pc *PassiveChecker) run() error {
	defer pc.wg.Done()
	mux := http.NewServeMux()
	mux.HandleFunc("/", pc.httpPassiveEndpoint)
	srv := &http.Server{
		Addr:    pc.config.ListenAddress,
		Handler: mux,
	}
	pc.exitWG.Add(1)
	var err error
	go func() {
		err = srv.ListenAndServe()
		if err == http.ErrServerClosed {
			err = nil
		}
		pc.exitWG.Done()
	}()

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

runLoop:
	for range ticker.C {
		select {
		case <-pc.shutdownChan:
			ctx, cancel := context.WithTimeout(context.Background(), 0)
			srv.Shutdown(ctx)
			cancel()
			break runLoop

		default:
			for host := range pc.config.AuthenticatedHosts {
				go pc.checkHostOnline(host)
			}
		}
	}
	pc.exitWG.Wait()
	return err
}

func (pc *PassiveChecker) httpPassiveEndpoint(w http.ResponseWriter, r *http.Request) {
	providedKey := r.URL.Query().Get("key")
	if providedKey == "" {
		fmt.Fprint(w, "check ok")
		slog.Debug("http plain check", "host", r.RemoteAddr)
		return
	}
	authHost := pc.config.HostKeys[providedKey]
	if authHost == "" {
		fmt.Fprintf(w, "auth failed")
		slog.Debug("http auth failed", "host", r.RemoteAddr)
		return
	}
	slog.Info("host online request", "host", authHost)

	fmt.Fprint(w, "auth check ok")
	slog.Debug("http auth success", "host", authHost)
	go pc.updateHostEntry(authHost, r.RemoteAddr)
}

func (pc *PassiveChecker) updateHostEntry(host, remoteAddress string) {
	oldStatus := pc.config.HostsStatus.Get(host)
	status := oldStatus
	status.LastCheck = time.Now()
	if !status.IsActive || status.HostAddress != remoteAddress {
		go pc.notifier.Notify(fmt.Sprintf(templateTitle, host, "up"), "up with address "+remoteAddress)
		status.Notified = true
		status.IsActive = true
		status.HostAddress = remoteAddress
	}
	if !pc.config.HostsStatus.CompareAndSwap(host, oldStatus, status) {
		slog.Warn("error saving current status", "host", host)
	}
}

func (pc *PassiveChecker) checkHostOnline(host string) {
	oldStatus := pc.config.HostsStatus.Get(host)
	if !oldStatus.IsActive {
		return
	}
	if time.Since(oldStatus.LastCheck) < time.Duration(pc.config.HostMaxSeconds)*time.Second {
		return
	}
	status := oldStatus
	slog.Debug("passive checker detected host down", "host", host, "lastCheck", oldStatus.LastCheck)
	status.IsActive = false
	if pc.config.HostsStatus.CompareAndSwap(host, oldStatus, status) {
		go pc.notifier.Notify(fmt.Sprintf(templateTitle, host, "down"), "down")
	} else {
		slog.Warn("error setting host as inactive", "host", host)
	}
}
