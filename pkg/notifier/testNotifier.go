package notifier

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"sync"

	"github.com/gonimals/goshawk/pkg/config"
)

type TestNotifier struct {
	mu              sync.Mutex
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
	tn.mu.Lock()
	defer tn.mu.Unlock()
	tn.NotificationLog = append(tn.NotificationLog, string(dataJson))
	return nil
}

func (tn *TestNotifier) GetLogs() []string {
	tn.mu.Lock()
	defer tn.mu.Unlock()
	logs := make([]string, len(tn.NotificationLog))
	copy(logs, tn.NotificationLog)
	return logs
}
