package control

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestMonitorState(t *testing.T) {
	Convey("pulse/control", t, func() {
		Convey("Runner", func() {
			Convey(".Start", func() {
				Convey("add the monitor", func() {
					r := newRunner()
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
