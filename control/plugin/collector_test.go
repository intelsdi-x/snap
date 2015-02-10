package plugin

import (
	"os"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

type MockCollectorPlugin struct{}

func (c *MockCollectorPlugin) Collect(args CollectorArgs, reply *CollectorReply) error {
	return nil
}

func (c *MockCollectorPlugin) GetMetricTypes(args GetMetricTypesArgs, reply *GetMetricTypesReply) error {
	reply.MetricTypes = []*MetricType{
		NewMetricType([]string{"org", "some_metric"}),
	}
	return nil
}

func TestCollector(t *testing.T) {

	Convey("StartCollector", t, func() {
		os.Args = []string{"", "{\"ListenPort\": \"9998\", \"RunAsDaemon\": false}"}
		meta := &PluginMeta{
			Name:    "test1",
			Version: 1,
		}
		policy := new(ConfigPolicy)

		e := StartCollector(meta, new(MockCollectorPlugin), policy)
		So(e, ShouldBeNil)
	})

}
