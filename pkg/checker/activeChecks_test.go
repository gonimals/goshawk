package checker_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gonimals/goshawk/pkg/checker"
	"github.com/gonimals/goshawk/pkg/config"
)

func launchTestingWebServer() (server *httptest.Server, address string) {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("ok"))
	})
	server = httptest.NewServer(mux)
	return server, server.Listener.Addr().String()
}

func TestCheckTCP(t *testing.T) {
	server, address := launchTestingWebServer()
	if server == nil {
		t.Fatal("failed to start server")
	}
	defer server.Close()

	err := checker.CheckTCP(&config.TCPAction{
		Address: address,
	})
	if err != nil {
		t.Fatalf("error checking TCP: %v", err)
	}
}

func TestCheckWebRequest(t *testing.T) {
	server, address := launchTestingWebServer()
	if server == nil {
		t.Fatal("failed to start server")
	}
	defer server.Close()

	err := checker.CheckWebRequest(&config.WebRequestAction{
		URL:            fmt.Sprintf("http://%s/", address),
		Method:         "GET",
		ExpectedStatus: 200,
	})
	if err != nil {
		t.Fatalf("error checking web request: %v", err)
	}

}

func TestCheckBashScript(t *testing.T) {
	err := checker.CheckBashScript(&config.BashScriptAction{
		Code:                 "echo 'hello world'",
		ExpectedOutputRegexp: "^hello world\n$",
	})
	if err != nil {
		t.Fatalf("error checking bash script: %v", err)
	}
	err = checker.CheckBashScript(&config.BashScriptAction{
		Code:                 "echo 'hello world'",
		ExpectedOutputRegexp: "^hello [a-z]{4}",
	})
	if err != nil {
		t.Fatalf("error checking bash script: %v", err)
	}

	err = checker.CheckBashScript(&config.BashScriptAction{
		Code:                 "echo 'hello world'",
		ExpectedOutputRegexp: "^goodbye\n$",
	})
	if err == nil {
		t.Fatalf("expected error checking bash script with wrong output")
	}

	err = checker.CheckBashScript(&config.BashScriptAction{
		Code:                 "exit 1",
		ExpectedOutputRegexp: ".*",
	})
	if err == nil {
		t.Fatalf("expected error checking bash script that fails")
	}
}
