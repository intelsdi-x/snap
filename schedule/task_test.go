package schedule

import (
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
)

type MockSchedule struct {
	tick bool
}

func (m *MockSchedule) Wait(t time.Time) ScheduleResponse {
	for !m.tick {
		time.Sleep(time.Millisecond * 100)
	}
	m.tick = false
	return ScheduleResponse{State: ScheduleActive}
}

func (m *MockSchedule) Tick() {
	m.tick = true
}

func TestTask(t *testing.T) {
	Convey("Task", t, func() {
		mockSchedule := &MockSchedule{
			tick: false,
		}
		Convey("Task is created and starts to spin", func() {
			task := NewTask(mockSchedule)
			task.Spin()
			time.Sleep(time.Millisecond * 200) // it is a race so we slow down the test
			So(task.state, ShouldEqual, TaskSpinning)
			Convey("Tick arrives from the schedule", func() {
				mockSchedule.Tick()
				time.Sleep(time.Millisecond * 200) // it is a race so we slow down the test
				So(task.state, ShouldEqual, TaskFiring)
			})
			Convey("Task is Stopped", func() {
				task.Stop()
				time.Sleep(time.Millisecond * 200) // it is a race so we slow down the test
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

	})
}
