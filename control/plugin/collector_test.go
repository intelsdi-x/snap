package plugin

import (
	"encoding/json"
	"time"

	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

type MockPlugin struct {
	Meta   PluginMeta
	Policy ConfigPolicy
}

func (f *MockPlugin) Collect(args CollectorArgs, reply *CollectorReply) error {
	return nil
}

func (c *MockPlugin) GetMetricTypes(args GetMetricTypesArgs, reply *GetMetricTypesReply) error {
	reply.MetricTypes = []*MetricType{
		NewMetricType([]string{"org", "some_metric"}, time.Now().Unix()),
	}
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
		err, _ := StartCollector(&m.Meta, m.Collector, &m.Policy, "/tmp/foo", string(b))

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

		err, _ := StartCollector(&m.Meta, m.Collector, &m.Policy, "/tmp/foo", string(b))
		So(err, ShouldBeNil)
	})

}
