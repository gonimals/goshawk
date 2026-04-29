package notifier

import (
	"strings"

	"github.com/gonimals/goshawk/pkg/config"
)

type templateNotifier struct {
	cfg *config.Config
}

func (tn *templateNotifier) parseMessages(data config.AssetStatus) (string, string) {
	titleSB := strings.Builder{}
	bodySB := strings.Builder{}
	tn.cfg.TemplateTitleParsed.Execute(&titleSB, data)
	tn.cfg.TemplateBodyParsed.Execute(&bodySB, data)
	return titleSB.String(), bodySB.String()
}
