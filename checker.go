package main

import (
	"fmt"
	"log"
	"net/http"
	"os/exec"
	"regexp"
	"time"

	"github.com/go-ping/ping" // TODO: replace this with TCP test
)

/*
conn, err := net.DialTimeout("tcp", "8.8.8.8:80", time.Second*2)
if err == nil {
    conn.Close()
    // Host is up!
}
*/

func checkService(service Service, notificationURL string) {
	var err error
	switch service.Action.Type {
	case "ping":
		err = checkPing(service.Action.Ping)
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

func checkPing(action *PingAction) error {
	pinger, err := ping.NewPinger(action.Host)
	if err != nil {
		return err
	}
	pinger.Count = 3
	pinger.Timeout = time.Second * 5
	pinger.SetPrivileged(true) // this may require root privileges
	err = pinger.Run()
	if err != nil {
		return err
	}
	stats := pinger.Statistics()
	if stats.PacketsRecv == 0 {
		return fmt.Errorf("no packets received")
	}
	return nil
}

func checkWebRequest(action *WebRequestAction) error {
	req, err := http.NewRequest(action.Method, action.URL, nil)
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
