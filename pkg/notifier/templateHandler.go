package notifier

import (
	"log/slog"
	"strings"

	"github.com/gonimals/goshawk/pkg/config"
)

type templateHandler struct {
	cfg *config.Config
}

func (tn *templateHandler) parseMessages(data config.AssetStatus) (string, string) {
	titleSB := strings.Builder{}
	bodySB := strings.Builder{}
	err := tn.cfg.TemplateTitleParsed.Execute(&titleSB, data)
	if err != nil {
		slog.Warn("error executing template title", "error", err, "data", data)
		titleSB = strings.Builder{}
		titleSB.WriteString("error executing template title")
	}
	err = tn.cfg.TemplateBodyParsed.Execute(&bodySB, data)
	if err != nil {
		slog.Warn("error executing template body", "error", err, "data", data)
		bodySB = strings.Builder{}
		bodySB.WriteString("error executing template body")
	}
	return titleSB.String(), bodySB.String()
}
