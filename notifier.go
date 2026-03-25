package main

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
)

func notify(url, serviceName, errorMsg string) {
	if url == "" {
		return
	}

	payload := map[string]string{
		"service": serviceName,
		"error":   errorMsg,
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		log.Printf("could not marshal notification payload: %v", err)
		return
	}

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonPayload))
	if err != nil {
		log.Printf("could not send notification: %v", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("notification endpoint returned non-200 status code: %d", resp.StatusCode)
	}
}
