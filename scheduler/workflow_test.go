package scheduler

import (
	"testing"
	"time"

	"github.com/intelsdilabs/pulse/core"

	. "github.com/smartystreets/goconvey/convey"
)

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
				manager := newWorkManager(int64(5), 1)
				sch := newSimpleSchedule(core.NewSimpleSchedule(time.Duration(5 * time.Second)))
				task := newTask(sch, []core.MetricType{}, &mockWorkflow{}, manager)
				wf.Start(task)
				So(wf.State(), ShouldEqual, core.WorkflowStarted)
			})
		})
	})
}
