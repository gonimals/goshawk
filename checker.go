package main

import (
	"fmt"
	"log"
	"log/slog"
	"net"
	"net/http"
	"os/exec"
	"regexp"
	"strings"
	"time"
)

func checkService(service Service, notificationURL string) {
	var err error
	switch service.Action.Type {
	case "tcp":
		err = checkTCP(service.Action.TCP)
	case "web_request":
		err = checkWebRequest(service.Action.WebRequest)
	case "bash_script":
		err = checkBashScript(service.Action.BashScript)
	default:
		err = fmt.Errorf("unknown action type: %s", service.Action.Type)
	}

	if err != nil {
		log.Printf("Service %s check failed: %v", service.Name, err)
		notify(notificationURL, service.Name, err.Error())
	} else {
		log.Printf("Service %s check ok", service.Name)
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
