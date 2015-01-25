package control

import (
	"time"
)

const (
	DefaultMonitorDuration = time.Second * time.Duration(60)

	MonitorStopped monitorState = iota - 1 // default is stopped
	MonitorStarted
)

var interval int = 60

type monitorState int

type monitor struct {
	State monitorState
	quit  chan struct{}
}

func newMonitor() *monitor {
	m := new(monitor)
	m.State = MonitorStopped
	return m
}

func (m *monitor) Start() {
	//start a routine that will be fired every X duration looping
	//over available plugins and firing a health check routine
	ticker := time.NewTicker(DefaultMonitorDuration)
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
				m.State = MonitorStopped
				return
			}
		}
	}()
	m.State = MonitorStarted
}

func (m *monitor) Stop() {
	close(m.quit)
	// m.Stop()
	m.State = MonitorStopped
}
