package checker

import (
	"fmt"
	"log/slog"
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
	switch service.Type {
	case "tcp":
		err = CheckTCP(service.TCP)
	case "web_request":
		err = CheckWebRequest(service.WebRequest)
	case "bash_script":
		err = CheckBashScript(service.BashScript)
	default:
		err = fmt.Errorf("unknown action type: %s", service.Type)
	}

	oldStatus := ac.config.ServicesStatus.Get(serviceName)
	status := oldStatus

	isActive := err == nil
	if status.IsActive != isActive {
		status.IsActive = isActive
		status.ConsecutiveFails = 0
		status.Notified = false
		status.LastChange = time.Now()
	}
	if status.Notified {
		return
	}
	defer func() {
		if !ac.config.ServicesStatus.CompareAndSwap(serviceName, oldStatus, status) {
			slog.Warn("error saving current status", "error", err, "service", serviceName)
		}
	}()

	if isActive {
		go ac.notifier.Notify(fmt.Sprintf(templateTitle, serviceName, "up"), "ok")
		status.Notified = true
		return
	}
	status.ConsecutiveFails++
	status.LastChange = time.Now()
	if status.ConsecutiveFails >= service.MaxFails {
		go ac.notifier.Notify(fmt.Sprintf(templateTitle, serviceName, "down"), "down")
		status.Notified = true
	}
}
