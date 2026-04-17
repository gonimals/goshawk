package main

import (
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os/exec"
	"regexp"
	"strings"
	"time"
)

func activeCheckerRoutine() {
	defer globalWaitGroup.Done()
	for serviceName := range runtimeConfig.Services {
		serviceStatus.Set(serviceName, &ServiceStatus{})
	}
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for range ticker.C {
		if gracefulShutdown {
			return
		}
		for serviceName := range runtimeConfig.Services {
			go checkService(serviceName)
		}
	}
}

func checkService(serviceName string) {
	var err error
	service := runtimeConfig.Services[serviceName]
	switch service.Type {
	case "tcp":
		err = checkTCP(service.TCP)
	case "web_request":
		err = checkWebRequest(service.WebRequest)
	case "bash_script":
		err = checkBashScript(service.BashScript)
	default:
		err = fmt.Errorf("unknown action type: %s", service.Type)
	}

	statusMutex.Lock()
	defer statusMutex.Unlock()

	status := serviceStatus.Get(serviceName)

	isActive := err == nil
	if status.isActive != isActive {
		status.isActive = isActive
		status.consecutiveFails = 0
		status.notified = false
		status.lastChange = time.Now()
	}
	if status.notified {
		return
	}
	if isActive {
		go Notify(serviceName)
		status.notified = true
		return
	}
	status.consecutiveFails++
	status.lastChange = time.Now()
	if status.consecutiveFails >= service.MaxFails {
		go Notify(serviceName)
		status.notified = true
	}
}

func checkTCP(action *TCPAction) error {
	defer func() {
		if recovered := recover(); recovered != nil {
			slog.Error("panic running check", "recovered", recovered)
		}
	}()
	conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%d", action.Host, action.Port), time.Second*2)
	if err == nil {
		conn.Close()
	}
	return err
}

func checkWebRequest(action *WebRequestAction) error {
	req, err := http.NewRequest(action.Method, action.URL, strings.NewReader(action.Body))
	if err != nil {
		return err
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != action.ExpectedStatus {
		return fmt.Errorf("unexpected status code: got %d, want %d", resp.StatusCode, action.ExpectedStatus)
	}

	return nil
}

func checkBashScript(action *BashScriptAction) error {
	cmd := exec.Command("bash", "-c", action.Code)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("script execution failed: %w - %s", err, output)
	}

	re, err := regexp.Compile(action.ExpectedOutputRegexp)
	if err != nil {
		return fmt.Errorf("invalid regexp: %w", err)
	}

	if !re.Match(output) {
		return fmt.Errorf("output does not match regexp: %s", output)
	}

	return nil
}
