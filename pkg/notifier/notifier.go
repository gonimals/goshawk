package notifier

type Notifier interface {
	Notify(title, body string) error
}
