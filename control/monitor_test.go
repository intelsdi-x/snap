package control

import (
	"testing"
	"time"

	"github.com/intelsdi-x/gomit"

	"github.com/intelsdi-x/pulse/control/plugin"
	"github.com/intelsdi-x/pulse/control/routing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestMonitor(t *testing.T) {
	Convey("monitor", t, func() {
		aps := newAvailablePlugins(&routing.RoundRobinStrategy{})

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
					So(ap.failedHealthChecks, ShouldBeGreaterThan, 3)
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
			So(m.duration, ShouldResemble, time.Second*1)
		})
	})
}
