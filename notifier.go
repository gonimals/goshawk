package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
)

const notificationTemplateUp = "The service %s is working"
const notificationTemplateDown = "The service %s is down"

func Notify(serviceName string) {
	template := notificationTemplateDown
	if serviceStatus.Get(serviceName).isActive {
		template = notificationTemplateUp
	}
	notify(runtimeConfig.NotificationURL, serviceName,
		fmt.Sprintf(template, serviceName))
}

func notify(url, serviceName, message string) {

	payload := map[string]string{
		"service": serviceName,
		"message": message,
	}

	if url == "" {
		slog.Info("[New Notification]", "data", payload)
		return
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		slog.Warn("could not marshal notification payload", "error", err)
		return
	}

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonPayload))
	if err != nil {
		slog.Warn("could not send notification", "error", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		slog.Warn("notification endpoint returned non-200", "status", resp.StatusCode)
	}
}
