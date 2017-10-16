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

package core

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/intelsdi-x/snap/core/serror"
	"github.com/intelsdi-x/snap/pkg/schedule"
	"github.com/intelsdi-x/snap/scheduler/wmap"
)

type TaskState int

const (
	TaskDisabled TaskState = iota - 1
	TaskStopped
	TaskSpinning
	TaskFiring
	TaskEnded
	TaskStopping
)

var (
	TaskStateLookup = map[TaskState]string{
		TaskDisabled: "Disabled", // on error, not resumable
		TaskStopped:  "Stopped",  // stopped but resumable
		TaskSpinning: "Running",  // running
		TaskFiring:   "Running",  // running (firing can happen so briefly we don't want to try and render it as a string state)
		TaskEnded:    "Ended",    // ended, but resumable if the schedule is still valid and might fire again
		TaskStopping: "Stopping", // channel has been closed, wait for TaskStopped state
	}
)

type TaskWatcherCloser interface {
	Close() error
}

type TaskWatcherHandler interface {
	CatchCollection([]Metric)
	CatchTaskStarted()
	CatchTaskStopped()
	CatchTaskEnded()
	CatchTaskDisabled(string)
}

func (t TaskState) String() string {
	return TaskStateLookup[t]
}

type Task interface {
	ID() string
	// Status() WorkflowState TODO, switch to string
	State() TaskState
	HitCount() uint
	GetName() string
	SetName(string)
	SetID(string)
	MissedCount() uint
	FailedCount() uint
	LastFailureMessage() string
	LastRunTime() *time.Time
	CreationTime() *time.Time
	DeadlineDuration() time.Duration
	SetDeadlineDuration(time.Duration)
	SetTaskID(id string)
	SetStopOnFailure(int)
	MaxCollectDuration() time.Duration
	SetMaxCollectDuration(time.Duration)
	MaxMetricsBuffer() int64
	SetMaxMetricsBuffer(int64)
	GetStopOnFailure() int
	Option(...TaskOption) TaskOption
	WMap() *wmap.WorkflowMap
	Schedule() schedule.Schedule
}

type TaskOption func(Task) TaskOption

// TaskDeadlineDuration sets the tasks deadline.
// The deadline is the amount of time that can pass before a worker begins
// processing the tasks collect job.
func TaskDeadlineDuration(v time.Duration) TaskOption {
	return func(t Task) TaskOption {
		previous := t.DeadlineDuration()
		t.SetDeadlineDuration(v)
		log.WithFields(log.Fields{
			"_module":                "core",
			"_block":                 "TaskDeadlineDuration",
			"task-id":                t.ID(),
			"task-name":              t.GetName(),
			"task deadline duration": t.DeadlineDuration(),
		}).Debug("Setting deadlineDuration on task")

		return TaskDeadlineDuration(previous)
	}
}

// TaskStopOnFailure sets the tasks stopOnFailure
// The stopOnFailure is the number of consecutive task failures that will
// trigger disabling the task
func OptionStopOnFailure(v int) TaskOption {
	return func(t Task) TaskOption {
		previous := t.GetStopOnFailure()
		t.SetStopOnFailure(v)
		log.WithFields(log.Fields{
			"_module":                   "core",
			"_block":                    "OptionStopOnFailure",
			"task-id":                   t.ID(),
			"task-name":                 t.GetName(),
			"consecutive failure limit": t.GetStopOnFailure(),
		}).Debug("Setting stop-on-failure limit for task")
		return OptionStopOnFailure(previous)
	}
}

// SetTaskName sets the name of the task.
// This is optional.
// If task name is not set, the task name is then defaulted to "Task-<task-id>"
func SetTaskName(name string) TaskOption {
	return func(t Task) TaskOption {
		previous := t.GetName()
		t.SetName(name)
		return SetTaskName(previous)
	}
}

func SetTaskID(id string) TaskOption {
	return func(t Task) TaskOption {
		previous := t.ID()
		t.SetID(id)
		return SetTaskID(previous)
	}
}

func SetMaxMetricsBuffer(b int64) TaskOption {
	return func(t Task) TaskOption {
		previous := t.MaxMetricsBuffer()
		t.SetMaxMetricsBuffer(b)
		return SetMaxMetricsBuffer(previous)
	}
}

func SetMaxCollectDuration(d time.Duration) TaskOption {
	return func(t Task) TaskOption {
		previous := t.MaxCollectDuration()
		t.SetMaxCollectDuration(d)
		return SetMaxCollectDuration(previous)
	}
}

type TaskErrors interface {
	Errors() []serror.SnapError
}

type TaskCreationRequest struct {
	Name               string            `json:"name"`
	Version            int               `json:"version"`
	Deadline           string            `json:"deadline"`
	Workflow           *wmap.WorkflowMap `json:"workflow"`
	Schedule           *Schedule         `json:"schedule"`
	Start              bool              `json:"start"`
	MaxFailures        int               `json:"max-failures"`
	MaxCollectDuration string            `json:"max-collect-duration"`
	MaxMetricsBuffer   int64             `json:"max-metrics-buffer"`
}

func (tr *TaskCreationRequest) UnmarshalJSON(data []byte) error {
	t := make(map[string]json.RawMessage)
	if err := json.Unmarshal(data, &t); err != nil {
		return err
	}
	for k, v := range t {
		switch k {
		case "name":
			if err := json.Unmarshal(v, &(tr.Name)); err != nil {
				return fmt.Errorf("%v (while parsing 'name')", err)
			}
		case "deadline":
			if err := json.Unmarshal(v, &(tr.Deadline)); err != nil {
				return fmt.Errorf("%v (while parsing 'deadline')", err)
			}
		case "workflow":
			if err := json.Unmarshal(v, &(tr.Workflow)); err != nil {
				return err
			}
		case "schedule":
			if err := json.Unmarshal(v, &(tr.Schedule)); err != nil {
				return err
			}
		case "start":
			if err := json.Unmarshal(v, &(tr.Start)); err != nil {
				return fmt.Errorf("%v (while parsing 'start')", err)
			}
		case "max-failures":
			if err := json.Unmarshal(v, &(tr.MaxFailures)); err != nil {
				return fmt.Errorf("%v (while parsing 'max-failures')", err)
			}
		case "version":
			if err := json.Unmarshal(v, &(tr.Version)); err != nil {
				return fmt.Errorf("%v (while parsing 'version')", err)
			}
		case "max-collect-duration":
			if err := json.Unmarshal(v, &(tr.MaxCollectDuration)); err != nil {
				return fmt.Errorf("%v (while parsing 'max-collect-duration')", err)
			}
		case "max-metrics-buffer":
			if err := json.Unmarshal(v, &(tr.MaxMetricsBuffer)); err != nil {
				return fmt.Errorf("%v (while parsing 'max-metrics-buffer')", err)
			}
		default:
			return fmt.Errorf("Unrecognized key '%v' in task creation request", k)
		}
	}
	return nil
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
		opts ...TaskOption) (Task, TaskErrors)) (Task, error) {

	tr, err := createTaskRequest(body)
	if err != nil {
		return nil, err
	}

	if err := validateTaskRequest(tr); err != nil {
		return nil, err
	}

	sch, err := makeSchedule(*tr.Schedule)
	if err != nil {
		return nil, err
	}

	var opts []TaskOption
	if tr.Deadline != "" {
		dl, err := time.ParseDuration(tr.Deadline)
		if err != nil {
			return nil, err
		}
		opts = append(opts, TaskDeadlineDuration(dl))
	}

	if tr.Name != "" {
		opts = append(opts, SetTaskName(tr.Name))
	}

	// if a MaxFailures value is included as part of the task creation request
	if tr.MaxFailures != 0 {
		// then set the appropriate value in the opts
		opts = append(opts, OptionStopOnFailure(tr.MaxFailures))
	}

	if mode == nil {
		mode = &tr.Start
	}

	if tr.MaxMetricsBuffer != 0 {
		opts = append(opts, SetMaxMetricsBuffer(tr.MaxMetricsBuffer))
	}

	if tr.MaxCollectDuration != "" {
		dl, err := time.ParseDuration(tr.MaxCollectDuration)
		if err != nil {
			return nil, err
		}
		opts = append(opts, SetMaxCollectDuration(dl))
	}

	if fp == nil {
		return nil, errors.New("Missing workflow creation routine")
	}
	task, errs := fp(sch, tr.Workflow, *mode, opts...)
	if errs != nil && len(errs.Errors()) != 0 {
		var errMsg string
		for _, e := range errs.Errors() {
			errMsg = errMsg + e.Error() + " -- "

			log.WithFields(log.Fields{
				"_file":     "core/task.go",
				"_function": "CreateTaskFromContent",
				"_error":    e.Error(),
				"_fields":   e.Fields(),
			}).Error("error creating task")
		}

		return nil, errors.New(errMsg[:len(errMsg)-4])
	}
	return task, nil
}

func createTaskRequest(body io.ReadCloser) (*TaskCreationRequest, error) {
	var tr TaskCreationRequest
	errCode, err := UnmarshalBody(&tr, body)
	if errCode != 0 && err != nil {
		return nil, err
	}
	return &tr, nil
}

func UnmarshalBody(in interface{}, body io.ReadCloser) (int, error) {
	b, err := ioutil.ReadAll(body)
	if err != nil {
		return 500, err
	}
	err = json.Unmarshal([]byte(os.ExpandEnv(string(b))), in)
	if err != nil {
		return 400, err
	}
	return 0, nil
}

func validateTaskRequest(tr *TaskCreationRequest) error {
	if tr.Schedule == nil || *tr.Schedule == (Schedule{}) {
		return fmt.Errorf("Task must include a schedule, and the schedule must not be empty")
	}

	if tr.Workflow == nil || *tr.Workflow == (wmap.WorkflowMap{}) {
		return fmt.Errorf("Task must include a workflow, and the workflow must not be empty")
	}
	return nil
}
