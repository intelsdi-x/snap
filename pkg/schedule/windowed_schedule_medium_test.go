// +build medium

/*
http://www.apache.org/licenses/LICENSE-2.0.txt


Copyright 2015 Intel Corporation

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package schedule

import (
	"testing"
	"time"

	log "github.com/sirupsen/logrus"

	. "github.com/smartystreets/goconvey/convey"
)

func TestWindowedSchedule(t *testing.T) {
	log.SetLevel(log.DebugLevel)

	Convey("Windowed Schedule expected to run forever", t, func() {
		interval := time.Millisecond * 20
		// set start and stop are nil, and the count is zero what means no limits
		w := NewWindowedSchedule(interval, nil, nil, 0)

		err := w.Validate()
		So(err, ShouldBeNil)

		Convey("with no misses ", func() {
			var r []Response
			last := *new(time.Time)
			before := time.Now()

			for len(r) <= 10 {
				r1 := w.Wait(last)
				last = time.Now()
				r = append(r, r1)
			}

			var missed uint
			for _, x := range r {
				missed += x.Missed()
			}
			So(missed, ShouldEqual, 0)

			// the task is expected to fire immediately
			So(
				r[0].LastTime().Sub(before).Seconds(),
				ShouldBeLessThan,
				interval.Seconds(),
			)
		})
		Convey("with a few misses ", func() {
			var r []Response
			last := *new(time.Time)

			before := time.Now()

			for len(r) <= 10 {
				r1 := w.Wait(last)
				last = time.Now()
				r = append(r, r1)
				// make it miss some
				if len(r) == 3 || len(r) == 7 {
					time.Sleep(w.Interval)
				}
				if len(r) == 9 {
					// miss two
					time.Sleep(2 * w.Interval)
				}
			}
			var missed uint
			for _, x := range r {
				missed += x.Missed()
			}
			So(missed, ShouldEqual, 4)

			// the task is expected to fire immediately
			So(
				r[0].LastTime().Sub(before).Seconds(),
				ShouldBeLessThan,
				interval.Seconds(),
			)
		})
	}) // the end of `Simple Windowed Schedule expected to run forever`

	Convey("Nominal windowed Schedule", t, func() {
		Convey("without misses", func() {
			startWait := time.Millisecond * 50
			windowSize := time.Millisecond * 400
			interval := time.Millisecond * 20

			start := time.Now().Add(startWait)
			stop := time.Now().Add(startWait + windowSize)
			w := NewWindowedSchedule(
				interval,
				&start,
				&stop,
				0,
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
		Convey("with a few misses", func() {
			startWait := time.Millisecond * 50
			windowSize := time.Millisecond * 400
			interval := time.Millisecond * 20

			start := time.Now().Add(startWait)
			stop := time.Now().Add(startWait + windowSize)
			w := NewWindowedSchedule(
				interval,
				&start,
				&stop,
				0,
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

			// the task is expected to fire on start time
			So(
				r[0].LastTime().Sub(before).Seconds(),
				ShouldBeBetweenOrEqual,
				(startWait).Seconds(),
				(startWait + interval).Seconds(),
			)
		})
		Convey("started in the past", func() {
			startWait := time.Millisecond * -200
			windowSize := time.Millisecond * 600
			interval := time.Millisecond * 20

			start := time.Now().Add(startWait)
			stop := time.Now().Add(startWait + windowSize)
			w := NewWindowedSchedule(
				interval,
				&start,
				&stop,
				0,
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
				ShouldBeLessThan,
				(interval).Seconds(),
			)
		})
		Convey("start without stop", func() {
			startWait := time.Millisecond * 50
			interval := time.Millisecond * 20

			start := time.Now().Add(startWait)
			w := NewWindowedSchedule(
				interval,
				&start,
				nil,
				0,
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

			// the task is expected to fire immediately on start time
			So(
				r[0].LastTime().Sub(before).Seconds(),
				ShouldBeBetweenOrEqual,
				(startWait).Seconds(),
				(startWait + interval).Seconds(),
			)
		})
		Convey("stop without start", func() {
			windowSize := time.Millisecond * 400
			interval := time.Millisecond * 20

			stop := time.Now().Add(windowSize)
			w := NewWindowedSchedule(
				interval,
				nil,
				&stop,
				0,
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
				ShouldBeLessThan,
				(interval).Seconds(),
			)

			var missed uint
			for _, x := range r {
				missed += x.Missed()
			}
			So(missed, ShouldEqual, 0)
		})
	}) // the end of `Nominal windowed Schedule`

	Convey("Windowed Schedule with determined the count of runs", t, func() {
		interval := time.Second

		Convey("expected to start immediately", func() {
			Convey("single run", func() {
				count := uint(1)
				w := NewWindowedSchedule(interval, nil, nil, count)
				err := w.Validate()
				So(err, ShouldBeNil)

				var r []Response
				var r1 Response
				last := *new(time.Time)

				state := Active
				before := time.Now()
				for state == Active {
					r1 = w.Wait(last)
					last = time.Now()
					state = r1.State()
					// skip the response about ending the task
					if state != Ended {
						r = append(r, r1)
					}
				}
				// for this schedule we expect to get 1 response
				// and 0 missed responses
				So(len(r), ShouldEqual, 1)
				So(r[0].Missed(), ShouldEqual, 0)

				// the task is expected to fire immediately
				So(
					r[0].LastTime().Sub(before).Seconds(),
					ShouldBeLessThan,
					interval.Seconds(),
				)
			})
			Convey("multiply runs", func() {
				count := uint(10)
				w := NewWindowedSchedule(interval, nil, nil, count)

				err := w.Validate()
				So(err, ShouldBeNil)

				Convey("with no misses", func() {
					var r []Response
					var r1 Response
					last := *new(time.Time)

					state := Active
					before := time.Now()
					for state == Active {
						r1 = w.Wait(last)
						last = time.Now()
						state = r1.State()
						// skip the response about ending the task
						if state != Ended {
							r = append(r, r1)
						}
					}
					// for this schedule we expect to get count=10 responses
					// and 0 missed responses
					So(len(r), ShouldEqual, count)
					var missed uint
					for _, x := range r {
						missed += x.Missed()
					}
					So(missed, ShouldEqual, 0)

					// the task is expected to fire immediately
					So(
						r[0].LastTime().Sub(before).Seconds(),
						ShouldBeLessThan,
						interval.Seconds(),
					)
				})
				Convey("with a few misses", func() {
					var r []Response
					var r1 Response
					last := *new(time.Time)

					state := Active
					before := time.Now()
					for state == Active {
						r1 = w.Wait(last)
						last = time.Now()
						state = r1.State()
						// skip the response about ending the task
						if state != Ended {
							r = append(r, r1)
						}
						if len(r) == 3 || len(r) == 7 {
							time.Sleep(w.Interval)
						}
					}
					// for this schedule we expect to get 10 responses minus 2 missed responses
					So(len(r), ShouldEqual, count-2)
					var missed uint
					for _, x := range r {
						missed += x.Missed()
					}
					So(missed, ShouldEqual, 2)

					// the task is expected to fire immediately
					So(
						r[0].LastTime().Sub(before).Seconds(),
						ShouldBeLessThan,
						interval.Seconds(),
					)
				})
			})
		})
		Convey("expected to start on start time", func() {
			startWait := time.Millisecond * 100

			Convey("single run", func() {
				count := uint(1)
				start := time.Now().Add(startWait)
				w := NewWindowedSchedule(interval, &start, nil, count)
				err := w.Validate()
				So(err, ShouldBeNil)

				var r []Response
				var r1 Response
				last := *new(time.Time)

				state := Active
				before := time.Now()
				for state == Active {
					r1 = w.Wait(last)
					last = time.Now()
					state = r1.State()
					// skip the response about ending the task
					if state != Ended {
						r = append(r, r1)
					}
				}
				// for this schedule we expect to get count=1 response
				// and 0 missed responses
				So(len(r), ShouldEqual, count)
				So(r[0].Missed(), ShouldEqual, 0)

				// the task is expected to fire on start timestamp
				So(
					r[0].LastTime().Sub(before).Seconds(),
					ShouldBeBetweenOrEqual,
					(startWait).Seconds(),
					(startWait + interval).Seconds(),
				)
			})
			Convey("multiply runs", func() {
				count := uint(10)
				start := time.Now().Add(startWait)
				w := NewWindowedSchedule(interval, &start, nil, count)

				err := w.Validate()
				So(err, ShouldBeNil)

				var r []Response
				var r1 Response
				last := *new(time.Time)

				state := Active
				before := time.Now()
				for state == Active {
					r1 = w.Wait(last)
					last = time.Now()
					state = r1.State()
					// skip the response about ending the task
					if state != Ended {
						r = append(r, r1)
					}
				}
				// for this schedule we expect to get count=10 responses
				// and 0 missed responses
				So(len(r), ShouldEqual, count)
				var missed uint
				for _, x := range r {
					missed += x.Missed()
				}
				So(missed, ShouldEqual, 0)

				// the task is expected to fire on start time
				So(
					r[0].LastTime().Sub(before).Seconds(),
					ShouldBeBetweenOrEqual,
					(startWait).Seconds(),
					(startWait + interval).Seconds(),
				)
			})
		})
		Convey("started in the past", func() {
			startWait := time.Millisecond * -200
			count := uint(1)
			start := time.Now().Add(startWait)
			w := NewWindowedSchedule(
				interval,
				&start,
				nil,
				count,
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
				if state != Ended {
					r = append(r, r1)
				}
			}

			So(len(r), ShouldEqual, 1)
			So(r[0].Missed(), ShouldEqual, 0)

			// start_time points to the past,
			// so the task is expected to fire immediately
			So(
				r[0].LastTime().Sub(before).Seconds(),
				ShouldBeLessThan,
				(interval).Seconds(),
			)
		})
		Convey("with determined stop", func() {
			startWait := time.Millisecond * 50
			windowSize := time.Millisecond * 200
			count := uint(1)

			start := time.Now().Add(startWait)
			stop := time.Now().Add(startWait + windowSize)
			w := NewWindowedSchedule(
				interval,
				&start,
				&stop,
				count,
			)

			Convey("expected ignoring the count and set as default to 0", func() {
				So(w.Count, ShouldEqual, 0)
				So(w.StartTime.Equal(start), ShouldBeTrue)

				Convey("another params should have a value as provided", func() {
					So(w.StartTime.Equal(start), ShouldBeTrue)
					So(w.StopTime.Equal(stop), ShouldBeTrue)
					So(w.Interval, ShouldEqual, interval)
				})
			})

			err := w.Validate()
			So(err, ShouldBeNil)
		})
	}) // the end of `Window schedule with determined the count of runs`
}
