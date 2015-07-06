package scheduler

import (
	"testing"

	"github.com/intelsdi-x/pulse/scheduler/wmap"

	log "github.com/Sirupsen/logrus"
	. "github.com/smartystreets/goconvey/convey"
)

func TestWorkflowString(t *testing.T) {
	log.SetLevel(log.FatalLevel)
	Convey("String", t, func() {
		w := wmap.NewWorkflowMap()
		w.CollectNode.AddMetric("fall", 1)
		pr1 := wmap.NewProcessNode("winter", 1)
		pr2 := wmap.NewProcessNode("summer", 1)
		pu1 := wmap.NewPublishNode("spring", 1)
		pu2 := wmap.NewPublishNode("autumn", 1)
		w.CollectNode.AddConfigItem("/foo/bar", "user", "rain")
		pr1.AddConfigItem("leaves", 1)
		pr2.AddConfigItem("flowers", 2)
		pu2.AddConfigItem("grass", 3)
		w.CollectNode.Add(pr1)
		w.CollectNode.Add(pu1)
		w.CollectNode.ProcessNodes[0].Add(pr2)
		w.CollectNode.ProcessNodes[0].Add(pu2)

		wf, err := wmapToWorkflow(w)
		So(err, ShouldBeNil)
		str := wf.String()
		//fmt.Printf("%v", str)
		So(str, ShouldNotBeEmpty)
		str2 := wf.processNodes[0].String("")
		So(str2, ShouldNotBeEmpty)
		metricString("", wf.metrics)
	})
}
