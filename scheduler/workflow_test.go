package scheduler

import (
	"os"
	"path"
	"testing"
	"time"

	"github.com/intelsdilabs/pulse/control"
	"github.com/intelsdilabs/pulse/control/plugin"
	"github.com/intelsdilabs/pulse/core"
	"github.com/intelsdilabs/pulse/core/cdata"
	"github.com/intelsdilabs/pulse/core/ctypes"
	"github.com/intelsdilabs/pulse/pkg/logger"

	. "github.com/smartystreets/goconvey/convey"
)

var (
	PulsePath = os.Getenv("PULSE_PATH")
)

type MockMetricType struct {
	namespace []string
}

func (m MockMetricType) Namespace() []string {
	return m.namespace
}

func (m MockMetricType) LastAdvertisedTime() time.Time {
	return time.Now()
}

func (m MockMetricType) Version() int {
	return 1
}

func (m MockMetricType) Config() *cdata.ConfigDataNode {
	return nil
}

func (m MockMetricType) Data() interface{} {
	return nil
}

func TestCollectPublishWorkflow(t *testing.T) {
	Convey("Given a started plugin control", t, func() {
		logger.SetLevel(logger.DebugLevel)
		c := control.New()
		c.Start()
		Convey("Start a collector and publisher plugin", func() {
			err := c.Load(path.Join(PulsePath, "plugin", "collector", "pulse-collector-dummy1"))
			So(err, ShouldBeNil)
			err = c.Load(path.Join(PulsePath, "plugin", "publisher", "pulse-publisher-file"))
			So(err, ShouldBeNil)
			time.Sleep(100 * time.Millisecond)

			config := map[string]ctypes.ConfigValue{
				"file": ctypes.ConfigValueStr{Value: "/tmp/pulse-TestCollectPublishWorkflow.out"},
			}
			c.SubscribePublisher("file", 1, config)

			cd := cdata.NewNode()
			cd.AddItem("password", &ctypes.ConfigValueStr{Value: "value"})
			mt := MockMetricType{namespace: []string{"intel", "dummy", "foo"}}
			smt, errs := c.SubscribeMetricType(mt, cd)
			So(errs, ShouldBeNil)

			sch := newSimpleSchedule(core.NewSimpleSchedule(time.Duration(1 * time.Second)))
			So(sch, ShouldNotBeNil)

			Convey("Workflow", func() {
				wf := newWorkflow()
				So(wf.state, ShouldNotBeNil)
				Convey("Add steps", func() {
					pubStep := NewPublishStep("file", 1, plugin.PulseGOBContentType, config)
					wf.rootStep.AddStep(pubStep)
					So(wf.rootStep, ShouldNotBeNil)
					So(wf.rootStep.Steps(), ShouldNotBeNil)
					Convey("Start", func() {
						workerKillChan = make(chan struct{})
						manager := newWorkManager()
						task := newTask(sch, []core.Metric{smt}, wf, manager, c)
						task.Spin()
						time.Sleep(4 * time.Second)
					})
				})
			})

		})
	})
}
