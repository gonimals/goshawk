package notifier

import "github.com/gonimals/goshawk/pkg/config"

type Notifier interface {
	Notify(data config.AssetStatus) error
}
