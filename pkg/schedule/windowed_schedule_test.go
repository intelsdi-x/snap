// +build legacy

package schedule

import (
	"testing"
	"time"

	log "github.com/Sirupsen/logrus"

	. "github.com/smartystreets/goconvey/convey"
)

func TestWindowedSchedule(t *testing.T) {
	log.SetLevel(log.DebugLevel)
	Convey("Windowed Schedule", t, func() {
		Convey("nominal window without misses", func() {
			startWait := time.Millisecond * 50
			windowSize := time.Millisecond * 200
			interval := time.Millisecond * 10

			start := time.Now().Add(startWait)
			stop := time.Now().Add(startWait + windowSize)
			w := NewWindowedSchedule(
				interval,
				&start,
				&stop,
			)

			err := w.Validate()
			So(err, ShouldBeNil)

			var r []Response
			last := *new(time.Time)

			state := Active
			before := time.Now()
			for state == Active {
				r1 := w.Wait(last)
				state = r1.State()
				last = time.Now()
				r = append(r, r1)
			}

			// there are 0 missed responses, so for this schedule
			// we expect to get between 19 - 22 responses
			So(len(r), ShouldBeBetweenOrEqual, 19, 22)

			var missed uint
			for _, x := range r {
				missed += x.Missed()
			}
			So(missed, ShouldEqual, 0)

			// the task is expected to fire immediately on determined start-time
			So(
				r[0].LastTime().Sub(before).Seconds(),
				ShouldBeBetweenOrEqual,
				(startWait).Seconds(),
				(startWait + interval).Seconds(),
			)
		})

		Convey("nominal window with a few misses", func() {
			startWait := time.Millisecond * 50
			windowSize := time.Millisecond * 200
			interval := time.Millisecond * 10

			start := time.Now().Add(startWait)
			stop := time.Now().Add(startWait + windowSize)
			w := NewWindowedSchedule(
				interval,
				&start,
				&stop,
			)

			err := w.Validate()
			So(err, ShouldBeNil)

			var r []Response
			last := *new(time.Time)

			state := Active
			before := time.Now()
			for state == Active {
				r1 := w.Wait(last)
				state = r1.State()
				last = time.Now()
				r = append(r, r1)
				// make it miss some
				if len(r) == 3 || len(r) == 7 {
					time.Sleep(w.Interval)
				}
				if len(r) == 9 {
					// Miss two
					time.Sleep(w.Interval * 2)
				}
			}

			// there are 4 missed responses, so for this schedule
			// we expect to get between 15 - 18 responses
			So(len(r), ShouldBeBetweenOrEqual, 15, 18)

			var missed uint
			for _, x := range r {
				missed += x.Missed()
			}
			So(missed, ShouldEqual, 4)

			// the task is expected to fire immediately on determined start-time
			So(
				r[0].LastTime().Sub(before).Seconds(),
				ShouldBeBetweenOrEqual,
				(startWait).Seconds(),
				(startWait + interval).Seconds(),
			)
		})

		Convey("started in the past", func() {
			startWait := time.Millisecond * -200
			windowSize := time.Millisecond * 400
			interval := time.Millisecond * 10

			start := time.Now().Add(startWait)
			stop := time.Now().Add(startWait + windowSize)
			w := NewWindowedSchedule(
				interval,
				&start,
				&stop,
			)

			err := w.Validate()
			So(err, ShouldBeNil)

			var r []Response
			last := *new(time.Time)

			before := time.Now()
			state := Active
			for state == Active {
				r1 := w.Wait(last)
				state = r1.State()
				last = time.Now()
				r = append(r, r1)
				// make it miss some
				if len(r) == 3 || len(r) == 7 {
					time.Sleep(w.Interval)
				}
				if len(r) == 9 {
					// Miss two
					time.Sleep(w.Interval * 2)
				}
			}
			// there are 4 missed responses, so for this schedule
			// we expect to get between 15 - 18 responses
			So(len(r), ShouldBeBetweenOrEqual, 15, 18)

			var missed uint
			for _, x := range r {
				missed += x.Missed()
			}
			So(missed, ShouldEqual, 4)

			// start_time points to the past,
			// so the task is expected to fire immediately
			So(
				r[0].LastTime().Sub(before).Seconds(),
				ShouldBeBetweenOrEqual,
				0,
				(interval).Seconds(),
			)
		})

		Convey("start without stop", func() {
			startWait := time.Millisecond * 50
			interval := time.Millisecond * 10

			start := time.Now().Add(startWait)
			w := NewWindowedSchedule(
				interval,
				&start,
				nil,
			)

			err := w.Validate()
			So(err, ShouldBeNil)

			var r []Response
			last := *new(time.Time)

			before := time.Now()
			for len(r) <= 10 {
				r1 := w.Wait(last)
				last = time.Now()
				r = append(r, r1)
			}

			// the task is expected to fire immediately on start_time
			So(
				r[0].LastTime().Sub(before).Seconds(),
				ShouldBeBetweenOrEqual,
				(startWait).Seconds(),
				(startWait + interval).Seconds(),
			)
		})

		Convey("stop without start", func() {
			windowSize := time.Millisecond * 200
			interval := time.Millisecond * 10

			stop := time.Now().Add(windowSize)
			w := NewWindowedSchedule(
				interval,
				nil,
				&stop,
			)

			err := w.Validate()
			So(err, ShouldBeNil)

			var r []Response
			last := *new(time.Time)

			before := time.Now()
			for len(r) <= 10 {
				r1 := w.Wait(last)
				last = time.Now()
				r = append(r, r1)
			}

			// the task should start immediately
			So(
				r[0].LastTime().Sub(before).Seconds(),
				ShouldBeBetweenOrEqual,
				0,
				(interval).Seconds(),
			)

			var missed uint
			for _, x := range r {
				missed += x.Missed()
			}
			So(missed, ShouldEqual, 0)
		})

		Convey("start immediatelly without stop (no window determined)", func() {
			interval := time.Millisecond * 10
			// schedule equivalent to simple schedule
			w := NewWindowedSchedule(
				interval,
				nil,
				nil,
			)

			err := w.Validate()
			So(err, ShouldBeNil)

			var r []Response
			last := *new(time.Time)

			before := time.Now()
			for len(r) <= 10 {
				r1 := w.Wait(last)
				last = time.Now()
				r = append(r, r1)
			}

			// the task should start immediately
			So(
				r[0].LastTime().Sub(before).Seconds(),
				ShouldBeBetweenOrEqual,
				0,
				(interval).Seconds(),
			)

		})

		Convey("start time in past is ok (as long as window ends in the future)", func() {
			start := time.Now().Add(time.Second * -10)
			stop := time.Now().Add(time.Second * 10)
			w := NewWindowedSchedule(time.Millisecond*100, &start, &stop)
			err := w.Validate()
			So(err, ShouldEqual, nil)
		})

		Convey("window in past", func() {
			start := time.Now().Add(time.Second * -20)
			stop := time.Now().Add(time.Second * -10)
			w := NewWindowedSchedule(time.Millisecond*100, &start, &stop)
			err := w.Validate()
			So(err, ShouldEqual, ErrInvalidStopTime)
		})

		Convey("cart before the horse", func() {
			start := time.Now().Add(time.Second * 100)
			stop := time.Now().Add(time.Second * 10)
			w := NewWindowedSchedule(time.Millisecond*100, &start, &stop)
			err := w.Validate()
			So(err, ShouldEqual, ErrStopBeforeStart)
		})

	})
}
