package control

import (
	"testing"
	"time"

	"github.com/intelsdilabs/gomit"

	"github.com/intelsdilabs/pulse/control/plugin"

	. "github.com/smartystreets/goconvey/convey"
)

type mockPluginClient struct{}

func (mp *mockPluginClient) Ping() error         { return nil }
func (mp *mockPluginClient) Kill(r string) error { return nil }

func TestMonitor(t *testing.T) {
	Convey("monitor", t, func() {
		aps := newAvailablePlugins()

		ap1 := &availablePlugin{
			Type:    plugin.CollectorPluginType,
			Version: 1,
			Name:    "test",
			Client:  &mockPluginClient{},

			eventManager: &gomit.EventController{},
		}
		ap1.makeKey()
		aps.Insert(ap1)

		ap2 := &availablePlugin{
			Type:    plugin.PublisherPluginType,
			Version: 1,
			Name:    "test",
			Client:  &mockPluginClient{},

			eventManager: &gomit.EventController{},
		}
		ap2.makeKey()
		aps.Insert(ap2)

		ap3 := &availablePlugin{
			Type:    plugin.ProcessorPluginType,
			Version: 1,
			Name:    "test",
			Client:  &mockPluginClient{},

			eventManager: &gomit.EventController{},
		}
		ap3.makeKey()
		aps.Insert(ap3)

		Convey("newMonitor", func() {
			m := newMonitor()
			So(m, ShouldHaveSameTypeAs, &monitor{})
		})
		Convey("start", func() {
			m := newMonitor()
			m.Start(aps)
			So(m.State, ShouldEqual, MonitorStarted)
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
			oldOpt := m.Option(MonitorDuration(time.Millisecond * 200))
			So(m.duration, ShouldResemble, time.Millisecond*200)
			m.Option(oldOpt)
			So(m.duration, ShouldResemble, time.Second*60)
		})
	})
}
