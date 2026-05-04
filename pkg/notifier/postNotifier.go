package notifier

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"

	"github.com/gonimals/goshawk/pkg/config"
	"golang.org/x/time/rate"
)

type PostNotifier struct {
	templateHandler
	limiter    *rate.Limiter
	httpClient *http.Client
}

func (pn *PostNotifier) Notify(data config.AssetStatus) error {
	if pn.limiter != nil {
		if err := pn.limiter.Wait(context.Background()); err != nil {
			slog.Warn("rate limiter error", "error", err)
			return fmt.Errorf("rate limiter error: %v", err)
		}
	}

	_, body := pn.parseMessages(data)

	resp, err := pn.httpClient.Post(pn.cfg.NotificationURL, "application/json", strings.NewReader(body))
	if err != nil {
		slog.Warn("could not send notification", "error", err)
		return fmt.Errorf("could not send notification: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, err := io.ReadAll(resp.Body)
		if err != nil {
			respBody = []byte(fmt.Sprintf("could not read response body: %v", err))
		}
		slog.Warn("notification endpoint returned non-200 status", "status", resp.StatusCode, "body", string(respBody))
		return fmt.Errorf("notification endpoint returned non-200 status: %v", resp.StatusCode)
	}
	return nil
}
