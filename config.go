package main

type Config struct {
	Services        []Service `json:"services"`
	NotificationURL string    `json:"notification_url"`
}

type Service struct {
	Name   string `json:"name"`
	Action Action `json:"action"`
}

type Action struct {
	Type          string        `json:"type"`
	Ping          *PingAction          `json:"ping,omitempty"`
	WebRequest    *WebRequestAction    `json:"web_request,omitempty"`
	BashScript    *BashScriptAction    `json:"bash_script,omitempty"`
}

type PingAction struct {
	Host string `json:"host"`
}

type WebRequestAction struct {
	URL          string `json:"url"`
	Method       string `json:"method"`
	ExpectedStatus int    `json:"expected_status"`
}

type BashScriptAction struct {
	Code         string `json:"code"`
	ExpectedOutputRegexp string `json:"expected_output_regexp"`
}
