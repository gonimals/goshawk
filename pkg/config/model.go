package config

import (
	"html/template"
	"regexp"
	"sync"
	"time"

	"github.com/gonimals/goshawk/pkg/util"
)

type Config struct {
	// NotificationURL defines the URL to send notifications to
	NotificationURL string `yaml:"notification_url"`

	// Maximum notifications allowed per second
	NotificationRateLimit int `yaml:"notification_rate_limit"`

	// ListenAddress defines where the HTTP endpoint for passive checking will be listening
	ListenAddress string `yaml:"listen_address"`

	// AuthenticatedHosts defines all the hosts which should be sending alive messages to this instance (format "host: key")
	AuthenticatedHosts map[string]string `yaml:"authenticated_hosts,omitempty"`

	// HostKeys is the inverted version of AuthenticatedHosts (format "key: host")
	HostKeys map[string]string `yaml:"-"`

	// HostMaxSeconds defines the maximum seconds without a message from an alive host
	HostMaxSeconds int `yaml:"host_max_seconds"`

	// Services defines the active checks this checker will be launching
	Services map[string]*Service `yaml:"services,omitempty"`

	// Default values for serivces
	DefaultServiceFrequency int `yaml:"default_service_frequency"`
	DefaultServiceMaxFails  int `yaml:"default_service_max_fails"`
	DefaultServiceTimeout   int `yaml:"default_service_timeout"`

	// TemplateTitle accepts attributes from AssetStatus
	TemplateTitle       string             `yaml:"template_title"`
	TemplateTitleParsed *template.Template `yaml:"-"`

	// TemplateBody accepts attributes from AssetStatus
	TemplateBody       string             `yaml:"template_body"`
	TemplateBodyParsed *template.Template `yaml:"-"`

	HostsStatus *util.SyncMap[string, AssetStatus] `yaml:"-"`
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
	Mutex            sync.Mutex        `yaml:"-"`
	Type             string            `yaml:"type"`
	TCP              *TCPAction        `yaml:"tcp,omitempty"`
	WebRequest       *WebRequestAction `yaml:"web_request,omitempty"`
	BashScript       *BashScriptAction `yaml:"bash_script,omitempty"`
	FrequencySeconds int               `yaml:"frequency_seconds,omitempty"`
	MaxFails         int               `yaml:"max_fails,omitempty"`
	TimeoutSeconds   int               `yaml:"timeout_seconds,omitempty"`
	Status           AssetStatus       `yaml:"-"`
}

type TCPAction struct {
	Address string `yaml:"address"`
}

type WebRequestAction struct {
	URL                  string         `yaml:"url"`
	Method               string         `yaml:"method"`
	Body                 string         `yaml:"body,omitempty"`
	ExpectedStatus       int            `yaml:"expected_status"`
	ExpectedOutput       string         `yaml:"expected_output_regexp"`
	ExpectedOutputRegexp *regexp.Regexp `yaml:"-"`
}

type BashScriptAction struct {
	Code                 string         `yaml:"code"`
	ExpectedOutput       string         `yaml:"expected_output_regexp"`
	ExpectedOutputRegexp *regexp.Regexp `yaml:"-"`
}
