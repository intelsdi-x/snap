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
		Convey("nominal window with a few misses", func() {
			startWait := time.Millisecond * 50
			windowSize := time.Millisecond * 200
			interval := time.Millisecond * 10
			// shouldWait := 1000.0 + float64(interval)

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
			// we should have either 16 or 17 minus 3 missed
			So(len(r), ShouldBeBetweenOrEqual, 15, 17)

			var missed uint
			for _, x := range r {
				missed += x.Missed()
			}
			So(
				r[0].LastTime().Sub(before).Seconds(),
				ShouldBeBetweenOrEqual,
				(startWait+interval).Seconds()*.9,
				(startWait+interval).Seconds()*1.5,
			)
		})

		Convey("started in the middle of the window", func() {
			startWait := time.Millisecond * -200
			windowSize := time.Millisecond * 350
			interval := time.Millisecond * 10
			// shouldWait := 1000.0 + float64(interval)

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
			// we should have either 16 or 17 minus 3 missed
			So(len(r), ShouldBeBetweenOrEqual, 10, 12)

			var missed uint
			for _, x := range r {
				missed += x.Missed()
			}
			So(missed, ShouldBeBetweenOrEqual, 22, 24)
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
			So(
				r[0].LastTime().Sub(before).Seconds(),
				ShouldBeBetweenOrEqual,
				(startWait+interval).Seconds()*.9,
				(startWait+interval).Seconds()*1.5,
			)
		})

		Convey("stop without start", func() {
			startWait := time.Millisecond * 50
			windowSize := time.Millisecond * 200
			interval := time.Millisecond * 10

			stop := time.Now().Add(startWait + windowSize)
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
			So(
				r[0].LastTime().Sub(before).Seconds(),
				ShouldBeBetweenOrEqual,
				(interval).Seconds()*.9,
				(interval).Seconds()*1.5,
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
