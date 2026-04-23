package notifier

import (
	"fmt"
	"log/slog"
)

type TestNotifier struct {
	NotificationLog []string
}

func NewTestNotifier() Notifier {
	return &TestNotifier{}
}

func (tn *TestNotifier) Notify(title, body string) error {
	slog.Debug("New Notification", "title", title, "body", body)
	tn.NotificationLog = append(tn.NotificationLog,
		fmt.Sprintf("\"title\": \"%s\", \"body\": \"%s\"", title, body))
	return nil
}
