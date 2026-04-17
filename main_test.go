package main

import (
	"log/slog"
	"testing"
	"time"
)

func TestOffline(t *testing.T) {
	config, err := loadConfig("test_files/offline_test.json", "c4b9c115e15ed60fd9578462914c352e4c4ade5017394e3c316d1b2f9be07cfd")
	if err != nil {
		slog.Error("error loading config", "error", err)
		t.FailNow()
	}
	runtimeConfig = config
	globalWaitGroup.Add(2)
	go activeCheckerRoutine()
	go passiveCheckerRoutine()

	time.Sleep(10 * time.Second)
	slog.Info("requesting shutdown")
	gracefulShutdown = true
	globalWaitGroup.Wait()
	slog.Info("successful exit")
}
