package notifier

import (
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/gonimals/goshawk/pkg/config"
)

type TestNotifier struct {
	NotificationLog []string
}

func NewTestNotifier() Notifier {
	return &TestNotifier{}
}

func (tn *TestNotifier) Notify(data config.AssetStatus) error {
	slog.Debug("New Notification", "data", data)
	dataJson, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("could not marshal notification payload: %v", err)
	}
	tn.NotificationLog = append(tn.NotificationLog, string(dataJson))
	return nil
}
