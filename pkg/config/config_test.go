package config_test

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
	"testing"

	"github.com/gonimals/goshawk/pkg/config"
)

const testConfigPath = "../../example_config.yaml"
const testConfigNotificationURL = "https://api.telegram.org/bot<YOUR_BOT_TOKEN>/sendMessage"

func TestLoadConfig(t *testing.T) {
	data, err := os.ReadFile(testConfigPath)
	if err != nil {
		t.Fatalf("Failed to read test config file: %v", err)
	}

	hash := sha256.Sum256(data)
	expectedHash := hex.EncodeToString(hash[:])

	cfg, err := config.LoadConfig(testConfigPath, expectedHash)
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	if cfg.NotificationURL != testConfigNotificationURL {
		t.Errorf("Expected URL %s, got %s", testConfigNotificationURL, cfg.NotificationURL)
	}

	if cfg.HostKeys["secret_passphrase"] != "remote_server" {
		t.Errorf("Expected host1 for key1, got %s", cfg.HostKeys["secret_passphrase"])
	}

	if val := cfg.HostsStatus.Get("remote_server"); val.HostAddress != "" || val.ConsecutiveFails != 0 {
		t.Errorf("Unexpected default asset status: %+v", val)
	}
}

func TestLoadConfigMismatchHash(t *testing.T) {
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "config.yaml")
	if err := os.WriteFile(configFile, []byte("services:\n  dummy:\n    type: tcp"), 0644); err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	_, err := config.LoadConfig(configFile, "wronghash")
	if err == nil {
		t.Errorf("Expected error for mismatched hash, got nil")
	}
}

func TestLoadConfigNoHash(t *testing.T) {
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "config.yaml")
	if err := os.WriteFile(configFile, []byte("services:\n  dummy:\n    type: tcp"), 0644); err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	_, err := config.LoadConfig(configFile, "")
	if err != nil {
		t.Fatalf("LoadConfig with empty hash failed: %v", err)
	}
}
