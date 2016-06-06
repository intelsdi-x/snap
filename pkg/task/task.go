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

package task

import (
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"time"

	"github.com/intelsdi-x/snap/core"
	"github.com/intelsdi-x/snap/pkg/schedule"
	"github.com/intelsdi-x/snap/scheduler/wmap"
)

type TaskCreationRequest struct {
	Name     string            `json:"name"`
	Deadline string            `json:"deadline"`
	Workflow *wmap.WorkflowMap `json:"workflow"`
	Schedule Schedule          `json:"schedule"`
	Start    bool              `json:"start"`
}

type Schedule struct {
	Type           string `json:"type,omitempty"`
	Interval       string `json:"interval,omitempty"`
	StartTimestamp *int64 `json:"start_timestamp,omitempty"`
	StopTimestamp  *int64 `json:"stop_timestamp,omitempty"`
}

type configItem struct {
	Key   string      `json:"key"`
	Value interface{} `json:"value"`
}

type task struct {
	ID                 uint64                  `json:"id"`
	Config             map[string][]configItem `json:"config"`
	Name               string                  `json:"name"`
	Deadline           string                  `json:"deadline"`
	Workflow           wmap.WorkflowMap        `json:"workflow"`
	Schedule           schedule.Schedule       `json:"schedule"`
	CreationTime       time.Time               `json:"creation_timestamp,omitempty"`
	LastRunTime        time.Time               `json:"last_run_timestamp,omitempty"`
	HitCount           uint                    `json:"hit_count,omitempty"`
	MissCount          uint                    `json:"miss_count,omitempty"`
	FailedCount        uint                    `json:"failed_count,omitempty"`
	LastFailureMessage string                  `json:"last_failure_message,omitempty"`
	State              string                  `json:"task_state"`
}

// Function used to create a task according to content (1st parameter)
// . Content can be retrieved from a configuration file or a HTTP REST request body
// . Mode is used to specify if the created task should start right away or not
// . function pointer is responsible for effectively creating and returning the created task
func CreateTaskFromContent(body io.ReadCloser,
	mode *bool,
	fp func(sch schedule.Schedule,
		wfMap *wmap.WorkflowMap,
		startOnCreate bool,
		opts ...core.TaskOption) (core.Task, core.TaskErrors)) (core.Task, error) {

	tr, err := marshalTask(body)
	if err != nil {
		return nil, err
	}

	sch, err := makeSchedule(tr.Schedule)
	if err != nil {
		return nil, err
	}

	var opts []core.TaskOption
	if tr.Deadline != "" {
		dl, err := time.ParseDuration(tr.Deadline)
		if err != nil {
			return nil, err
		}
		opts = append(opts, core.TaskDeadlineDuration(dl))
	}

	if tr.Name != "" {
		opts = append(opts, core.SetTaskName(tr.Name))
	}
	opts = append(opts, core.OptionStopOnFailure(10))

	if mode == nil {
		mode = &tr.Start
	}
	if fp == nil {
		return nil, errors.New("Missing workflow creation routine")
	}
	task, errs := fp(sch, tr.Workflow, *mode, opts...)
	if errs != nil && len(errs.Errors()) != 0 {
		var errMsg string
		for _, e := range errs.Errors() {
			errMsg = errMsg + e.Error() + " -- "
		}
		return nil, errors.New(errMsg[:len(errMsg)-4])
	}
	return task, nil
}

func marshalTask(body io.ReadCloser) (*TaskCreationRequest, error) {
	var tr TaskCreationRequest
	errCode, err := MarshalBody(&tr, body)
	if errCode != 0 && err != nil {
		return nil, err
	}
	return &tr, nil
}

func makeSchedule(s Schedule) (schedule.Schedule, error) {
	switch s.Type {
	case "simple":
		d, err := time.ParseDuration(s.Interval)
		if err != nil {
			return nil, err
		}
		sch := schedule.NewSimpleSchedule(d)

		err = sch.Validate()
		if err != nil {
			return nil, err
		}
		return sch, nil
	case "windowed":
		d, err := time.ParseDuration(s.Interval)
		if err != nil {
			return nil, err
		}

		var start, stop *time.Time
		if s.StartTimestamp != nil {
			t := time.Unix(*s.StartTimestamp, 0)
			start = &t
		}
		if s.StopTimestamp != nil {
			t := time.Unix(*s.StopTimestamp, 0)
			stop = &t
		}
		sch := schedule.NewWindowedSchedule(
			d,
			start,
			stop,
		)

		err = sch.Validate()
		if err != nil {
			return nil, err
		}
		return sch, nil
	case "cron":
		if s.Interval == "" {
			return nil, errors.New("missing cron entry")
		}
		sch := schedule.NewCronSchedule(s.Interval)

		err := sch.Validate()
		if err != nil {
			return nil, err
		}
		return sch, nil
	default:
		return nil, errors.New("unknown schedule type " + s.Type)
	}
}

func MarshalBody(in interface{}, body io.ReadCloser) (int, error) {
	b, err := ioutil.ReadAll(body)
	if err != nil {
		return 500, err
	}
	err = json.Unmarshal(b, in)
	if err != nil {
		return 400, err
	}
	return 0, nil
}
