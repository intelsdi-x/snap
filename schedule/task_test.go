package schedule

import (
	"errors"
	"testing"
	"time"

	"github.com/intelsdilabs/pulse/core/cdata"
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

func (m MockScheduleResponse) MissedIntervals() int {
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
		Convey("given a core.MetricType and config create a new MetrictType", func() {
			m := &MockMetricType{
				namespace:               []string{"foo", "bar"},
				version:                 1,
				lastAdvertisedTimestamp: 0,
			}
			config := cdata.NewNode()
			mt := newMetricType(m, config)
			So(mt.Config(), ShouldNotBeNil)
			So(mt.LastAdvertisedTimestamp(), ShouldEqual, 0)
		})

		Convey("task + simple schedule", func() {
			sch := NewSimpleSchedule(time.Millisecond * 100)
			task := NewTask(sch, nil)
			task.Spin()
			time.Sleep(time.Millisecond * 10) // it is a race so we slow down the test
			So(task.state, ShouldEqual, TaskSpinning)
		})

		Convey("Task is created and starts to spin", func() {
			mockSchedule := &MockSchedule{
				tick: false,
			}
			task := NewTask(mockSchedule, nil)
			task.Spin()
			time.Sleep(time.Millisecond * 10) // it is a race so we slow down the test
			So(task.state, ShouldEqual, TaskSpinning)
			Convey("Tick arrives from the schedule", func() {
				mockSchedule.Tick()
				time.Sleep(time.Millisecond * 200) // it is a race so we slow down the test
				So(task.state, ShouldEqual, TaskFiring)
			})
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

	})
}
