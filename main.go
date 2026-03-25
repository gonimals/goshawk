package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

func main() {
	configFile := flag.String("config", "config.json", "Configuration file")
	configHash := flag.String("hash", "", "SHA256 hash of the configuration file")
	passphrase := flag.String("passphrase", "", "Passphrase to confirm who is online")
	flag.Parse()

	config, err := loadConfig(*configFile, *configHash)
	if err != nil {
		log.Printf("Error loading configuration: %v", err)
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if *passphrase != "" {
			if r.URL.Query().Get("passphrase") != *passphrase {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}
		}
		fmt.Fprint(w, "check ok")
	})

	go func() {
		log.Fatal(http.ListenAndServe(":8080", nil))
	}()

	if config != nil {
		ticker := time.NewTicker(60 * time.Second)
		defer ticker.Stop()

		for range ticker.C {
			for _, service := range config.Services {
				go checkService(service, config.NotificationURL)
			}
		}
	}

	select {}
}

func loadConfig(configFile, expectedHash string) (*Config, error) {
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

	if expectedHash != "" && actualHash != expectedHash {
		fmt.Printf("Calculated SHA256: %s\n", actualHash)
		return nil, fmt.Errorf("configuration hash mismatch")
	}

	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("could not parse config file: %w", err)
	}

	return &config, nil
}


