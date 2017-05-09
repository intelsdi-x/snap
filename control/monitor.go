/*
http://www.apache.org/licenses/LICENSE-2.0.txt


Copyright 2015 Intel Corporation

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package control

import "time"

const (
	// MonitorStopped - enum representation of monitor stopped state
	MonitorStopped monitorState = iota - 1 // default is stopped
	// MonitorStarted - enum representation of monitor started state
	MonitorStarted

	// DefaultMonitorDuration - the default monitor duration.
	DefaultMonitorDuration = time.Second * 5
)

type monitorState int

type monitor struct {
	State monitorState

	duration time.Duration
	quit     chan struct{}
}

type monitorOption func(m *monitor) monitorOption

// Option sets the options specified.
// Returns an option to optionally restore the last arg's previous value.
func (m *monitor) Option(opts ...monitorOption) monitorOption {
	var previous monitorOption
	for _, opt := range opts {
		previous = opt(m)
	}
	return previous
}

// MonitorDurationOption sets monitor's duration to v.
func MonitorDurationOption(v time.Duration) monitorOption {
	return func(m *monitor) monitorOption {
		previous := m.duration
		m.duration = v
		return MonitorDurationOption(previous)
	}
}

func newMonitor(opts ...monitorOption) *monitor {
	mon := &monitor{
		State:    MonitorStopped,
		duration: DefaultMonitorDuration,
	}
	//set options
	for _, opt := range opts {
		opt(mon)
	}
	return mon
}

// Start starts the monitor
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
					availablePlugins.RLock()
					for _, ap := range availablePlugins.all() {
						if !ap.IsRemote() {
							go ap.CheckHealth()
						}
					}
					availablePlugins.RUnlock()
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

// Stop stops the monitor
func (m *monitor) Stop() {
	close(m.quit)
	m.State = MonitorStopped
}
