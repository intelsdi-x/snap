/*
http://www.apache.org/licenses/LICENSE-2.0.txt


Copyright 2017 Intel Corporation

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

package v2

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"

	"github.com/intelsdi-x/snap/core"
	"github.com/intelsdi-x/snap/core/serror"
	"github.com/intelsdi-x/snap/pkg/schedule"
	"github.com/intelsdi-x/snap/scheduler/wmap"
	"github.com/julienschmidt/httprouter"
)

// TasksResp returns a list of created tasks.
//
// swagger:response TasksResponse
type TasksResp struct {
	// in: body
	Body struct {
		Tasks Tasks `json:"tasks"`
	}
}

// TaskResp returns the giving task.
//
// swagger:response TaskResponse
type TaskResp struct {
	// in: body
	Task Task `json:"task"`
}

// RemoveTaskError unsuccessful generic response to a failed API call
//
// swagger:response TaskErrorResponse
type RemoveTaskError struct {
	// in: body
	Message string `json:"message"`
}

// TasksResponse returns a list of created tasks.
type TasksResponse struct {
	Tasks Tasks `json:"tasks"`
}

// TaskParam contains task id.
//
// swagger:parameters getTask watchTask updateTaskState removeTask
type TaskParam struct {
	// in: path
	// required: true
	ID string `json:"id"`
}

// TaskPostParams defines task POST and PUT entity.
// swagger:parameters addTask
type TaskPostParams struct {
	// Create or update a task.
	//
	// in: formData
	// required: true
	Task string `json:"task"yaml:"task"`
}

// TaskPutParams defines the type for updating a task.
//
// swagger:parameters updateTaskState
type TaskPutParams struct {
	// Update the state of a task
	//
	// in: formData
	//
	// required: true
	Action string `json:"action"`
}

// Task represents Snap task definition.
type Task struct {
	Version  int    `json:"version,omitempty"`
	ID       string `json:"id,omitempty"`
	Name     string `json:"name,omitempty"`
	Deadline string `json:"deadline,omitempty"`
	// required: true
	Workflow *wmap.WorkflowMap `json:"workflow"`
	// required: true
	Schedule           *core.Schedule `json:"schedule"`
	CreationTimestamp  int64          `json:"creation_timestamp,omitempty"`
	LastRunTimestamp   int64          `json:"last_run_timestamp,omitempty"`
	HitCount           int            `json:"hit_count,omitempty"`
	MissCount          int            `json:"miss_count,omitempty"`
	FailedCount        int            `json:"failed_count,omitempty"`
	LastFailureMessage string         `json:"last_failure_message,omitempty"`
	State              string         `json:"state,omitempty"`
	Href               string         `json:"href,omitempty"`
	Start              bool           `json:"start,omitempty"`
	MaxFailures        int            `json:"max-failures,omitempty"`
}

// Tasks a slice of Task
type Tasks []Task

func (s Tasks) Len() int {
	return len(s)
}

func (s Tasks) Less(i, j int) bool {
	return s[j].CreationTime().After(s[i].CreationTime())
}

func (s Tasks) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

// CreationTime Defines the time a task created.
func (s *Task) CreationTime() time.Time {
	return time.Unix(s.CreationTimestamp, 0)
}

func (s *apiV2) addTask(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	err := addTaskHelper(r)
	if err != nil {
		Write(500, FromError(err), w)
		return
	}

	task, err := core.CreateTaskFromContent(r.Body, nil, s.taskManager.CreateTask)
	if err != nil {
		Write(500, FromError(err), w)
		return
	}

	taskB := AddSchedulerTaskFromTask(task)
	taskB.Href = taskURI(r.Host, task)
	Write(201, taskB, w)
}

func (s *apiV2) getTasks(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	// get tasks from the task manager
	sts := s.taskManager.GetTasks()

	// create the task list response
	tasks := make(Tasks, len(sts))
	i := 0
	for _, t := range sts {
		tasks[i] = SchedulerTaskFromTask(t)
		tasks[i].Href = taskURI(r.Host, t)
		i++
	}
	sort.Sort(tasks)

	Write(200, TasksResponse{Tasks: tasks}, w)
}

func (s *apiV2) getTask(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	id := p.ByName("id")
	t, err := s.taskManager.GetTask(id)
	if err != nil {
		Write(404, FromError(err), w)
		return
	}
	task := AddSchedulerTaskFromTask(t)
	task.Href = taskURI(r.Host, t)
	Write(200, task, w)
}

func (s *apiV2) updateTaskState(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	errs := make([]serror.SnapError, 0, 1)
	id := p.ByName("id")

	action, err := updateTaskStateHelper(r)
	if err != nil {
		errs = append(errs, err)
	} else {
		switch action[0] {
		case "enable":
			_, err := s.taskManager.EnableTask(id)
			if err != nil {
				errs = append(errs, serror.New(err))
			}
		case "start":
			errs = s.taskManager.StartTask(id)
		case "stop":
			errs = s.taskManager.StopTask(id)
		default:
			errs = append(errs, serror.New(ErrWrongAction))
		}
	}

	if len(errs) > 0 {
		statusCode := 500
		switch errs[0].Error() {
		case ErrNoActionSpecified.Error():
			statusCode = 400
		case ErrWrongAction.Error():
			statusCode = 400
		case ErrTaskNotFound:
			statusCode = 404
		case ErrTaskDisabledNotRunnable:
			statusCode = 409
		case ErrReadRequestBody.Error():
			statusCode = 400
		}
		Write(statusCode, FromSnapErrors(errs), w)
		return
	}
	Write(204, nil, w)
}

func (s *apiV2) removeTask(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	id := p.ByName("id")
	err := s.taskManager.RemoveTask(id)
	if err != nil {
		if strings.Contains(err.Error(), ErrTaskNotFound) {
			Write(404, FromError(err), w)
			return
		}
		Write(500, FromError(err), w)
		return
	}
	Write(204, nil, w)
}

func taskURI(host string, t core.Task) string {
	return fmt.Sprintf("%s://%s/%s/tasks/%s", protocolPrefix, host, version, t.ID())
}

// functions to convert a core.Task to a Task
func AddSchedulerTaskFromTask(t core.Task) Task {
	st := SchedulerTaskFromTask(t)
	(&st).assertSchedule(t.Schedule())
	st.Workflow = t.WMap()
	return st
}

func SchedulerTaskFromTask(t core.Task) Task {
	st := Task{
		ID:                 t.ID(),
		Name:               t.GetName(),
		Deadline:           t.DeadlineDuration().String(),
		CreationTimestamp:  t.CreationTime().Unix(),
		LastRunTimestamp:   t.LastRunTime().Unix(),
		HitCount:           int(t.HitCount()),
		MissCount:          int(t.MissedCount()),
		FailedCount:        int(t.FailedCount()),
		LastFailureMessage: t.LastFailureMessage(),
		State:              t.State().String(),
	}
	if st.LastRunTimestamp < 0 {
		st.LastRunTimestamp = -1
	}
	return st
}

func (t *Task) assertSchedule(s schedule.Schedule) {
	switch v := s.(type) {
	case *schedule.WindowedSchedule:
		t.Schedule = &core.Schedule{
			Type:           "windowed",
			Interval:       v.Interval.String(),
			StartTimestamp: v.StartTime,
			StopTimestamp:  v.StopTime,
		}
		return
	case *schedule.CronSchedule:
		t.Schedule = &core.Schedule{
			Type:     "cron",
			Interval: v.Entry(),
		}
		return
	}
}

// addTaskHelper deals with different forms of request data and make it acceeptable by method addTask.
// currently it supports clients of go-swagger, swagger-ui and Snap CLI.
func addTaskHelper(r *http.Request) error {
	buf, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return err
	}

	r.Body = ioutil.NopCloser(bytes.NewBuffer(buf))

	dm := map[string]string{}
	err = json.Unmarshal(buf, &dm)

	sw := true
	// Do not return an error here
	// As it explores different request formats.
	if err != nil {
		sw = false

		// from go-swagger client
		data, _ := url.QueryUnescape(string(buf))
		tokens := strings.Split(data, "=")
		if len(tokens) == 2 {
			dm["task"] = tokens[1]
			sw = true
		}
	}

	if sw {
		r.Body = ioutil.NopCloser(strings.NewReader(dm["task"]))
		r.ContentLength = int64(len(dm["task"]))
	}
	return nil
}

// updateTaskStateHelper deals with different forms of request data and make it acceptable by the method updateTaskState.
// currently it accepts clients of go-swagger, swagger-ui, and Snap CLI.
func updateTaskStateHelper(r *http.Request) ([]string, serror.SnapError) {
	buf, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return nil, serror.New(ErrReadRequestBody)
	}

	dm := map[string]string{}
	err = json.Unmarshal(buf, &dm)

	sw := true
	action := []string{}
	if err != nil {
		sw = false

		// from go-swagger client
		tokens := strings.Split(string(buf), "=")
		if len(tokens) == 2 {
			dm["action"] = tokens[1]
			sw = true
		}
	}

	if sw {
		if strings.Trim(dm["action"], " ") == "" {
			return nil, serror.New(ErrNoActionSpecified)
		}
		action = append(action, dm["action"])
	} else {
		action, exist := r.URL.Query()["action"]
		if !exist {
			return action, serror.New(ErrNoActionSpecified)
		}
	}
	return action, nil
}
