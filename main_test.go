package main

import (
	"io"
	"log/slog"
	"net/http"
	"testing"
	"time"

	"github.com/gonimals/goshawk/pkg/checker"
	"github.com/gonimals/goshawk/pkg/config"
	"github.com/gonimals/goshawk/pkg/notifier"
)

func TestOffline(t *testing.T) {
	completeTest(t, "test_files/offline_test.json", "f3ebcffc6a0fed8b386fa79ab3a5b27114f999a0fe4d7abccde842c8b6a60b69", true)
}

func TestOnline(t *testing.T) {
	completeTest(t, "test_files/online_test.json", "dca94de2b91a98be29", false)
}

func completeTest(t *testing.T, configFile, expectedHash string, includeAuthHost bool) {
	slog.SetLogLoggerLevel(slog.LevelDebug)
	config, err := config.LoadConfig(configFile, expectedHash)
	if err != nil {
		t.Fatalf("Error loading configuration: %v", err)
	}

	sender := notifier.NewTestNotifier()
	activeChecker := checker.NewActiveChecker(config, sender)
	passiveChecker := checker.NewPassiveChecker(config, sender)

	if includeAuthHost {
		time.Sleep(1 * time.Second)
		resp, err := http.Get("http://127.0.0.1:12345/?key=12345678")
		if err != nil {
			t.Fatalf("error sending request: %v", err)
		}
		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Fatalf("error reading response body: %v", err)
		}
		if "auth check ok" != string(bodyBytes) {
			t.Fatalf("unexpected response body for auth request: %s", string(bodyBytes))
		}
	}

	time.Sleep(5 * time.Second)
	slog.Debug("shutting down debug checkers")
	err = activeChecker.Stop()
	if err != nil {
		t.Fatalf("error stopping active checker: %v", err)
	} else {
		slog.Debug("active checker stopped")
	}
	err = passiveChecker.Stop()
	if err != nil {
		t.Fatalf("error stopping passive checker: %v", err)
	} else {
		slog.Debug("passive checker stopped")
	}
	slog.Info("exiting")
}
