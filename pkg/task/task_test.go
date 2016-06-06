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

package task

import (
	"errors"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/intelsdi-x/snap/core"
	"github.com/intelsdi-x/snap/core/serror"
	"github.com/intelsdi-x/snap/pkg/schedule"
	"github.com/intelsdi-x/snap/scheduler/wmap"
	. "github.com/smartystreets/goconvey/convey"
)

type TaskReq1 struct {
	Name  string `json:"name"`
	Start bool   `json:"start"`
}

type taskErrors struct {
	errs []serror.SnapError
}

func (t *taskErrors) Errors() []serror.SnapError {
	return t.errs
}

const (
	DUMMY_FILE = "dummy.txt"
	YAML_FILE  = "../../examples/tasks/mock-file.yaml"
	JSON_FILE  = "../../examples/tasks/mock-file.json"
	DUMMY_TYPE = "dummy"
)

func koRoutine(sch schedule.Schedule,
	wfMap *wmap.WorkflowMap,
	startOnCreate bool,
	opts ...core.TaskOption) (core.Task, core.TaskErrors) {
	// Create a container for task errors
	te := &taskErrors{
		errs: make([]serror.SnapError, 0),
	}
	te.errs = append(te.errs, serror.New(errors.New("Dummy error")))
	return nil, te
}

func okRoutine(sch schedule.Schedule,
	wfMap *wmap.WorkflowMap,
	startOnCreate bool,
	opts ...core.TaskOption) (core.Task, core.TaskErrors) {
	return nil, nil
}

func TestMarshalBodyTask(t *testing.T) {

	Convey("Non existing file", t, func() {
		file, err := os.Open(DUMMY_FILE)
		So(file, ShouldBeNil)
		So(err.Error(), ShouldEqual, fmt.Sprintf("open %s: no such file or directory", DUMMY_FILE))
		code, err := MarshalBody(nil, file)
		So(code, ShouldEqual, 500)
		So(err, ShouldNotBeNil)
		So(err.Error(), ShouldEqual, "invalid argument")
	})

	Convey("Bad JSON file", t, func() {
		var tr TaskReq1
		file, err := os.Open(YAML_FILE)
		So(file, ShouldNotBeNil)
		So(err, ShouldBeNil)
		code, err := MarshalBody(&tr, file)
		So(code, ShouldEqual, 400)
		So(err, ShouldNotBeNil)
		So(err.Error(), ShouldEqual, "invalid character '-' in numeric literal")
	})

	Convey("Proper JSON file", t, func() {
		var tr TaskReq1
		file, err := os.Open(JSON_FILE)
		So(file, ShouldNotBeNil)
		So(err, ShouldBeNil)
		code, err := MarshalBody(&tr, file)
		So(code, ShouldEqual, 0)
		So(err, ShouldBeNil)
	})
}

func TestMarshalTask(t *testing.T) {

	Convey("Non existing file", t, func() {
		file, err := os.Open(DUMMY_FILE)
		So(file, ShouldBeNil)
		So(err.Error(), ShouldEqual, fmt.Sprintf("open %s: no such file or directory", DUMMY_FILE))
		task, err := marshalTask(file)
		So(task, ShouldBeNil)
		So(err, ShouldNotBeNil)
		So(err.Error(), ShouldEqual, "invalid argument")
	})

	Convey("Bad JSON file", t, func() {
		file, err := os.Open(YAML_FILE)
		So(file, ShouldNotBeNil)
		So(err, ShouldBeNil)
		task, err := marshalTask(file)
		So(task, ShouldBeNil)
		So(err, ShouldNotBeNil)
		So(err.Error(), ShouldEqual, "invalid character '-' in numeric literal")
	})

	Convey("Proper JSON file", t, func() {
		file, err := os.Open(JSON_FILE)
		So(file, ShouldNotBeNil)
		So(err, ShouldBeNil)
		task, err := marshalTask(file)
		So(err, ShouldBeNil)
		So(task, ShouldNotBeNil)
		So(task.Name, ShouldEqual, "")
		So(task.Deadline, ShouldEqual, "")
		So(task.Schedule.Type, ShouldEqual, "simple")
		So(task.Schedule.Interval, ShouldEqual, "1s")
		So(task.Schedule.StartTimestamp, ShouldBeNil)
		So(task.Schedule.StopTimestamp, ShouldBeNil)
		So(task.Start, ShouldEqual, false)
	})
}

func TestMakeSchedule(t *testing.T) {

	Convey("Bad schedule type", t, func() {
		sched1 := &Schedule{Type: DUMMY_TYPE}
		rsched, err := makeSchedule(*sched1)
		So(rsched, ShouldBeNil)
		So(err, ShouldNotBeNil)
		So(err.Error(), ShouldEqual, fmt.Sprintf("unknown schedule type %s", DUMMY_TYPE))
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

	Convey("Windowed schedule with bad duration", t, func() {
		sched1 := &Schedule{Type: "windowed", Interval: "dummy"}
		rsched, err := makeSchedule(*sched1)
		So(rsched, ShouldBeNil)
		So(err, ShouldNotBeNil)
		So(err.Error(), ShouldStartWith, "time: invalid duration ")
	})

	Convey("Windowed schedule with invalid duration", t, func() {
		sched1 := &Schedule{Type: "windowed", Interval: "-1s"}
		rsched, err := makeSchedule(*sched1)
		So(rsched, ShouldBeNil)
		So(err, ShouldNotBeNil)
		So(err.Error(), ShouldEqual, "Interval must be greater than 0")
	})

	Convey("Windowed schedule with stop in the past", t, func() {
		now := time.Now()
		startSecs := now.Unix()
		stopSecs := startSecs - 3600
		sched1 := &Schedule{Type: "windowed", Interval: "1s",
			StartTimestamp: &startSecs, StopTimestamp: &stopSecs}
		rsched, err := makeSchedule(*sched1)
		So(rsched, ShouldBeNil)
		So(err, ShouldNotBeNil)
		So(err.Error(), ShouldEqual, "Stop time is in the past")
	})

	Convey("Windowed schedule with stop before start", t, func() {
		now := time.Now()
		startSecs := now.Unix()
		stopSecs := startSecs + 600
		startSecs = stopSecs + 600
		sched1 := &Schedule{Type: "windowed", Interval: "1s",
			StartTimestamp: &startSecs, StopTimestamp: &stopSecs}
		rsched, err := makeSchedule(*sched1)
		So(rsched, ShouldBeNil)
		So(err, ShouldNotBeNil)
		So(err.Error(), ShouldEqual, "Stop time cannot occur before start time")
	})

	Convey("Windowed schedule with stop before start", t, func() {
		now := time.Now()
		startSecs := now.Unix()
		stopSecs := startSecs + 600
		sched1 := &Schedule{Type: "windowed", Interval: "1s",
			StartTimestamp: &startSecs, StopTimestamp: &stopSecs}
		rsched, err := makeSchedule(*sched1)
		So(err, ShouldBeNil)
		So(rsched, ShouldNotBeNil)
		So(rsched.GetState(), ShouldEqual, 0)
	})

	Convey("Cron schedule with bad duration", t, func() {
		sched1 := &Schedule{Type: "cron", Interval: ""}
		rsched, err := makeSchedule(*sched1)
		So(rsched, ShouldBeNil)
		So(err, ShouldNotBeNil)
		So(err.Error(), ShouldEqual, "missing cron entry")
	})

	Convey("Cron schedule with invalid duration", t, func() {
		sched1 := &Schedule{Type: "windowed", Interval: "-1s"}
		rsched, err := makeSchedule(*sched1)
		So(rsched, ShouldBeNil)
		So(err, ShouldNotBeNil)
		So(err.Error(), ShouldEqual, "Interval must be greater than 0")
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

func TestCreateTaskFromContent(t *testing.T) {

	Convey("Non existing file", t, func() {
		file, err := os.Open(DUMMY_FILE)
		So(file, ShouldBeNil)
		So(err.Error(), ShouldEqual, fmt.Sprintf("open %s: no such file or directory", DUMMY_FILE))
		autoStart := true
		task, err := CreateTaskFromContent(file, &autoStart, nil)
		So(task, ShouldBeNil)
		So(err, ShouldNotBeNil)
		So(err.Error(), ShouldEqual, "invalid argument")
	})

	Convey("Bad JSON file", t, func() {
		file, err := os.Open(YAML_FILE)
		So(file, ShouldNotBeNil)
		So(err, ShouldBeNil)
		autoStart := true
		task, err := CreateTaskFromContent(file, &autoStart, nil)
		So(task, ShouldBeNil)
		So(err, ShouldNotBeNil)
		So(err.Error(), ShouldEqual, "invalid character '-' in numeric literal")
	})

	Convey("Proper JSON file no workflow routine", t, func() {
		file, err := os.Open(JSON_FILE)
		So(file, ShouldNotBeNil)
		So(err, ShouldBeNil)
		autoStart := true
		task, err := CreateTaskFromContent(file, &autoStart, nil)
		So(task, ShouldBeNil)
		So(err, ShouldNotBeNil)
		So(err.Error(), ShouldEqual, "Missing workflow creation routine")
	})

	Convey("Proper JSON file erroring routine", t, func() {
		file, err := os.Open(JSON_FILE)
		So(file, ShouldNotBeNil)
		So(err, ShouldBeNil)
		autoStart := true
		task, err := CreateTaskFromContent(file, &autoStart, koRoutine)
		So(task, ShouldBeNil)
		So(err, ShouldNotBeNil)
		So(err.Error(), ShouldEqual, "Dummy error")
	})

	Convey("Proper JSON file proper routine", t, func() {
		file, err := os.Open(JSON_FILE)
		So(file, ShouldNotBeNil)
		So(err, ShouldBeNil)
		autoStart := true
		task, err := CreateTaskFromContent(file, &autoStart, okRoutine)
		So(task, ShouldBeNil)
		So(err, ShouldBeNil)
	})
}
