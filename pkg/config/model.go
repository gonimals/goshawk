package config

import (
	"html/template"
	"time"

	"github.com/gonimals/goshawk/pkg/util"
)

type Config struct {
	//NotificationURL defines the URL to send notifications to
	NotificationURL string `json:"notification_url"`
	//ListenAddress defines where the HTTP endpoint for passive checking will be listening
	ListenAddress string `json:"listen_address"`
	//AuthenticatedHosts defines all the hosts which should be sending alive messages to this checker
	AuthenticatedHosts map[string]string `json:"authenticated_hosts"`
	//HostKeys is the inverted version of AuthenticatedHosts
	HostKeys map[string]string `json:"-"`
	//HostMaxSeconds defines the maximum seconds without a message from an alive host
	HostMaxSeconds int `json:"host_max_seconds"`
	//Services defines the active checks this checker will be launching
	Services map[string]Service `json:"services"`

	//TemplateTitle accepts attributes from AssetStatus
	TemplateTitle       string             `json:"template_title"`
	TemplateTitleParsed *template.Template `json:"-"`
	//TemplateBody accepts attributes from AssetStatus
	TemplateBody       string             `json:"template_body"`
	TemplateBodyParsed *template.Template `json:"-"`

	ServicesStatus *util.SyncMap[string, AssetStatus] `json:"-"`
	HostsStatus    *util.SyncMap[string, AssetStatus] `json:"-"`
}

// AssetStatus holds the state of a service or a host
type AssetStatus struct {
	ServiceName      string
	ConsecutiveFails int
	LastCheck        time.Time
	LastChange       time.Time
	IsActive         bool
	Notified         bool
	HostAddress      string
	DownReason       string
}

type Service struct {
	Type             string            `json:"type"`
	TCP              *TCPAction        `json:"tcp,omitempty"`
	WebRequest       *WebRequestAction `json:"web_request,omitempty"`
	BashScript       *BashScriptAction `json:"bash_script,omitempty"`
	FrequencySeconds int               `json:"frequency_seconds"`
	MaxFails         int               `json:"max_fails"`
}

type TCPAction struct {
	Address        string `json:"address"`
	TimeoutSeconds int    `json:"timeout_seconds,omitempty"`
}

type WebRequestAction struct {
	URL            string `json:"url"`
	Method         string `json:"method"`
	Body           string `json:"body,omitempty"`
	ExpectedStatus int    `json:"expected_status"`
	TimeoutSeconds int    `json:"timeout_seconds,omitempty"`
}

type BashScriptAction struct {
	Code                 string `json:"code"`
	ExpectedOutputRegexp string `json:"expected_output_regexp"`
}
