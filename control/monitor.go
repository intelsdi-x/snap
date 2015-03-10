package control

import "time"

const (
	MonitorStopped monitorState = iota - 1 // default is stopped
	MonitorStarted

	DefaultMonitorDuration = time.Second * 60
)

type monitorState int

type monitor struct {
	State monitorState

	duration time.Duration
	quit     chan struct{}
}

type option func(m *monitor) option

// Option sets the options specified.
// Returns an option to optionally restore the last arg's previous value.
func (m *monitor) Option(opts ...option) option {
	var previous option
	for _, opt := range opts {
		previous = opt(m)
	}
	return previous
}

// MonitorDuration sets monitor's duration to v.
func MonitorDuration(v time.Duration) option {
	return func(m *monitor) option {
		previous := m.duration
		m.duration = v
		return MonitorDuration(previous)
	}
}

func newMonitor() *monitor {
	return &monitor{
		State:    MonitorStopped,
		duration: DefaultMonitorDuration,
	}
}

// start the monitor
func (m *monitor) Start(availablePlugins *availablePlugins) {
	//start a routine that will be fired every X duration looping
	//over available plugins and firing a health check routine
	ticker := time.NewTicker(m.duration)
	m.quit = make(chan struct{})
	go func() {
		for {
			select {
			case <-ticker.C:
				go func() {
					availablePlugins.Collectors.Lock()
					for availablePlugins.Collectors.Next() {
						_, apc := availablePlugins.Collectors.Item()
						for _, ap := range *apc.Plugins {
							go ap.CheckHealth()
						}
					}
					availablePlugins.Collectors.Unlock()
				}()
				go func() {
					availablePlugins.Publishers.Lock()
					for availablePlugins.Publishers.Next() {
						_, apc := availablePlugins.Publishers.Item()
						for _, ap := range *apc.Plugins {
							go ap.CheckHealth()
						}
					}
					availablePlugins.Publishers.Unlock()
				}()
				go func() {
					availablePlugins.Processors.Lock()
					for availablePlugins.Processors.Next() {
						_, apc := availablePlugins.Processors.Item()
						for _, ap := range *apc.Plugins {
							go ap.CheckHealth()
						}
					}
					availablePlugins.Processors.Unlock()
				}()
			case <-m.quit:
				ticker.Stop()
				m.State = MonitorStopped
				return
			}
		}
	}()
	m.State = MonitorStarted
}

// stop the monitor
func (m *monitor) Stop() {
	close(m.quit)
	// m.Stop()
	m.State = MonitorStopped
}
