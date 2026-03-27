package main

type Config struct {
	Services        []Service         `json:"services"`
	NotificationURL string            `json:"notification_url"`
	HostKeys        map[string]string `json:"host_keys"`
}

type Service struct {
	Name   string `json:"name"`
	Action Action `json:"action"`
}

type Action struct {
	Type       string            `json:"type"`
	TCP        *TCPAction        `json:"tcp,omitempty"`
	WebRequest *WebRequestAction `json:"web_request,omitempty"`
	BashScript *BashScriptAction `json:"bash_script,omitempty"`
	Frequency  int               `json:"frequency"`
	MaxFails   int               `json:"max_fails"`
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
