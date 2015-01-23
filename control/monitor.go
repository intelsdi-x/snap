package control

import (
	"time"
)

const (
	StoppedState monitorState = "stopped"
	StartedState monitorState = "started"
	RunningState monitorState = "running"

	interval int = 60
)

type monitorState string

type monitor struct {
	State monitorState
	//Unix  time since last run
	LastRun time.Time
	quit    chan struct{}
}

func newMonitor() *monitor {
	m := new(monitor)
	m.State = StoppedState
	return m
}

func (m *monitor) Start() {
	//start a routine that will be fired every X duration to loop
	//over available plugins and fire a health check

	ticker := time.NewTicker(time.Second * time.Duration(interval))
	m.quit = make(chan struct{})
	go func() {
		for {
			select {
			case <-ticker.C:
				for _, ap := range availablePlugins {
					if ap.State == PluginRunning {
						go ap.checkHealth()
					}
				}
			case <-m.quit:
				ticker.Stop()
				return
			}
		}
	}()

	m.State = StartedState
}

func (m *monitor) Stop() {
	close(m.quit)
	m.State = StoppedState
}
