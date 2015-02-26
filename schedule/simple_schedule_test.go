package schedule

import (
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
)

func TestSimpleSchedule(t *testing.T) {
	Convey("test Wait()", t, func() {
		interval := 100
		overage := 467
		shouldWait := float64(500 - overage)
		last := time.Now()

		time.Sleep(time.Millisecond * time.Duration(overage))
		s := NewSimpleSchedule(time.Millisecond * time.Duration(interval))

		before := time.Now()
		r := s.Wait(last)
		after := time.Now().Sub(before)

		So(r.State(), ShouldEqual, ScheduleActive)
		So(r.MissedIntervals(), ShouldResemble, 4)
		So(r.Error(), ShouldEqual, nil)
		// We are ok at this precision with being within 10% over or under (10ms)
		afterMS := after.Nanoseconds() / 1000 / 1000
		So(afterMS, ShouldBeGreaterThan, shouldWait-10)
		So(afterMS, ShouldBeLessThan, shouldWait+10)
	})
}
