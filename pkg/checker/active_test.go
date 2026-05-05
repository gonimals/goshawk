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
	"github.com/gonimals/goshawk/pkg/util"
)

func TestActiveChecker(t *testing.T) {
	servicesStatus := util.NewSyncMap[string, config.AssetStatus]()
	servicesStatus.Set("test_service", config.AssetStatus{})
	servicesStatus.Set("failing_service", config.AssetStatus{})

	cfg := &config.Config{
		Services: map[string]*config.Service{
			"test_service": {
				Type: "bash_script",
				BashScript: &config.BashScriptAction{
					Code:                 "echo 'ok'",
					ExpectedOutputRegexp: regexp.MustCompile("ok"),
				},
				MaxFails: 2,
			},
			"failing_service": {
				Type: "bash_script",
				BashScript: &config.BashScriptAction{
					Code:                 "exit 1",
					ExpectedOutputRegexp: regexp.MustCompile(".*"),
				},
				MaxFails: 1,
			},
		},
		ServicesStatus: servicesStatus,
	}

	notif := notifier.NewTestNotifier()

	ac := NewActiveChecker(cfg, notif)

	// Wait a bit for the checker to run at least one tick and check services
	time.Sleep(1500 * time.Millisecond)

	err := ac.Stop()
	if err != nil {
		t.Fatalf("unexpected error on stop: %v", err)
	}

	status := servicesStatus.Get("test_service")
	if !status.IsActive {
		t.Errorf("expected test_service to be active")
	}

	statusFail := servicesStatus.Get("failing_service")
	if statusFail.IsActive {
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
