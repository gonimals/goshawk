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

var SignalChan = make(chan os.Signal, 1)

func main() {
	os.Exit(run(os.Args[1:]))
}

func run(args []string) int {
	fs := flag.NewFlagSet("goshawk", flag.ContinueOnError)
	configFile := fs.String("config", "config.json", "Configuration file")
	configHash := fs.String("hash", "", "Expected SHA256 hash of the configuration file")
	verboseMode := fs.Bool("v", false, "Enable verbose logging")

	if err := fs.Parse(args); err != nil {
		return 1
	}

	if *verboseMode {
		slog.SetLogLoggerLevel(slog.LevelDebug)
	}

	config, err := config.LoadConfig(*configFile, *configHash)
	if err != nil {
		slog.Error("Error loading configuration", "error", err)
		return 1
	}

	var sender notifier.Notifier
	sender = notifier.NewNotifier(config)

	activeChecker := checker.NewActiveChecker(config, sender)
	passiveChecker := checker.NewPassiveChecker(config, sender)

	signal.Notify(SignalChan, os.Interrupt)

	sig := <-SignalChan
	if sig != os.Interrupt {
		slog.Warn("unexpected signal in handler", "signal", sig)
		return 1
	}
	slog.Info("received interrupt", "signal", sig)
	returnCode := 0
	err = activeChecker.Stop()
	if err != nil {
		slog.Warn("error stopping active checker", "error", err)
		returnCode = 1
	} else {
		slog.Info("active checker stopped")
	}
	err = passiveChecker.Stop()
	if err != nil {
		slog.Warn("error stopping passive checker", "error", err)
		returnCode = 1
	} else {
		slog.Info("passive checker stopped")
	}
	slog.Info("exiting")
	return returnCode
}
