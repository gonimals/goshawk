package main

import (
	"flag"
	"log/slog"
	"os"
)

func main() {
	configFile := flag.String("config", "config.json", "Configuration file")
	configHash := flag.String("hash", "", "SHA256 hash of the configuration file")
	flag.Parse()

	config, err := loadConfig(*configFile, *configHash)
	if err != nil {
		slog.Error("Error loading configuration", "error", err)
		os.Exit(1)
	}
	runtimeConfig = config
	go activeCheckerRoutine()
	go passiveCheckerRoutine()

	select {}
}
