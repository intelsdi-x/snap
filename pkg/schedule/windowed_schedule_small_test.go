// +build small

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

	log "github.com/Sirupsen/logrus"

	. "github.com/smartystreets/goconvey/convey"
)

func TestWindowedScheduleValidation(t *testing.T) {
	log.SetLevel(log.DebugLevel)
	Convey("invalid an interval", t, func() {
		Convey("zero value", func() {
			interval := time.Millisecond * 0
			w := NewWindowedSchedule(interval, nil, nil, 0)
			err := w.Validate()
			So(err, ShouldEqual, ErrInvalidInterval)
		})
		Convey("negative value", func() {
			interval := time.Millisecond * -1
			w := NewWindowedSchedule(interval, nil, nil, 0)
			err := w.Validate()
			So(err, ShouldEqual, ErrInvalidInterval)
		})
	})
	Convey("start time in past is ok (as long as window ends in the future)", t, func() {
		start := time.Now().Add(time.Second * -10)
		stop := time.Now().Add(time.Second * 10)
		w := NewWindowedSchedule(time.Millisecond*100, &start, &stop, 0)
		err := w.Validate()
		So(err, ShouldEqual, nil)
	})
	Convey("window in past", t, func() {
		start := time.Now().Add(time.Second * -20)
		stop := time.Now().Add(time.Second * -10)
		w := NewWindowedSchedule(time.Millisecond*100, &start, &stop, 0)
		err := w.Validate()
		So(err, ShouldEqual, ErrInvalidStopTime)
	})
	Convey("cart before the horse", t, func() {
		start := time.Now().Add(time.Second * 100)
		stop := time.Now().Add(time.Second * 10)
		w := NewWindowedSchedule(time.Millisecond*100, &start, &stop, 0)
		err := w.Validate()
		So(err, ShouldEqual, ErrStopBeforeStart)
	})
	Convey("test Wait()", t, func() {
		interval := 100
		overage := 467
		shouldWait := float64(500 - overage)
		last := time.Now()

		time.Sleep(time.Millisecond * time.Duration(overage))
		s := NewWindowedSchedule(time.Millisecond*time.Duration(interval), nil, nil, 0)
		err := s.Validate()
		So(err, ShouldBeNil)

		before := time.Now()
		r := s.Wait(last)
		after := time.Since(before)

		So(r.Error(), ShouldEqual, nil)
		So(r.State(), ShouldEqual, Active)
		So(r.Missed(), ShouldEqual, 4)
		// We are ok at this precision with being within 10% over or under (10ms)
		afterMS := after.Nanoseconds() / 1000 / 1000
		So(afterMS, ShouldBeGreaterThan, shouldWait-10)
		So(afterMS, ShouldBeLessThan, shouldWait+10)
	})
}
