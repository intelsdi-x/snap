package control

import (
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

func TestMonitorState(t *testing.T) {
	Convey("pulse/control", t, func() {
		Convey("Runner", func() {
			Convey(".Start", func() {
				Convey("add the monitor", func() {
					r := NewRunner()
					r.AddDelegates(new(MockHandlerDelegate))
					err := r.Start()
					So(err, ShouldBeNil)
					So(r.monitor, ShouldNotBeNil)
					So(r.monitor.State, ShouldEqual, MonitorStarted)
					r.monitor.Stop()
					So(r.monitor.State, ShouldEqual, MonitorStopped)
				})
			})
		})
	})
}
