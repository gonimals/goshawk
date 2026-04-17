package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"time"
)

func passiveCheckerRoutine() error {
	defer globalWaitGroup.Done()
	mux := http.NewServeMux()
	mux.HandleFunc("/", httpPassiveEndpoint)
	srv := &http.Server{
		Addr:    runtimeConfig.ListenAddress,
		Handler: mux,
	}
	go func() {
		ticker := time.NewTicker(time.Second)
		for range ticker.C {
			if gracefulShutdown {
				break
			}
		}
		ticker.Stop()
		ctx, cancel := context.WithTimeout(context.Background(), 0)
		srv.Shutdown(ctx)
		cancel()
	}()
	err := srv.ListenAndServe()
	if err == http.ErrServerClosed {
		err = nil
	}
	return err
}

func httpPassiveEndpoint(w http.ResponseWriter, r *http.Request) {
	if runtimeConfig == nil {
		fmt.Fprint(w, "no config mode")
		slog.Debug("http check without config")
		return
	}
	providedKey := r.URL.Query().Get("key")
	if providedKey == "" {
		fmt.Fprint(w, "check ok")
		slog.Debug("http plain check")
		return
	}
	authHost := runtimeConfig.HostKeys[providedKey]
	if authHost == "" {
		fmt.Fprintf(w, "auth failed")
		slog.Debug("http auth failed", "host", r.RemoteAddr)
		return
	}
	slog.Info("host online request", "host", authHost)
	if r.RemoteAddr == authHost {
		fmt.Fprint(w, "auth check ok")
		slog.Debug("http auth success", "host", authHost)
		return
	}
	fmt.Fprintf(w, "auth failed")
	slog.Warn("http auth for differnt host", "remoteAddr", r.RemoteAddr, "databaseAddr", authHost)
}
