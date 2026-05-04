package notifier

import (
	"net/http"
	"time"

	"github.com/gonimals/goshawk/pkg/config"
	"golang.org/x/time/rate"
)

type Notifier interface {
	// Notify sends the notification about the asset status
	// It is better to launch it in a dedicated go routine
	Notify(data config.AssetStatus) error
}

func NewNotifier(cfg *config.Config) Notifier {
	handler := templateHandler{
		cfg: cfg,
	}

	if cfg.NotificationURL == "" {
		return &LogNotifier{
			templateHandler: handler,
		}
	}

	var limiter *rate.Limiter
	if cfg.NotificationRateLimit > 0 {
		limiter = rate.NewLimiter(rate.Limit(cfg.NotificationRateLimit), 1)
	}

	return &PostNotifier{
		templateHandler: handler,
		limiter:         limiter,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}
