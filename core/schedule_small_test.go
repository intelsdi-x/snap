// +build small

/*
http://www.apache.org/licenses/LICENSE-2.0.txt


Copyright 2016 Intel Corporation

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

package core

import (
	"fmt"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
)

const (
	DUMMY_TYPE = "dummy"
)

func TestMakeSchedule(t *testing.T) {

	Convey("Bad schedule type", t, func() {
		sched1 := &Schedule{Type: DUMMY_TYPE}
		rsched, err := makeSchedule(*sched1)
		So(rsched, ShouldBeNil)
		So(err, ShouldNotBeNil)
		So(err.Error(), ShouldEqual, fmt.Sprintf("unknown schedule type `%s`", DUMMY_TYPE))
	})

	Convey("Simple schedule with missing interval in configuration", t, func() {
		sched1 := &Schedule{Type: "simple"}
		rsched, err := makeSchedule(*sched1)
		So(rsched, ShouldBeNil)
		So(err, ShouldNotBeNil)
		So(err, ShouldEqual, ErrMissingScheduleInterval)
	})

	Convey("Simple schedule with bad duration", t, func() {
		sched1 := &Schedule{Type: "simple", Interval: "dummy"}
		rsched, err := makeSchedule(*sched1)
		So(rsched, ShouldBeNil)
		So(err, ShouldNotBeNil)
		So(err.Error(), ShouldStartWith, "time: invalid duration ")
	})

	Convey("Simple schedule with invalid duration", t, func() {
		sched1 := &Schedule{Type: "simple", Interval: "-1s"}
		rsched, err := makeSchedule(*sched1)
		So(rsched, ShouldBeNil)
		So(err, ShouldNotBeNil)
		So(err.Error(), ShouldEqual, "Interval must be greater than 0")
	})

	Convey("Simple schedule with proper duration", t, func() {
		sched1 := &Schedule{Type: "simple", Interval: "1s"}
		rsched, err := makeSchedule(*sched1)
		So(err, ShouldBeNil)
		So(rsched, ShouldNotBeNil)
		So(rsched.GetState(), ShouldEqual, 0)
	})

	Convey("Simple schedule with determined count", t, func() {
		sched1 := &Schedule{Type: "simple", Interval: "1s", Count: 1}
		rsched, err := makeSchedule(*sched1)
		So(err, ShouldBeNil)
		So(rsched, ShouldNotBeNil)
		So(rsched.GetState(), ShouldEqual, 0)
	})

	Convey("Windowed schedule with missing interval", t, func() {
		sched1 := &Schedule{Type: "windowed"}
		rsched, err := makeSchedule(*sched1)
		So(rsched, ShouldBeNil)
		So(err, ShouldNotBeNil)
		So(err, ShouldEqual, ErrMissingScheduleInterval)
	})

	Convey("Windowed schedule with bad duration", t, func() {
		now := time.Now()
		sched1 := &Schedule{Type: "windowed", Interval: "dummy", StartTimestamp: &now, StopTimestamp: &now}
		rsched, err := makeSchedule(*sched1)
		So(rsched, ShouldBeNil)
		So(err, ShouldNotBeNil)
		So(err.Error(), ShouldStartWith, "time: invalid duration ")
	})

	Convey("Windowed schedule with invalid duration", t, func() {
		startTime := time.Now().Add(time.Minute)
		stopTime := time.Now().Add(2 * time.Minute)
		sched1 := &Schedule{Type: "windowed", Interval: "-1s", StartTimestamp: &startTime, StopTimestamp: &stopTime}
		rsched, err := makeSchedule(*sched1)
		So(rsched, ShouldBeNil)
		So(err, ShouldNotBeNil)
		So(err.Error(), ShouldEqual, "Interval must be greater than 0")
	})

	Convey("Windowed schedule with determined start_timestamp and count", t, func() {
		startTime := time.Now().Add(time.Minute)
		sched1 := &Schedule{Type: "simple", Interval: "1s", StartTimestamp: &startTime, Count: 1}
		rsched, err := makeSchedule(*sched1)
		So(err, ShouldBeNil)
		So(rsched, ShouldNotBeNil)
		So(rsched.GetState(), ShouldEqual, 0)
	})

	Convey("Windowed schedule without determined start_timestamp", t, func() {
		stopTime := time.Now().Add(time.Second)
		sched1 := &Schedule{Type: "windowed", Interval: "1s", StopTimestamp: &stopTime}
		rsched, err := makeSchedule(*sched1)
		So(err, ShouldBeNil)
		So(rsched, ShouldNotBeNil)
	})

	Convey("Windowed schedule without determined stop_timestamp", t, func() {
		startTime := time.Now().Add(time.Second)
		sched1 := &Schedule{Type: "windowed", Interval: "1s", StartTimestamp: &startTime}
		rsched, err := makeSchedule(*sched1)
		So(err, ShouldBeNil)
		So(rsched, ShouldNotBeNil)
	})

	Convey("Windowed schedule without determined start and stop", t, func() {
		sched1 := &Schedule{Type: "windowed", Interval: "1s"}
		rsched, err := makeSchedule(*sched1)
		So(err, ShouldBeNil)
		So(rsched, ShouldNotBeNil)
	})

	Convey("Windowed schedule with start in the past", t, func() {
		startTime := time.Now().Add(-2 * time.Minute)
		stopTime := time.Now().Add(1 * time.Minute)
		sched1 := &Schedule{Type: "windowed", Interval: "1s",
			StartTimestamp: &startTime, StopTimestamp: &stopTime}
		rsched, err := makeSchedule(*sched1)
		So(err, ShouldBeNil)
		So(rsched, ShouldNotBeNil)
	})

	Convey("Windowed schedule with stop in the past", t, func() {
		startTime := time.Now().Add(time.Minute)
		stopTime := time.Now().Add(-60 * time.Minute)
		sched1 := &Schedule{Type: "windowed", Interval: "1s",
			StartTimestamp: &startTime, StopTimestamp: &stopTime}
		rsched, err := makeSchedule(*sched1)
		So(rsched, ShouldBeNil)
		So(err, ShouldNotBeNil)
		So(err.Error(), ShouldEqual, "Stop time is in the past")
	})

	Convey("Windowed schedule with stop before start", t, func() {
		startTime := time.Now().Add(2 * time.Minute)
		stopTime := time.Now().Add(1 * time.Minute)
		sched1 := &Schedule{Type: "windowed", Interval: "1s",
			StartTimestamp: &startTime, StopTimestamp: &stopTime}
		rsched, err := makeSchedule(*sched1)
		So(rsched, ShouldBeNil)
		So(err, ShouldNotBeNil)
		So(err.Error(), ShouldEqual, "Stop time cannot occur before start time")
	})

	Convey("Cron schedule with missing interval duration", t, func() {
		sched1 := &Schedule{Type: "cron"}
		rsched, err := makeSchedule(*sched1)
		So(rsched, ShouldBeNil)
		So(err, ShouldNotBeNil)
		So(err, ShouldEqual, ErrMissingScheduleInterval)
	})

	Convey("Cron schedule with invalid duration", t, func() {
		sched1 := &Schedule{Type: "cron", Interval: "-1 -2 -3 -4 -5 -6"}
		rsched, err := makeSchedule(*sched1)
		So(rsched, ShouldBeNil)
		So(err, ShouldNotBeNil)
		So(err.Error(), ShouldStartWith, "Failed to parse int from")
	})

	Convey("Cron schedule with too few fields entry", t, func() {
		sched1 := &Schedule{Type: "cron", Interval: "1 2 3"}
		rsched, err := makeSchedule(*sched1)
		So(rsched, ShouldBeNil)
		So(err, ShouldNotBeNil)
		So(err.Error(), ShouldStartWith, "Expected 5 or 6 fields, found ")
	})

	Convey("Cron schedule with 5 fields entry", t, func() {
		sched1 := &Schedule{Type: "cron", Interval: "1 2 3 4 5"}
		rsched, err := makeSchedule(*sched1)
		So(err, ShouldBeNil)
		So(rsched, ShouldNotBeNil)
	})

	Convey("Cron schedule with 6 fields entry", t, func() {
		sched1 := &Schedule{Type: "cron", Interval: "1 2 3 4 5 6"}
		rsched, err := makeSchedule(*sched1)
		So(err, ShouldBeNil)
		So(rsched, ShouldNotBeNil)
	})

	Convey("Cron schedule with too many fields entry", t, func() {
		sched1 := &Schedule{Type: "cron", Interval: "1 2 3 4 5 6 7 8"}
		rsched, err := makeSchedule(*sched1)
		So(rsched, ShouldBeNil)
		So(err, ShouldNotBeNil)
		So(err.Error(), ShouldStartWith, "Expected 5 or 6 fields, found ")
	})
}
