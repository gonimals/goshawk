package config

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"os"

	"github.com/gonimals/goshawk/pkg/util"
)

func LoadConfig(configFile, expectedHash string) (*Config, error) {
	file, err := os.Open(configFile)
	if err != nil {
		return nil, fmt.Errorf("could not open config file: %w", err)
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("could not read config file: %w", err)
	}

	hash := sha256.Sum256(data)
	actualHash := hex.EncodeToString(hash[:])

	if len(expectedHash) == 0 {
		slog.Warn("no expected hash provided, skipping hash check")
	} else if actualHash[:len(expectedHash)] != expectedHash {
		slog.Warn("config hash mismatch", "calculatedHash", actualHash)
		return nil, fmt.Errorf("configuration hash mismatch")
	}

	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("could not parse config file: %w", err)
	}
	config.HostKeys = make(map[string]string)
	for host, key := range config.AuthenticatedHosts {
		config.HostKeys[key] = host
	}

	config.HostsStatus = util.NewSyncMap[string, AssetStatus]()
	for host := range config.AuthenticatedHosts {
		config.HostsStatus.Set(host, AssetStatus{})
	}

	config.ServicesStatus = util.NewSyncMap[string, AssetStatus]()
	for serviceName := range config.Services {
		config.ServicesStatus.Set(serviceName, AssetStatus{})
	}

	return &config, nil
}
