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
	"fmt"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/intelsdi-x/snap/core"
	"github.com/intelsdi-x/snap/core/serror"
	"github.com/intelsdi-x/snap/pkg/schedule"
	"github.com/intelsdi-x/snap/scheduler/wmap"
	"github.com/julienschmidt/httprouter"
)

// TasksResponse returns a list of created tasks.
//
// swagger:response TasksResponse
type TasksResp struct {
	// in: body
	Body struct {
		Tasks Tasks `json:"tasks"`
	}
}

// TaskResponse returns a task.
//
// swagger:response TaskResponse
type TaskResp struct {
	// in: body
	Task Task `json:"task"`
}

// TaskErrorResponse returns removing a task error.
//
// swagger:response TaskErrorResponse
type RemoveTaskError struct {
	// in: body
	Message string `json:"message"`
}

type TasksResponse struct {
	Tasks Tasks `json:"tasks"`
}

// TaskParam defines the API path task id.
//
// swagger:parameters getTask watchTask updateTaskState removeTask
type TaskParam struct {
	// in: path
	// required: true
	ID string `json:"id"`
}

// TaskPostParams defines task POST and PUT string representation content.
//
// swagger:parameters addTask
type TaskPostParams struct {
	// Create a task.
	//
	// in: body
	//
	// required: true
	Task Task `json:"task"yaml:"task"`
}

// TaskPutParams defines a task state
//
// swagger:parameters updateTaskState
type TaskPutParams struct {
	// Update the state of a task
	//
	// in: query
	//
	// required: true
	Action string `json:"action"`
}

// Task represents Snap task definition.
type Task struct {
	ID                 string            `json:"id,omitempty"`
	Name               string            `json:"name,omitempty"`
	Version            int               `json:"version,omitempty"`
	Deadline           string            `json:"deadline,omitempty"`
	Workflow           *wmap.WorkflowMap `json:"workflow,omitempty"`
	Schedule           *core.Schedule    `json:"schedule,omitempty"`
	CreationTimestamp  int64             `json:"creation_timestamp,omitempty"`
	LastRunTimestamp   int64             `json:"last_run_timestamp,omitempty"`
	HitCount           int               `json:"hit_count,omitempty"`
	MissCount          int               `json:"miss_count,omitempty"`
	FailedCount        int               `json:"failed_count,omitempty"`
	LastFailureMessage string            `json:"last_failure_message,omitempty"`
	TaskState          string            `json:"task_state,omitempty"`
	Href               string            `json:"href,omitempty"`
	Start              bool              `json:"start,omitempty"`
	MaxFailures        int               `json:"max-failures,omitempty"`
}

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

func (s *Task) CreationTime() time.Time {
	return time.Unix(s.CreationTimestamp, 0)
}

func (s *apiV2) addTask(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
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
	action, exist := r.URL.Query()["action"]
	if !exist && len(action) > 0 {
		errs = append(errs, serror.New(ErrNoActionSpecified))
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
		TaskState:          t.State().String(),
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
