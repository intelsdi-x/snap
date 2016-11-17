// +build legacy

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

import (
	"testing"
	"time"

	"github.com/intelsdi-x/gomit"

	"github.com/intelsdi-x/snap/control/plugin"

	. "github.com/smartystreets/goconvey/convey"
)

func TestMonitor(t *testing.T) {
	Convey("monitor", t, func() {
		aps := newAvailablePlugins()

		ap1 := &availablePlugin{
			pluginType: plugin.CollectorPluginType,
			version:    1,
			name:       "test",
			client:     new(MockUnhealthyPluginCollectorClient),
			healthChan: make(chan error, 1),
			emitter:    gomit.NewEventController(),
		}
		aps.insert(ap1)

		ap2 := &availablePlugin{
			pluginType: plugin.PublisherPluginType,
			version:    1,
			name:       "test",
			client:     new(MockUnhealthyPluginCollectorClient),
			healthChan: make(chan error, 1),
			emitter:    gomit.NewEventController(),
		}
		aps.insert(ap2)

		ap3 := &availablePlugin{
			pluginType: plugin.ProcessorPluginType,
			version:    1,
			name:       "test",
			client:     new(MockUnhealthyPluginCollectorClient),
			healthChan: make(chan error, 1),
			emitter:    gomit.NewEventController(),
		}
		aps.insert(ap3)

		Convey("newMonitor", func() {
			m := newMonitor(MonitorDurationOption(time.Millisecond * 123))
			So(m, ShouldHaveSameTypeAs, &monitor{})
			So(m.duration, ShouldResemble, time.Millisecond*123)
		})
		Convey("start", func() {
			m := newMonitor()
			m.Option(MonitorDurationOption(time.Millisecond * 200))
			So(m.duration, ShouldResemble, time.Millisecond*200)
			m.Start(aps)

			So(m.State, ShouldEqual, MonitorStarted)
			time.Sleep(1 * time.Second)
			Convey("health monitor", func() {
				for _, ap := range aps.all() {
					So(ap, ShouldNotBeNil)
					So(ap.(*availablePlugin).failedHealthChecks, ShouldBeGreaterThan, 3)
				}
			})
		})
		Convey("stop", func() {
			m := newMonitor()
			m.Start(aps)
			So(m.State, ShouldEqual, MonitorStarted)
			m.Stop()
			So(m.State, ShouldEqual, MonitorStopped)
		})
		Convey("override MonitorDuration", func() {
			m := newMonitor()
			oldOpt := m.Option(MonitorDurationOption(time.Millisecond * 200))
			So(m.duration, ShouldResemble, time.Millisecond*200)
			m.Option(oldOpt)
			So(m.duration, ShouldResemble, time.Second*5)
		})
	})
}
