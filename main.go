package main

import (
	"flag"
	"log/slog"
	"os"
	"os/signal"

	"github.com/gonimals/goshawk/pkg/checker"
	"github.com/gonimals/goshawk/pkg/config"
	"github.com/gonimals/goshawk/pkg/notifier"
)

func main() {
	configFile := flag.String("config", "config.json", "Configuration file")
	configHash := flag.String("hash", "", "Expected SHA256 hash of the configuration file")
	verboseMode := flag.Bool("v", false, "Enable verbose logging")
	flag.Parse()

	if *verboseMode {
		slog.SetLogLoggerLevel(slog.LevelDebug)
	}

	config, err := config.LoadConfig(*configFile, *configHash)
	if err != nil {
		slog.Error("Error loading configuration", "error", err)
		os.Exit(1)
	}

	var sender notifier.Notifier
	if config.NotificationURL == "" {
		sender = notifier.NewLogNotifier(config)
	} else {
		sender = notifier.NewPostNotifier(config)
	}

	activeChecker := checker.NewActiveChecker(config, sender)
	passiveChecker := checker.NewPassiveChecker(config, sender)

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt)

	sig := <-signalChan
	if sig != os.Interrupt {
		slog.Warn("unexpected signal in handler", "signal", sig)
		os.Exit(1)
	}
	slog.Info("received interrupt", "signal", sig)
	err = activeChecker.Stop()
	if err != nil {
		slog.Warn("error stopping active checker", "error", err)
	} else {
		slog.Info("active checker stopped")
	}
	err = passiveChecker.Stop()
	if err != nil {
		slog.Warn("error stopping passive checker", "error", err)
	} else {
		slog.Info("passive checker stopped")
	}
	slog.Info("exiting")
	os.Exit(0)
}
