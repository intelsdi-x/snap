package schedule

import (
	"errors"
	"testing"
	"time"

	"github.com/intelsdilabs/pulse/core"

	. "github.com/smartystreets/goconvey/convey"
)

type MockSchedule struct {
	tick                   bool
	failValidatingSchedule bool
}

type MockScheduleResponse struct {
}

func (m MockScheduleResponse) State() ScheduleState {
	return ScheduleActive
}

func (m MockScheduleResponse) Error() error {
	return nil
}

func (m MockScheduleResponse) MissedIntervals() uint {
	return 0
}

func (m *MockSchedule) Wait(t time.Time) ScheduleResponse {
	for !m.tick {
		time.Sleep(time.Millisecond * 100)
	}
	m.tick = false
	return MockScheduleResponse{}
}

func (m *MockSchedule) Tick() {
	m.tick = true
}

func (m *MockSchedule) Validate() error {
	if m.failValidatingSchedule {
		return errors.New("schedule error")
	}
	return nil
}

func TestTask(t *testing.T) {
	Convey("Task", t, func() {
		Convey("task + simple schedule", func() {
			sch := NewSimpleSchedule(time.Millisecond * 100)
			task := newTask(sch, []core.MetricType{}, &mockWorkflow{}, newWorkManager(int64(5), 1))
			task.Spin()
			time.Sleep(time.Millisecond * 10) // it is a race so we slow down the test
			So(task.state, ShouldEqual, TaskSpinning)
			task.Stop()
		})

		Convey("Task is created and starts to spin", func() {
			mockSchedule := &MockSchedule{
				tick: false,
			}
			task := newTask(mockSchedule, []core.MetricType{}, &mockWorkflow{}, newWorkManager(int64(5), 1))
			task.Spin()
			time.Sleep(time.Millisecond * 10) // it is a race so we slow down the test
			So(task.state, ShouldEqual, TaskSpinning)
			Convey("Task is Stopped", func() {
				task.Stop()
				time.Sleep(time.Millisecond * 10) // it is a race so we slow down the test
				So(task.state, ShouldEqual, TaskStopped)
				Convey("Stopping a stopped tasks should not send to kill channel", func() {
					task.Stop()
					b := false
					select {
					case <-task.killChan:
						b = true
					default:
						b = false
					}
					So(task.state, ShouldEqual, TaskStopped)
					So(b, ShouldBeFalse)
				})
			})
		})

		Convey("task fires", func() {
			sch := NewSimpleSchedule(time.Millisecond * 10)
			task := newTask(sch, []core.MetricType{}, &mockWorkflow{}, newWorkManager(int64(5), 1))
			task.Spin()
			time.Sleep(time.Millisecond * 100)
			So(task.hitCount, ShouldBeGreaterThan, 2)
			So(task.missedIntervals, ShouldBeGreaterThan, 2)
			task.Stop()
		})
	})

	Convey("Create task collection", t, func() {

		sch := NewSimpleSchedule(time.Millisecond * 10)
		task := newTask(sch, []core.MetricType{}, &mockWorkflow{}, newWorkManager(int64(5), 1))
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
				So(t.State(), ShouldEqual, TaskStopped)
				So(t.Status(), ShouldEqual, WorkflowStopped)
				So(t.LastRunTime().IsZero(), ShouldBeTrue)
			})

			Convey("Attempt to get task with an invalid Id", func() {
				t := taskCollection.Get("1234")
				So(t, ShouldBeNil)
			})

		})

	})
}
