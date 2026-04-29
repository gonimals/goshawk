package checker

import (
	"sync"

	"github.com/gonimals/goshawk/pkg/config"
	"github.com/gonimals/goshawk/pkg/notifier"
)

type baseDaemon struct {
	wg           *sync.WaitGroup
	err          error
	shutdownChan chan bool
	config       *config.Config
	notifier     notifier.Notifier
}

func (d *baseDaemon) Stop() error {
	d.shutdownChan <- true
	d.wg.Wait()
	return d.err
}
