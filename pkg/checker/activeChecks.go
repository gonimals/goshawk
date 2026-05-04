package checker

import (
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os/exec"
	"regexp"
	"strings"
	"time"

	"github.com/gonimals/goshawk/pkg/config"
)

func CheckTCP(action *config.TCPAction) error {
	defer func() {
		if recovered := recover(); recovered != nil {
			slog.Error("panic running check", "recovered", recovered)
		}
	}()
	timeoutSeconds := action.TimeoutSeconds
	if timeoutSeconds == 0 {
		timeoutSeconds = 2
	}
	conn, err := net.DialTimeout("tcp", action.Address, time.Second*time.Duration(timeoutSeconds))
	if err == nil {
		conn.Close()
	}
	return err
}

func CheckWebRequest(action *config.WebRequestAction) error {
	req, err := http.NewRequest(action.Method, action.URL, strings.NewReader(action.Body))
	if err != nil {
		return err
	}

	timeoutSeconds := action.TimeoutSeconds
	if timeoutSeconds == 0 {
		timeoutSeconds = 10
	}
	client := &http.Client{Timeout: time.Second * time.Duration(timeoutSeconds)}
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

func CheckBashScript(action *config.BashScriptAction) error {
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
