package checker

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gonimals/goshawk/pkg/config"
	"github.com/gonimals/goshawk/pkg/notifier"
	"github.com/gonimals/goshawk/pkg/util"
)

func TestPassiveChecker(t *testing.T) {
	hostsStatus := util.NewSyncMap[string, config.AssetStatus]()
	hostsStatus.Set("host1", config.AssetStatus{
		IsActive:  true,
		LastCheck: time.Now().Add(-10 * time.Second),
	})

	cfg := &config.Config{
		ListenAddress:  "127.0.0.1:0", // Use dynamic port to prevent binding collisions
		HostMaxSeconds: 2,
		AuthenticatedHosts: map[string]string{
			"host1": "key1",
		},
		HostKeys: map[string]string{
			"key1": "host1",
		},
		HostsStatus: hostsStatus,
	}

	notif := notifier.NewTestNotifier()
	pc := NewPassiveChecker(cfg, notif)

	// Wait for server to start
	time.Sleep(100 * time.Millisecond)

	// Test unauthenticated request
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()
	pc.httpPassiveEndpoint(w, req)

	resp := w.Result()
	body, _ := io.ReadAll(resp.Body)
	if string(body) != "check ok" {
		t.Errorf("expected 'check ok', got %s", string(body))
	}

	// Test failed auth
	req = httptest.NewRequest(http.MethodGet, "/?key=wrong", nil)
	w = httptest.NewRecorder()
	pc.httpPassiveEndpoint(w, req)

	resp = w.Result()
	body, _ = io.ReadAll(resp.Body)
	if string(body) != "auth failed" {
		t.Errorf("expected 'auth failed', got %s", string(body))
	}

	// Test successful auth
	req = httptest.NewRequest(http.MethodGet, "/?key=key1", nil)
	w = httptest.NewRecorder()
	pc.httpPassiveEndpoint(w, req)

	resp = w.Result()
	body, _ = io.ReadAll(resp.Body)
	if string(body) != "auth check ok" {
		t.Errorf("expected 'auth check ok', got %s", string(body))
	}

	// Wait for a tick to let it check offline hosts
	time.Sleep(1500 * time.Millisecond)

	err := pc.Stop()
	if err != nil {
		t.Fatalf("unexpected error on stop: %v", err)
	}

	status := hostsStatus.Get("host1")
	if !status.IsActive {
		t.Errorf("expected host1 to be active")
	}
}

func TestPassiveCheckerTimeout(t *testing.T) {
	hostsStatus := util.NewSyncMap[string, config.AssetStatus]()
	hostsStatus.Set("host1", config.AssetStatus{
		IsActive:  true,
		LastCheck: time.Now().Add(-10 * time.Second),
	})

	cfg := &config.Config{
		ListenAddress:  "127.0.0.1:0",
		HostMaxSeconds: 1, // Will timeout immediately
		AuthenticatedHosts: map[string]string{
			"host1": "key1",
		},
		HostKeys: map[string]string{
			"key1": "host1",
		},
		HostsStatus: hostsStatus,
	}

	notif := notifier.NewTestNotifier()
	pc := NewPassiveChecker(cfg, notif)

	// Let the ticker run and detect the timeout
	time.Sleep(1500 * time.Millisecond)

	pc.Stop()

	status := hostsStatus.Get("host1")
	if status.IsActive {
		t.Errorf("expected host1 to be inactive due to timeout")
	}
}
