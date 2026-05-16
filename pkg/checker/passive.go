package checker

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"regexp"
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
	slog.Debug("starting passive checker")
	defer pc.wg.Done()
	mux := http.NewServeMux()
	mux.HandleFunc("/", pc.httpPassiveEndpoint)
	srv := &http.Server{
		Addr:              pc.config.ListenAddress,
		Handler:           mux,
		ReadHeaderTimeout: 1 * time.Second,
		ReadTimeout:       1 * time.Second,
		WriteTimeout:      2 * time.Second,
		IdleTimeout:       2 * time.Second,
	}
	pc.exitWG.Add(1)
	var err error
	go func() {
		err = srv.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			pc.err = err
			slog.Error("error running passive checker: sending interrupt signal")
			process, err := os.FindProcess(os.Getpid())
			if err != nil {
				slog.Error("error finding process", "error", err)
			}
			err = process.Signal(os.Interrupt)
			if err != nil {
				slog.Error("error sending interrupt signal", "error", err)
			}
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

var remoteAddressPortRegexp = regexp.MustCompile(`:[0-9]+$`)

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
	remoteIP := remoteAddressPortRegexp.ReplaceAllString(r.RemoteAddr, "")
	pc.updateOnlineHostEntry(authHost, remoteIP)
}

func (pc *PassiveChecker) updateOnlineHostEntry(host, remoteAddress string) {
	oldStatus := pc.config.HostsStatus.Get(host)
	status := oldStatus
	status.LastCheck = time.Now()
	if !status.IsActive {
		status.IsActive = true
		go pc.notifier.Notify(status)
	}
	if status.HostAddress != remoteAddress {
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
		go pc.notifier.Notify(status)
	} else {
		slog.Warn("error setting host as inactive", "host", host)
	}
}
