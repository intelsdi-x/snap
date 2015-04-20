package scheduler

import (
	"os"
	"path"
	"testing"
	"time"

	"github.com/intelsdilabs/pulse/control"
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

func TestWorkflow(t *testing.T) {
	Convey("Workflow", t, func() {
		wf := newWorkflow()
		So(wf.state, ShouldNotBeNil)
		Convey("Add steps", func() {
			pubStep := new(publishStep)
			procStep := new(processStep)
			wf.rootStep.AddStep(pubStep).AddStep(procStep)
			So(wf.rootStep, ShouldNotBeNil)
			So(wf.rootStep.Steps(), ShouldNotBeNil)
			Convey("Start", func() {
				workerKillChan = make(chan struct{})
				manager := newWorkManager()
				sch := newSimpleSchedule(core.NewSimpleSchedule(time.Duration(5 * time.Second)))
				task := newTask(sch, []core.MetricType{}, &mockWorkflow{}, manager, &mockMetricManager{})
				wf.Start(task)
				So(wf.State(), ShouldEqual, core.WorkflowStarted)
			})
		})
	})
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

			c.SubscribePublisher("file", 1)

			cd := cdata.NewNode()
			cd.AddItem("password", &ctypes.ConfigValueStr{Value: "value"})
			mt := MockMetricType{namespace: []string{"intel", "dummy", "foo"}}
			_, errs := c.SubscribeMetricType(mt, cd)
			So(errs, ShouldBeNil)

			sch := newSimpleSchedule(core.NewSimpleSchedule(time.Duration(1 * time.Second)))
			So(sch, ShouldNotBeNil)

			Convey("Workflow", func() {
				wf := newWorkflow()
				So(wf.state, ShouldNotBeNil)
				Convey("Add steps", func() {
					pubStep := NewPublishStep("file", 1)
					wf.rootStep.AddStep(pubStep)
					So(wf.rootStep, ShouldNotBeNil)
					So(wf.rootStep.Steps(), ShouldNotBeNil)
					Convey("Start", func() {
						workerKillChan = make(chan struct{})
						manager := newWorkManager()
						task := newTask(sch, []core.MetricType{mt}, wf, manager, c)
						task.Spin()
						time.Sleep(4 * time.Second)
					})
				})
			})

		})
	})
}
