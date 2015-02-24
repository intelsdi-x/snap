package schedule

import (
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
)

type MockSchedule struct {
	tick chan struct{}
}

func (m *MockSchedule) Wait(t time.Time) chan struct{} {
	return m.tick
}

func (m *MockSchedule) Tick() {
	m.tick <- struct{}{}
}

func TestTask(t *testing.T) {
	Convey("Task", t, func() {
		mockSchedule := &MockSchedule{
			tick: make(chan struct{}),
		}
		Convey("Task is created and starts to spin", func() {
			task := NewTask(mockSchedule)
			task.Spin()
			So(task.state, ShouldEqual, TaskSpinning)
			Convey("Tick arrives from the schedule", func() {
				mockSchedule.Tick()
				So(task.state, ShouldEqual, TaskFiring)
				Convey("Calling tick while task is running results in noop", func() {
					task.state = TaskFiring
					task.fire()
					So(task.state, ShouldEqual, TaskFiring)
				})
			})
			Convey("Task is Stopped", func() {
				task.Stop()
				So(task.state, ShouldEqual, TaskStopped)
				Convey("Stopping a stopped tasks has no effect", func() {
					task.Stop()
					So(task.state, ShouldEqual, TaskStopped)
				})
			})
		})

	})
}
