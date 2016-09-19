// +build medium

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
	"errors"
	"fmt"
	"os"
	"testing"

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
	YAML_FILE  = "../examples/tasks/mock-file.yaml"
	JSON_FILE  = "../examples/tasks/mock-file.json"
	DUMMY_TYPE = "dummy"
)

func koRoutine(sch schedule.Schedule,
	wfMap *wmap.WorkflowMap,
	startOnCreate bool,
	opts ...TaskOption) (Task, TaskErrors) {
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
	opts ...TaskOption) (Task, TaskErrors) {
	return nil, nil
}

func TestUnmarshalBodyTask(t *testing.T) {

	Convey("Non existing file", t, func() {
		file, err := os.Open(DUMMY_FILE)
		So(file, ShouldBeNil)
		So(err.Error(), ShouldEqual, fmt.Sprintf("open %s: no such file or directory", DUMMY_FILE))
		code, err := UnmarshalBody(nil, file)
		So(code, ShouldEqual, 500)
		So(err, ShouldNotBeNil)
		So(err.Error(), ShouldEqual, "invalid argument")
	})

	Convey("Bad JSON file", t, func() {
		var tr TaskReq1
		file, err := os.Open(YAML_FILE)
		So(file, ShouldNotBeNil)
		So(err, ShouldBeNil)
		code, err := UnmarshalBody(&tr, file)
		So(code, ShouldEqual, 400)
		So(err, ShouldNotBeNil)
		So(err.Error(), ShouldEqual, "invalid character '-' in numeric literal")
	})

	Convey("Proper JSON file", t, func() {
		var tr TaskReq1
		file, err := os.Open(JSON_FILE)
		So(file, ShouldNotBeNil)
		So(err, ShouldBeNil)
		code, err := UnmarshalBody(&tr, file)
		So(code, ShouldEqual, 0)
		So(err, ShouldBeNil)
	})
}

func TestCreateTaskRequest(t *testing.T) {

	Convey("Non existing file", t, func() {
		file, err := os.Open(DUMMY_FILE)
		So(file, ShouldBeNil)
		So(err.Error(), ShouldEqual, fmt.Sprintf("open %s: no such file or directory", DUMMY_FILE))
		task, err := createTaskRequest(file)
		So(task, ShouldBeNil)
		So(err, ShouldNotBeNil)
		So(err.Error(), ShouldEqual, "invalid argument")
	})

	Convey("Bad JSON file", t, func() {
		file, err := os.Open(YAML_FILE)
		So(file, ShouldNotBeNil)
		So(err, ShouldBeNil)
		task, err := createTaskRequest(file)
		So(task, ShouldBeNil)
		So(err, ShouldNotBeNil)
		So(err.Error(), ShouldEqual, "invalid character '-' in numeric literal")
	})

	Convey("Proper JSON file", t, func() {
		file, err := os.Open(JSON_FILE)
		So(file, ShouldNotBeNil)
		So(err, ShouldBeNil)
		task, err := createTaskRequest(file)
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
