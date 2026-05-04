package checker_test

import (
	"testing"
	"time"

	"github.com/gonimals/goshawk/pkg/checker"
	"github.com/gonimals/goshawk/pkg/config"
	"github.com/gonimals/goshawk/pkg/notifier"
	"github.com/gonimals/goshawk/pkg/util"
)

func TestActiveChecker(t *testing.T) {
	servicesStatus := util.NewSyncMap[string, config.AssetStatus]()
	servicesStatus.Set("test_service", config.AssetStatus{})
	servicesStatus.Set("failing_service", config.AssetStatus{})

	cfg := &config.Config{
		Services: map[string]config.Service{
			"test_service": {
				Type: "bash_script",
				BashScript: &config.BashScriptAction{
					Code:                 "echo 'ok'",
					ExpectedOutputRegexp: "ok",
				},
				MaxFails: 2,
			},
			"failing_service": {
				Type: "bash_script",
				BashScript: &config.BashScriptAction{
					Code:                 "exit 1",
					ExpectedOutputRegexp: ".*",
				},
				MaxFails: 1,
			},
		},
		ServicesStatus: servicesStatus,
	}

	notif := notifier.NewTestNotifier()

	ac := checker.NewActiveChecker(cfg, notif)

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
}
