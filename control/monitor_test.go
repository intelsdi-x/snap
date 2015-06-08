package control

import (
	"testing"
	"time"

	"github.com/intelsdi-x/gomit"

	"github.com/intelsdi-x/pulse/control/plugin"

	. "github.com/smartystreets/goconvey/convey"
)

type mockPluginClient struct{}

func (mp *mockPluginClient) Ping() error {
	return nil
}

func (mp *mockPluginClient) Kill(r string) error {
	return nil
}

func TestMonitor(t *testing.T) {
	Convey("monitor", t, func() {
		aps := newAvailablePlugins()

		ap1 := &availablePlugin{
			Type:       plugin.CollectorPluginType,
			version:    1,
			name:       "test",
			Client:     new(MockUnhealthyPluginCollectorClient),
			healthChan: make(chan error, 1),
			emitter:    gomit.NewEventController(),
		}
		ap1.makeKey()
		aps.Insert(ap1)

		ap2 := &availablePlugin{
			Type:    plugin.PublisherPluginType,
			version: 1,
			name:    "test",
			Client:  &mockPluginClient{},
			emitter: &gomit.EventController{},
		}
		ap2.makeKey()
		aps.Insert(ap2)

		ap3 := &availablePlugin{
			Type:    plugin.ProcessorPluginType,
			version: 1,
			name:    "test",
			Client:  &mockPluginClient{},
			emitter: &gomit.EventController{},
		}
		ap3.makeKey()
		aps.Insert(ap3)

		Convey("newMonitor", func() {
			m := newMonitor(MonitorDurationOption(time.Millisecond * 123))
			So(m, ShouldHaveSameTypeAs, &monitor{})
			So(m.duration, ShouldResemble, time.Millisecond*123)
		})
		Convey("start", func() {
			m := newMonitor()
			m.Option(MonitorDurationOption(time.Millisecond * 200))
			m.Start(aps)

			So(m.State, ShouldEqual, MonitorStarted)
			time.Sleep(1 * time.Second)
			Convey("health monitor", func() {
				for aps.Collectors.Next() {
					_, item := aps.Collectors.Item()
					So(item, ShouldNotBeNil)
					So((*(*item).Plugins)[0].failedHealthChecks, ShouldBeGreaterThan, 3)
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
