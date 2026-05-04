package config_test

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
	"testing"

	"github.com/gonimals/goshawk/pkg/config"
	yaml "gopkg.in/yaml.v3"
)

func TestLoadConfig(t *testing.T) {
	cfgData := &config.Config{
		NotificationURL: "http://example.com",
		ListenAddress:   ":8080",
		AuthenticatedHosts: map[string]string{
			"host1": "key1",
		},
		HostMaxSeconds: 60,
		Services: map[string]*config.Service{
			"service1": {
				Type: "tcp",
			},
		},
	}

	data, err := yaml.Marshal(cfgData)
	if err != nil {
		t.Fatalf("Failed to marshal config: %v", err)
	}

	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "config.json")
	if err := os.WriteFile(configFile, data, 0644); err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	hash := sha256.Sum256(data)
	expectedHash := hex.EncodeToString(hash[:])

	cfg, err := config.LoadConfig(configFile, expectedHash)
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	if cfg.NotificationURL != cfgData.NotificationURL {
		t.Errorf("Expected URL %s, got %s", cfgData.NotificationURL, cfg.NotificationURL)
	}

	if cfg.HostKeys["key1"] != "host1" {
		t.Errorf("Expected host1 for key1, got %s", cfg.HostKeys["key1"])
	}

	if val := cfg.ServicesStatus.Get("service1"); val.HostAddress != "" || val.ConsecutiveFails != 0 {
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
