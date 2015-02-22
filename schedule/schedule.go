package schedule

type Schedule interface {
	Wait() chan struct{}
}
