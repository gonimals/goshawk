package checker

import (
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"os/exec"
	"strings"
	"sync"
	"time"

	"github.com/gonimals/goshawk/pkg/config"
	"github.com/gonimals/goshawk/pkg/notifier"
)

type ActiveChecker struct {
	baseDaemon
}

func NewActiveChecker(config *config.Config, notifier notifier.Notifier) *ActiveChecker {
	output := &ActiveChecker{
		baseDaemon: baseDaemon{
			wg:           &sync.WaitGroup{},
			shutdownChan: make(chan bool),
			config:       config,
			notifier:     notifier,
		},
	}
	go output.run()
	output.wg.Add(1)
	return output
}

func (ac *ActiveChecker) run() {
	defer ac.wg.Done()
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for range ticker.C {
		select {
		case <-ac.shutdownChan:
			return
		default:
			//default will go through without blocking
		}
		for serviceName := range ac.config.Services {
			go ac.checkService(serviceName)
		}
	}
}

func (ac *ActiveChecker) checkService(serviceName string) {
	var err error
	service := ac.config.Services[serviceName]
	if service.Mutex.TryLock() {
		defer service.Mutex.Unlock()
	} else {
		// avoid being too verbose
		//slog.Debug("service already being checked", "name", serviceName)
		return
	}

	if service.FrequencySeconds > 0 &&
		time.Since(service.Status.LastCheck) < time.Duration(service.FrequencySeconds)*time.Second {
		return
	}
	slog.Debug("checking service", "name", serviceName)

	switch service.Type {
	case "tcp":
		err = ac.CheckTCP(service.TCP)
	case "web_request":
		err = ac.CheckWebRequest(service.WebRequest)
	case "bash_script":
		err = ac.CheckBashScript(service.BashScript)
	default:
		err = fmt.Errorf("unknown action type: %s", service.Type)
	}

	isActive := err == nil
	slog.Debug("service checked", "name", serviceName, "active", isActive, "error", err)
	if service.Status.IsActive != isActive {
		service.Status.IsActive = isActive
		service.Status.ConsecutiveFails = 0
		// Avoid notification if back to online before max consecutive fails
		service.Status.Notified = !service.Status.Notified && service.Status.IsActive
		service.Status.LastChange = time.Now()
		if err != nil {
			service.Status.DownReason = err.Error()
		} else {
			service.Status.DownReason = ""
		}
	}
	if service.Status.Notified {
		service.Status.LastCheck = time.Now()
		return
	}

	if isActive {
		service.Status.Notified = true
		service.Status.LastCheck = time.Now()
		go ac.notifier.Notify(service.Status)
		return
	}
	service.Status.ConsecutiveFails++
	if service.Status.ConsecutiveFails < service.MaxFails {
		return // the check will run again immediately
	}
	service.Status.Notified = true
	service.Status.LastCheck = time.Now()
	go ac.notifier.Notify(service.Status)
}

func (ac *ActiveChecker) CheckTCP(action *config.TCPAction) error {
	defer func() {
		if recovered := recover(); recovered != nil {
			slog.Error("panic running check", "recovered", recovered)
		}
	}()
	timeoutSeconds := action.TimeoutSeconds
	if timeoutSeconds == 0 {
		timeoutSeconds = ac.config.DefaultServiceTimeout
	}
	conn, err := net.DialTimeout("tcp", action.Address, time.Second*time.Duration(timeoutSeconds))
	if err == nil {
		conn.Close()
	}
	return err
}

func (ac *ActiveChecker) CheckWebRequest(action *config.WebRequestAction) error {
	req, err := http.NewRequest(action.Method, action.URL, strings.NewReader(action.Body))
	if err != nil {
		return err
	}

	timeoutSeconds := action.TimeoutSeconds
	if timeoutSeconds == 0 {
		timeoutSeconds = ac.config.DefaultServiceTimeout
	}
	client := &http.Client{
		Timeout: time.Second * time.Duration(timeoutSeconds),
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != action.ExpectedStatus {
		return fmt.Errorf("unexpected status code: got %d, want %d", resp.StatusCode, action.ExpectedStatus)
	}

	if action.ExpectedOutput == "" {
		return nil
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if !action.ExpectedOutputRegexp.Match(bodyBytes) {
		return fmt.Errorf("output does not match regexp: %s", bodyBytes)
	}

	return nil
}

func (ac *ActiveChecker) CheckBashScript(action *config.BashScriptAction) error {
	cmd := exec.Command("bash", "-c", action.Code)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("script execution failed: %w - %s", err, output)
	}

	if !action.ExpectedOutputRegexp.Match(output) {
		return fmt.Errorf("output does not match regexp: %s", output)
	}

	return nil
}
