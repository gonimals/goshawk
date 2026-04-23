package notifier

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

type PostNotifier struct {
	notificationUrl string
}

func NewPostNotifier(notificationUrl string) Notifier {
	return &PostNotifier{
		notificationUrl: notificationUrl,
	}
}

func (pn *PostNotifier) Notify(title, body string) error {
	payload := map[string]string{
		"title": title,
		"body":  body,
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("could not marshal notification payload: %v", err)
	}

	resp, err := http.Post(pn.notificationUrl, "application/json", bytes.NewBuffer(jsonPayload))
	if err != nil {
		return fmt.Errorf("could not send notification: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("notification endpoint returned non-200 status: %v", resp.StatusCode)
	}
	return nil
}
