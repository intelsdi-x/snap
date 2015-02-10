package plugin

import (
	"encoding/json"
	"time"

	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

type MockPlugin struct {
	Meta      PluginMeta
	Collector CollectorPlugin
	Policy    ConfigPolicy
}

func (f *MockPlugin) Collect(args CollectorArgs, reply *CollectorReply) error {
	return nil
}

func TestStartCollector(t *testing.T) {
	Convey("Daemon mode", t, func() {
		// These setting ensure it exists before test timeout
		PingTimeoutDuration = time.Millisecond * 100
		PingTimeoutLimit = 1

		r := Arg{RunAsDaemon: true}
		b, e := json.Marshal(r)
		So(e, ShouldBeNil)

		m := new(MockPlugin)
		m.Meta.Name = "mock"
		m.Meta.Version = 1

		err := StartCollector(&m.Meta, m.Collector, &m.Policy, "/tmp/foo", string(b))
		So(err, ShouldBeNil)
	})

	Convey("Non-daemon mode", t, func() {
		// These setting ensure it exists before test timeout
		PingTimeoutDuration = time.Millisecond * 100
		PingTimeoutLimit = 1

		r := Arg{RunAsDaemon: false}
		b, e := json.Marshal(r)
		So(e, ShouldBeNil)

		m := new(MockPlugin)
		m.Meta.Name = "mock"
		m.Meta.Version = 1

		err := StartCollector(&m.Meta, m.Collector, &m.Policy, "/tmp/foo", string(b))
		So(err, ShouldBeNil)
	})

}
