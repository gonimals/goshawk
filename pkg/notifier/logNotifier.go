package notifier

import (
	"log/slog"

	"github.com/gonimals/goshawk/pkg/config"
)

type LogNotifier struct {
	templateHandler
}

func (ln *LogNotifier) Notify(data config.AssetStatus) error {
	title, body := ln.parseMessages(data)
	slog.Info("New Notification", "title", title, "body", body)
	return nil
}
