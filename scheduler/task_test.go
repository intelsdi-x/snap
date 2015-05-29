package scheduler

import (
	"testing"
	"time"

	"github.com/intelsdi-x/pulse/control/plugin"
	"github.com/intelsdi-x/pulse/core"
	"github.com/intelsdi-x/pulse/pkg/schedule"
	"github.com/intelsdi-x/pulse/scheduler/wmap"

	. "github.com/smartystreets/goconvey/convey"
)

func TestTask(t *testing.T) {
	Convey("Task", t, func() {
		sampleWFMap := wmap.Sample()
		wf, errs := wmapToWorkflow(sampleWFMap)
		So(errs, ShouldBeEmpty)
		c := &mockMetricManager{}
		c.setAcceptedContentType("rabbitmq", core.PublisherPluginType, 5, []string{plugin.PulseGOBContentType})
		err := wf.BindPluginContentTypes(c)
		So(err, ShouldBeNil)
		Convey("task + simple schedule", func() {
			sch := schedule.NewSimpleSchedule(time.Millisecond * 100)
			task := newTask(sch, []core.Metric{}, wf, newWorkManager(), c)
			task.Spin()
			time.Sleep(time.Millisecond * 10) // it is a race so we slow down the test
			So(task.state, ShouldEqual, core.TaskSpinning)
			task.Stop()
		})

		Convey("Task deadline duration test", func() {
			sch := schedule.NewSimpleSchedule(time.Millisecond * 100)
			task := newTask(sch, []core.Metric{}, wf, newWorkManager(), c, core.TaskDeadlineDuration(20*time.Second))
			task.Spin()
			So(task.deadlineDuration, ShouldEqual, 20*time.Second)
			task.Option(core.TaskDeadlineDuration(20 * time.Second))

			So(core.TaskDeadlineDuration(2*time.Second), ShouldNotBeEmpty)

		})

		Convey("Tasks are created and creation of task table is checked", func() {
			sch := schedule.NewSimpleSchedule(time.Millisecond * 100)
			task := newTask(sch, []core.Metric{}, wf, newWorkManager(), c)
			task1 := newTask(sch, []core.Metric{}, wf, newWorkManager(), c)
			task1.Spin()
			task.Spin()
			tC := newTaskCollection()
			tC.add(task)
			tC.add(task1)
			taskTable := tC.Table()

			So(len(taskTable), ShouldEqual, 2)

		})

		Convey("Task is created and starts to spin", func() {
			sch := schedule.NewSimpleSchedule(time.Second * 5)
			task := newTask(sch, []core.Metric{}, wf, newWorkManager(), c)
			task.Spin()
			So(task.state, ShouldEqual, core.TaskSpinning)
			Convey("Task is Stopped", func() {
				task.Stop()
				time.Sleep(time.Millisecond * 10) // it is a race so we slow down the test
				So(task.state, ShouldEqual, core.TaskStopped)
				Convey("Stopping a stopped tasks should not send to kill channel", func() {
					task.Stop()
					b := false
					select {
					case <-task.killChan:
						b = true
					default:
						b = false
					}
					So(task.state, ShouldEqual, core.TaskStopped)
					So(b, ShouldBeFalse)
				})
			})
		})

		Convey("task fires", func() {
			sch := schedule.NewSimpleSchedule(time.Millisecond * 10)
			task := newTask(sch, []core.Metric{}, wf, newWorkManager(), c)
			task.Spin()
			time.Sleep(time.Millisecond * 100)
			So(task.hitCount, ShouldBeGreaterThan, 2)
			So(task.missedIntervals, ShouldBeGreaterThan, 2)
			task.Stop()
		})
	})

	Convey("Create task collection", t, func() {
		sampleWFMap := wmap.Sample()
		wf, errs := wmapToWorkflow(sampleWFMap)
		So(errs, ShouldBeEmpty)

		sch := schedule.NewSimpleSchedule(time.Millisecond * 10)
		task := newTask(sch, []core.Metric{}, wf, newWorkManager(), &mockMetricManager{})
		So(task.id, ShouldNotBeEmpty)
		So(task.id, ShouldNotBeNil)
		taskCollection := newTaskCollection()

		Convey("Add task to collection", func() {

			err := taskCollection.add(task)
			So(err, ShouldBeNil)
			So(len(taskCollection.table), ShouldEqual, 1)

			Convey("Attempt to add the same task again", func() {
				err := taskCollection.add(task)
				So(err, ShouldNotBeNil)
			})

			Convey("Get task from collection", func() {
				t := taskCollection.Get(task.id)
				So(t, ShouldNotBeNil)
				So(t.Id(), ShouldEqual, task.id)
				So(t.CreationTime().Nanosecond(), ShouldBeLessThan, time.Now().Nanosecond())
				So(t.HitCount(), ShouldEqual, 0)
				So(t.MissedCount(), ShouldEqual, 0)
				So(t.State(), ShouldEqual, core.TaskStopped)
				So(t.Status(), ShouldEqual, core.WorkflowStopped)
				So(t.LastRunTime().IsZero(), ShouldBeTrue)
			})

			Convey("Attempt to get task with an invalid Id", func() {
				t := taskCollection.Get(1234)
				So(t, ShouldBeNil)
			})

			Convey("Create another task and compare the id", func() {
				task2 := newTask(sch, []core.Metric{}, wf, newWorkManager(), &mockMetricManager{})
				So(task2.id, ShouldEqual, task.id+1)
			})

		})

	})
}
