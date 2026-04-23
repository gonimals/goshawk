package notifier

import "log/slog"

type LogNotifier struct {
}

func NewLogNotifier() Notifier {
	return &LogNotifier{}
}

func (n *LogNotifier) Notify(title, body string) error {
	slog.Info("New Notification", "title", title, "body", body)
	return nil
}
