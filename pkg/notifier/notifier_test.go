package notifier_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gonimals/goshawk/pkg/config"
	"github.com/gonimals/goshawk/pkg/notifier"
)

var testingLogCfg = func() *config.Config {
	out, err := config.ParseConfigBytes([]byte(`{
	"services": { "dummy": { "type": "tcp" } },
	"template_title": "Service {{ .ServiceName }} is {{ if .IsActive }}up{{ else }}down{{ end }}",
	"template_body": "The service ({{ .ServiceName }}) has been switched to {{ if .IsActive }}up{{ else }}down{{ end }} on {{ .LastChange.Format \"15:04:05\" }}{{ if not .IsActive }}\nThe service has been down after {{ .ConsecutiveFails }} consecutive failures with initial reason: {{ .DownReason }}{{ end }}"
	}`))
	if err != nil {
		panic(err)
	}
	return out
}()

var testingPostCfg = func() *config.Config {
	out, err := config.ParseConfigBytes([]byte(`{
	"services": { "dummy": { "type": "tcp" } },
	"template_body": "{ \"chat_id\": 120963222, \"text\": \"Service {{ .ServiceName }} is {{ if .IsActive }}up{{ else }}down{{ end }}\" }"
	}`))
	if err != nil {
		panic(err)
	}
	return out
}()

var testingAssetStatus = config.AssetStatus{
	ServiceName: "test",
	IsActive:    true,
	HostAddress: "127.0.0.1",
	LastChange:  time.Time{},
}

const expectedLog = "{\"ServiceName\":\"test\",\"ConsecutiveFails\":0,\"LastCheck\":\"0001-01-01T00:00:00Z\",\"LastChange\":\"0001-01-01T00:00:00Z\",\"IsActive\":true,\"Notified\":false,\"HostAddress\":\"127.0.0.1\",\"DownReason\":\"\"}"
const expectedTitle = "Service test is up"

func TestLogNotifier(t *testing.T) {
	n := notifier.NewNotifier(testingLogCfg)
	if err := n.Notify(testingAssetStatus); err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
}

func TestTestNotifier(t *testing.T) {
	n := notifier.NewTestNotifier()
	if err := n.Notify(testingAssetStatus); err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	tn, ok := n.(*notifier.TestNotifier)
	if !ok {
		t.Fatal("Expected TestNotifier type")
	}
	if len(tn.NotificationLog) != 1 {
		t.Fatalf("Expected 1 notification log, got %d", len(tn.NotificationLog))
	}
	if tn.NotificationLog[0] != expectedLog {
		t.Errorf("Expected log %q, got %q", expectedLog, tn.NotificationLog[0])
	}
}

func TestPostNotifier(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST method, got %s", r.Method)
		}
		var payload map[string]any
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Errorf("Failed to decode payload: %v", err)
		}
		if payload["text"] != expectedTitle {
			t.Errorf("Unexpected payload: %v", payload)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	testingPostCfg.NotificationURL = server.URL

	n := notifier.NewNotifier(testingPostCfg)
	if err := n.Notify(testingAssetStatus); err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
}

func TestPostNotifierError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	testingLogCfg.NotificationURL = server.URL

	n := notifier.NewNotifier(testingLogCfg)
	if err := n.Notify(testingAssetStatus); err == nil {
		t.Errorf("Expected error for non-200 status, got none")
	}
}

func TestPostNotifierTelegram(t *testing.T) {
	t.SkipNow() //comment this line to run this test
	testingPostCfg.NotificationURL = "https://api.telegram.org/bot8228951974:AAHRVMEF644x0EjACErU9Kd9Nr7bSzQV7cU/sendMessage"

	n := notifier.NewNotifier(testingPostCfg)
	if err := n.Notify(testingAssetStatus); err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
}
