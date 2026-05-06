package checker

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"
	"time"

	"github.com/gonimals/goshawk/pkg/config"
	"github.com/gonimals/goshawk/pkg/notifier"
)

const activeTestConfig = `
services:
  test_service:
    type: bash_script
    max_fails: 1
    bash_script:
      code: echo 'ok'
      expected_output_regexp: 'ok'
  failing_service:
    type: bash_script
    max_fails: 1
    bash_script:
      code: exit 1
      expected_output_regexp: '.*'
`

func TestActiveChecker(t *testing.T) {

	cfg, err := config.ParseConfigBytes([]byte(activeTestConfig))
	if err != nil {
		t.Fatalf("error parsing test config: %v", err)
	}

	notif := notifier.NewTestNotifier()

	ac := NewActiveChecker(cfg, notif)

	// Wait a bit for the checker to run at least one tick and check services
	time.Sleep(1500 * time.Millisecond)

	err = ac.Stop()
	if err != nil {
		t.Fatalf("unexpected error on stop: %v", err)
	}

	if !cfg.Services["test_service"].Status.IsActive {
		t.Errorf("expected test_service to be active")
	}

	if cfg.Services["failing_service"].Status.IsActive {
		t.Errorf("expected failing_service to be inactive")
	}

	testNotif, _ := notif.(*notifier.TestNotifier)
	if len(testNotif.GetLogs()) == 0 {
		t.Errorf("expected notifications to be sent")
	}

	testCheckTCP(t, ac)
	testCheckWebRequest(t, ac)
	testCheckBashScript(t, ac)
}

func launchTestingWebServer() (server *httptest.Server, address string) {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("ok"))
	})
	server = httptest.NewServer(mux)
	return server, server.Listener.Addr().String()
}

func testCheckTCP(t *testing.T, ac *ActiveChecker) {
	server, address := launchTestingWebServer()
	if server == nil {
		t.Fatal("failed to start server")
	}
	defer server.Close()

	err := ac.CheckTCP(&config.TCPAction{
		Address: address,
	})
	if err != nil {
		t.Fatalf("error checking TCP: %v", err)
	}
}

func testCheckWebRequest(t *testing.T, ac *ActiveChecker) {
	server, address := launchTestingWebServer()
	if server == nil {
		t.Fatal("failed to start server")
	}
	defer server.Close()

	err := ac.CheckWebRequest(&config.WebRequestAction{
		URL:            fmt.Sprintf("http://%s/", address),
		Method:         "GET",
		ExpectedStatus: 200,
	})
	if err != nil {
		t.Fatalf("error checking web request: %v", err)
	}

}

func testCheckBashScript(t *testing.T, ac *ActiveChecker) {
	err := ac.CheckBashScript(&config.BashScriptAction{
		Code:                 "echo 'hello world'",
		ExpectedOutputRegexp: regexp.MustCompile("^hello world\n$"),
	})
	if err != nil {
		t.Fatalf("error checking bash script: %v", err)
	}
	err = ac.CheckBashScript(&config.BashScriptAction{
		Code:                 "echo 'hello world'",
		ExpectedOutputRegexp: regexp.MustCompile("^hello [a-z]{4}"),
	})
	if err != nil {
		t.Fatalf("error checking bash script: %v", err)
	}

	err = ac.CheckBashScript(&config.BashScriptAction{
		Code:                 "echo 'hello world'",
		ExpectedOutputRegexp: regexp.MustCompile("^goodbye\n$"),
	})
	if err == nil {
		t.Fatalf("expected error checking bash script with wrong output")
	}

	err = ac.CheckBashScript(&config.BashScriptAction{
		Code:                 "exit 1",
		ExpectedOutputRegexp: regexp.MustCompile(".*"),
	})
	if err == nil {
		t.Fatalf("expected error checking bash script that fails")
	}
}
