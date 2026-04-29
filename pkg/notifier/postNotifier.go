package notifier

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gonimals/goshawk/pkg/config"
)

type PostNotifier struct {
	notificationUrl string
	templateNotifier
}

func NewPostNotifier(cfg *config.Config) Notifier {
	return &PostNotifier{
		notificationUrl: cfg.NotificationURL,
		templateNotifier: templateNotifier{
			cfg: cfg,
		},
	}
}

func (pn *PostNotifier) Notify(data config.AssetStatus) error {
	title, body := pn.parseMessages(data)
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
