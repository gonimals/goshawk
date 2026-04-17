package main

import (
	"sync"
	"time"
)

var (
	serviceStatus    = NewSafeMap[string, *ServiceStatus]()
	statusMutex      = &sync.Mutex{}
	runtimeConfig    *Config
	gracefulShutdown bool
	globalWaitGroup  sync.WaitGroup
)

type Config struct {
	ListenAddress   string             `json:"listen_address"`
	Services        map[string]Service `json:"services"`
	NotificationURL string             `json:"notification_url"`
	HostKeys        map[string]string  `json:"host_keys"`
}

// ServiceStatus holds the state of a service
type ServiceStatus struct {
	consecutiveFails int
	lastCheck        time.Time
	lastChange       time.Time
	isActive         bool
	notified         bool
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
	Host string `json:"host"`
	Port int    `json:"port"`
}

type WebRequestAction struct {
	URL            string `json:"url"`
	Method         string `json:"method"`
	Body           string `json:"body,omitempty"`
	ExpectedStatus int    `json:"expected_status"`
}

type BashScriptAction struct {
	Code                 string `json:"code"`
	ExpectedOutputRegexp string `json:"expected_output_regexp"`
}
