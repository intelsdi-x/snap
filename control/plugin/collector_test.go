package plugin

import (
	"testing"

	"github.com/intelsdi-x/pulse/control/plugin/cpolicy"
	. "github.com/smartystreets/goconvey/convey"
)

type MockPlugin struct {
	Meta PluginMeta
}

func (f *MockPlugin) GetConfigPolicy() (*cpolicy.ConfigPolicy, error) {
	return &cpolicy.ConfigPolicy{}, nil
}

func (f *MockPlugin) CollectMetrics(_ []PluginMetricType) ([]PluginMetricType, error) {
	return []PluginMetricType{}, nil
}

func (c *MockPlugin) GetMetricTypes() ([]PluginMetricType, error) {
	return []PluginMetricType{
		PluginMetricType{Namespace_: []string{"foo", "bar"}},
	}, nil
}

func TestStartCollector(t *testing.T) {
	Convey("Collector", t, func() {
		Convey("start with dynamic port", func() {
			m := &PluginMeta{
				RPCType: JSONRPC,
				Type:    CollectorPluginType,
			}
			c := new(MockPlugin)
			err, rc := Start(m, c, "{}")
			So(err, ShouldBeNil)
			So(rc, ShouldEqual, 0)
			Convey("RPC service already registered", func() {
				So(func() { Start(m, c, "{}") }, ShouldPanic)
			})
		})
	})
}
