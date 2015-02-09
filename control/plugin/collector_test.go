package plugin

import (
	"bytes"
	. "github.com/smartystreets/goconvey/convey"
	"io"
	"os"
	"testing"
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

func TestStartCollector(t *testing.T) {

	Convey("collector.StartCollector", t, func() {
		os.Args = []string{"", "{\"ListenPort\": \"9998\", \"RunAsDaemon\": false}"}
		meta := &PluginMeta{
			Name:    "test1",
			Version: 1,
		}
		policy := new(ConfigPolicy)
		old := os.Stdout // keep backup of the real stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		outC := make(chan string)
		// copy the output in a separate goroutine so printing can't block indefinitely
		go func() {
			var buf bytes.Buffer
			io.Copy(&buf, r)
			outC <- buf.String()
		}()

		e := StartCollector(meta, new(MockCollectorPlugin), policy)
		So(e, ShouldBeNil)

		// back to normal state
		w.Close()
		os.Stdout = old // restoring the real stdout
		out := <-outC
		println(out)
		So(out, ShouldContainSubstring, "some_metric")
	})
}
