package main

import (
	"io"
	"log/slog"
	"net/http"
	"os"
	"testing"
	"time"
)

func TestOffline(t *testing.T) {
	testMainSkel(t, "test_configs/offline_test.yaml", "462a", true)
}

func TestTelegramNotification(t *testing.T) {
	t.SkipNow() //comment this line to run this test
	testMainSkel(t, "test_configs/telegram_test.yaml", "", true)
}

func testMainSkel(t *testing.T, configFile, expectedHash string, includeAuthHost bool) {
	// Injecting parameters directly as a string slice
	args := []string{
		"-config", configFile,
		"-hash", expectedHash, "-v",
	}

	go companionRoutine(t, includeAuthHost)

	exitCode := run(args)

	if exitCode != 0 {
		t.Errorf("expected exit code 0, got %d", exitCode)
	}
}

func companionRoutine(t *testing.T, includeAuthHost bool) {
	defer func() {
		slog.Info("requesting to exit from companion routine")
		SignalChan <- os.Interrupt
	}()

	if includeAuthHost {
		time.Sleep(1 * time.Second)
		resp, err := http.Get("http://127.0.0.1:12345/?key=12345678")
		if err != nil {
			t.Errorf("error sending request: %v", err)
			return
		}
		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Errorf("error reading response body: %v", err)
		}
		defer resp.Body.Close()
		slog.Info("response body", "body", string(bodyBytes))
		if "auth check ok" != string(bodyBytes) {
			t.Errorf("unexpected response body for auth request: %s", string(bodyBytes))
		}
	}

	time.Sleep(5 * time.Second)
}

func TestPortInUse(t *testing.T) {
	go func() {
		err := http.ListenAndServe("127.0.0.1:12345", http.NewServeMux())
		if err != nil {
			t.Errorf("error starting server: %v", err)
		}
	}()
	args := []string{"-config", "test_configs/portinuse_test.yaml", "-v"}
	exitCode := run(args)
	if exitCode != 1 {
		t.Errorf("expected exit code 1, got %d", exitCode)
	}
}
