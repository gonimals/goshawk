package main

import (
	"fmt"
	"io"
	"log/slog"
	"net/http"
)

func main() {
	http.HandleFunc("/notify", func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			slog.Warn("error reading body", "error", err)
			http.Error(w, "Error reading request body", http.StatusInternalServerError)
			return
		}
		slog.Info("notification received", "body", body)
		fmt.Fprint(w, "notification received")
	})

	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "ok")
	})

	slog.Error("web server exited", "error", http.ListenAndServe(":8081", nil))
}
