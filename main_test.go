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
	testMainSkel(t, "test_files/offline_test.yaml", "468c", true)
}

func TestOnline(t *testing.T) {
	testMainSkel(t, "test_files/online_test.yaml", "5189f8b68749a3e6a26e6ea2ed0a91947e56b76f8e1f73b2d600eaf3cc732eb7", false)
}

func TestTelegramNotification(t *testing.T) {
	t.SkipNow() //comment this line to run this test
	testMainSkel(t, "test_files/telegram_test.yaml", "", true)
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
		if "auth check ok" != string(bodyBytes) {
			t.Errorf("unexpected response body for auth request: %s", string(bodyBytes))
		}
	}

	time.Sleep(5 * time.Second)
}
