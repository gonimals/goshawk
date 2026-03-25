package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
)

func main() {
	http.HandleFunc("/notify", func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			log.Printf("Error reading body: %v", err)
			http.Error(w, "Error reading request body", http.StatusInternalServerError)
			return
		}
		log.Printf("Notification received: %s", body)
		fmt.Fprint(w, "notification received")
	})

	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "ok")
	})

	log.Fatal(http.ListenAndServe(":8081", nil))
}
