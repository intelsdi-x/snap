/*
http://www.apache.org/licenses/LICENSE-2.0.txt


Copyright 2015,2016 Intel Corporation

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

package client

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/intelsdi-x/snap/core"
	"github.com/intelsdi-x/snap/mgmt/rest/v1/rbody"
	"github.com/intelsdi-x/snap/scheduler/wmap"
)

type Schedule struct {
	// Type specifies the type of the schedule. Currently, the type of "simple", "windowed" and "cron" are supported.
	Type string `json:"type,omitempty"`
	// Interval specifies the time duration.
	Interval string `json:"interval,omitempty"`
	// StartTimestamp specifies the beginning time.
	StartTimestamp *time.Time `json:"start_timestamp,omitempty"`
	// StopTimestamp specifies the end time.
	StopTimestamp *time.Time `json:"stop_timestamp,omitempty"`
	// Count specifies the number of expected runs (defaults to 0 what means no limit, set to 1 means single run task).
	// Count is supported by "simple" and "windowed" schedules
	Count uint `json:"count,omitempty"`
}

// CreateTask creates a task given the schedule, workflow, task name, and task state.
// If the startTask flag is true, the newly created task is started after the creation.
// Otherwise, it's in the Stopped state. CreateTask is accomplished through a POST HTTP JSON request.
// A ScheduledTask is returned if it succeeds, otherwise an error is returned.
func (c *Client) CreateTask(s *Schedule, wf *wmap.WorkflowMap, name string, deadline string, startTask bool, maxFailures int) *CreateTaskResult {
	t := core.TaskCreationRequest{
		Schedule: &core.Schedule{
			Type:           s.Type,
			Interval:       s.Interval,
			StartTimestamp: s.StartTimestamp,
			StopTimestamp:  s.StopTimestamp,
			Count:          s.Count,
		},
		Workflow:    wf,
		Start:       startTask,
		MaxFailures: maxFailures,
	}
	if name != "" {
		t.Name = name
	}
	if deadline != "" {
		t.Deadline = deadline
	}
	// Marshal to JSON for request body
	j, err := json.Marshal(t)
	if err != nil {
		return &CreateTaskResult{Err: err}
	}

	resp, err := c.do("POST", "/tasks", ContentTypeJSON, j)
	if err != nil {
		return &CreateTaskResult{Err: err}
	}

	switch resp.Meta.Type {
	case rbody.AddScheduledTaskType:
		// Success
		return &CreateTaskResult{resp.Body.(*rbody.AddScheduledTask), nil}
	case rbody.ErrorType:
		return &CreateTaskResult{Err: resp.Body.(*rbody.Error)}
	default:
		return &CreateTaskResult{Err: ErrAPIResponseMetaType}
	}
}

// WatchTask retrieves running tasks by running a goroutine to
// interactive with Event and Done channels. An HTTP GET request retrieves tasks.
// StreamedTaskEvent returns if it succeeds. Otherwise, an error is returned.
func (c *Client) WatchTask(id string) *WatchTasksResult {
	// during watch we don't want to have a timeout
	// Store the old timeout so we can restore when we are through
	oldTimeout := c.http.Timeout
	c.http.Timeout = time.Duration(0)

	r := &WatchTasksResult{
		EventChan: make(chan *rbody.StreamedTaskEvent),
		DoneChan:  make(chan struct{}),
	}

	url := fmt.Sprintf("%s/tasks/%v/watch", c.prefix, id)
	req, err := http.NewRequest("GET", url, nil)
	addAuth(req, c.Username, c.Password)
	if err != nil {
		r.Err = err
		r.Close()
		return r
	}
	resp, err := c.http.Do(req)
	if err != nil {
		if strings.Contains(err.Error(), "tls: oversized record") || strings.Contains(err.Error(), "malformed HTTP response") {
			r.Err = fmt.Errorf("error connecting to API URI: %s. Do you have an http/https mismatch?", c.URL)
		} else {
			r.Err = err
		}
		r.Close()
		return r
	}

	if resp.StatusCode != 200 {
		ar, err := httpRespToAPIResp(resp)
		if err != nil {
			r.Err = err
		} else {
			r.Err = errors.New(ar.Meta.Message)
		}
		r.Close()
		return r
	}

	// Start watching
	go func() {
		reader := bufio.NewReader(resp.Body)
		defer func() { c.http.Timeout = oldTimeout }()
		for {
			select {
			case <-r.DoneChan:
				resp.Body.Close()
				return
			default:
				line, _ := reader.ReadBytes('\n')
				sline := string(line)
				if sline == "" || sline == "\n" {
					continue
				}
				if strings.HasPrefix(sline, "data:") {
					sline = strings.TrimPrefix(sline, "data:")
					line = []byte(sline)
				}
				ste := &rbody.StreamedTaskEvent{}
				err := json.Unmarshal(line, ste)
				if err != nil {
					r.Err = err
					r.Close()
					return
				}
				switch ste.EventType {
				case rbody.TaskWatchTaskDisabled:
					r.EventChan <- ste
					r.Close()
				case rbody.TaskWatchTaskStopped, rbody.TaskWatchTaskEnded, rbody.TaskWatchTaskStarted, rbody.TaskWatchMetricEvent:
					r.EventChan <- ste
				}
			}
		}
	}()
	return r
}

// GetTasks retrieves a slice of tasks through an HTTP GET call.
// A list of scheduled tasks returns if it succeeds.
// Otherwise. an error is returned.
func (c *Client) GetTasks() *GetTasksResult {
	resp, err := c.do("GET", "/tasks", ContentTypeJSON, nil)
	if err != nil {
		return &GetTasksResult{Err: err}
	}

	switch resp.Meta.Type {
	case rbody.ScheduledTaskListReturnedType:
		// Success
		return &GetTasksResult{resp.Body.(*rbody.ScheduledTaskListReturned), nil}
	case rbody.ErrorType:
		return &GetTasksResult{Err: resp.Body.(*rbody.Error)}
	default:
		return &GetTasksResult{Err: ErrAPIResponseMetaType}
	}
}

// GetTask retrieves the task given a task id through an HTTP GET call.
// A scheduled task returns if it succeeds. Otherwise, an error is returned.
func (c *Client) GetTask(id string) *GetTaskResult {
	resp, err := c.do("GET", fmt.Sprintf("/tasks/%v", id), ContentTypeJSON, nil)
	if err != nil {
		return &GetTaskResult{Err: err}
	}
	switch resp.Meta.Type {
	case rbody.ScheduledTaskReturnedType:
		// Success
		return &GetTaskResult{resp.Body.(*rbody.ScheduledTaskReturned), nil}
	case rbody.ErrorType:
		return &GetTaskResult{Err: resp.Body.(*rbody.Error)}
	default:
		return &GetTaskResult{Err: ErrAPIResponseMetaType}
	}
}

// StartTask starts a task given a task id. The scheduled task will be in
// the started state if it succeeds. Otherwise, an error is returned.
func (c *Client) StartTask(id string) *StartTasksResult {
	resp, err := c.do("PUT", fmt.Sprintf("/tasks/%v/start", id), ContentTypeJSON)

	if err != nil {
		return &StartTasksResult{Err: err}
	}

	switch resp.Meta.Type {
	case rbody.ScheduledTaskStartedType:
		// Success
		return &StartTasksResult{resp.Body.(*rbody.ScheduledTaskStarted), nil}
	case rbody.ErrorType:
		return &StartTasksResult{Err: resp.Body.(*rbody.Error)}
	default:
		return &StartTasksResult{Err: ErrAPIResponseMetaType}
	}
}

// StopTask stops a running task given a task id. It uses an HTTP PUT call.
// The stopped task id returns if it succeeds. Otherwise, an error is returned.
func (c *Client) StopTask(id string) *StopTasksResult {
	resp, err := c.do("PUT", fmt.Sprintf("/tasks/%v/stop", id), ContentTypeJSON)
	if err != nil {
		return &StopTasksResult{Err: err}
	}

	if resp == nil {
		return nil
	}
	switch resp.Meta.Type {
	case rbody.ScheduledTaskStoppedType:
		// Success
		return &StopTasksResult{resp.Body.(*rbody.ScheduledTaskStopped), nil}
	case rbody.ErrorType:
		return &StopTasksResult{Err: resp.Body.(*rbody.Error)}
	default:
		return &StopTasksResult{Err: ErrAPIResponseMetaType}
	}
}

// RemoveTask removes a task from the schedule tasks given a task id. It's through an HTTP DELETE call.
// The removed task id returns if it succeeds. Otherwise, an error is returned.
func (c *Client) RemoveTask(id string) *RemoveTasksResult {
	resp, err := c.do("DELETE", fmt.Sprintf("/tasks/%v", id), ContentTypeJSON)
	if err != nil {
		return &RemoveTasksResult{Err: err}
	}

	switch resp.Meta.Type {
	case rbody.ScheduledTaskRemovedType:
		// Success
		return &RemoveTasksResult{resp.Body.(*rbody.ScheduledTaskRemoved), nil}
	case rbody.ErrorType:
		return &RemoveTasksResult{Err: resp.Body.(*rbody.Error)}
	default:
		return &RemoveTasksResult{Err: ErrAPIResponseMetaType}
	}
}

// EnableTask enables a disabled task given a task id. The request is an HTTP PUT call.
// The enabled task id returns if it succeeds. Otherwise, an error is returned.
func (c *Client) EnableTask(id string) *EnableTaskResult {
	resp, err := c.do("PUT", fmt.Sprintf("/tasks/%v/enable", id), ContentTypeJSON)
	if err != nil {
		return &EnableTaskResult{Err: err}
	}

	switch resp.Meta.Type {
	case rbody.ScheduledTaskEnabledType:
		return &EnableTaskResult{resp.Body.(*rbody.ScheduledTaskEnabled), nil}
	case rbody.ErrorType:
		return &EnableTaskResult{Err: resp.Body.(*rbody.Error)}
	default:
		return &EnableTaskResult{Err: ErrAPIResponseMetaType}
	}
}

// CreateTaskResult is the response from snap/client on a CreateTask call.
type CreateTaskResult struct {
	*rbody.AddScheduledTask
	Err error
}

// WatchTaskResult is the response from snap/client on a WatchTask call.
type WatchTasksResult struct {
	count     int
	Err       error
	EventChan chan *rbody.StreamedTaskEvent
	DoneChan  chan struct{}
}

func (w *WatchTasksResult) Close() {
	close(w.DoneChan)
}

// GetTasksResult is the response from snap/client on a GetTasks call.
type GetTasksResult struct {
	*rbody.ScheduledTaskListReturned
	Err error
}

// GetTaskResult is the response from snap/client on a GetTask call.
type GetTaskResult struct {
	*rbody.ScheduledTaskReturned
	Err error
}

// StartTasksResult is the response from snap/client on a StartTask call.
type StartTasksResult struct {
	*rbody.ScheduledTaskStarted
	Err error
}

// StopTasksResult is the response from snap/client on a StopTask call.
type StopTasksResult struct {
	*rbody.ScheduledTaskStopped
	Err error
}

// RemoveTasksResult is the response from snap/client on a RemoveTask call.
type RemoveTasksResult struct {
	*rbody.ScheduledTaskRemoved
	Err error
}

// EnableTasksResult is the response from snap/client on a EnableTask call.
type EnableTaskResult struct {
	*rbody.ScheduledTaskEnabled
	Err error
}
