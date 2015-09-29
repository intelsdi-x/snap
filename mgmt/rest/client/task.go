/*
http://www.apache.org/licenses/LICENSE-2.0.txt


Copyright 2015 Intel Coporation

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
	"time"

	"github.com/intelsdi-x/pulse/core"
	"github.com/intelsdi-x/pulse/mgmt/rest/rbody"
	"github.com/intelsdi-x/pulse/mgmt/rest/request"
	"github.com/intelsdi-x/pulse/scheduler/wmap"
)

type Schedule struct {
	Type      string
	Interval  string
	StartTime *time.Time
	StopTime  *time.Time
}

var (
	disabledErr = errors.New("Enable can only be used for a disabled task")
)

func (c *Client) CreateTask(s *Schedule, wf *wmap.WorkflowMap, name string, startTask bool) *CreateTaskResult {
	t := request.TaskCreationRequest{
		Schedule: request.Schedule{
			Type:     s.Type,
			Interval: s.Interval,
		},
		Workflow: wf,
		Start:    startTask,
	}
	// Add start and/or stop timestamps if they exist
	if s.StartTime != nil {
		u := s.StartTime.Unix()
		t.Schedule.StartTimestamp = &u
	}
	if s.StopTime != nil {
		u := s.StopTime.Unix()
		t.Schedule.StopTimestamp = &u
	}

	if name != "" {
		t.Name = name
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

func (c *Client) WatchTask(id uint) *WatchTasksResult {
	r := &WatchTasksResult{
		EventChan: make(chan *rbody.StreamedTaskEvent),
		DoneChan:  make(chan struct{}),
		killChan:  make(chan struct{}),
	}

	url := fmt.Sprintf("%s/tasks/%v/watch", c.prefix, id)
	resp, err := http.Get(url)

	if err != nil {
		r.Err = err
	}

	go func() {
		select {
		case <-r.DoneChan:
			// We killed so just exit select
		case <-r.killChan:
			// We were killed so close resp to signal to server and exit
			resp.Body.Close()
			close(r.DoneChan)
		}
	}()
	// Start watching
	go func() {
		reader := bufio.NewReader(resp.Body)
		for {
			line, _ := reader.ReadBytes('\n')
			ste := &rbody.StreamedTaskEvent{}
			err := json.Unmarshal(line, ste)
			if err != nil {
				r.Err = err
				resp.Body.Close()
				close(r.DoneChan)
				return
			}
			switch ste.EventType {
			case rbody.TaskWatchTaskDisabled:
				r.EventChan <- ste
				resp.Body.Close()
				close(r.DoneChan)
				return
			case rbody.TaskWatchTaskStopped, rbody.TaskWatchTaskStarted, rbody.TaskWatchMetricEvent:
				r.EventChan <- ste
			}
		}
	}()
	return r
}

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

func (c *Client) GetTask(id uint) *GetTaskResult {
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

func (c *Client) StartTask(id int) *StartTasksResult {
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

func (c *Client) StopTask(id int) *StopTasksResult {
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

func (c *Client) RemoveTask(id int) *RemoveTasksResult {
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

//EnableTask can enable a disabled task by cloning it.
//The newly cloned task is started on the creation.
func (c *Client) EnableTask(id int) *CreateTaskResult {
	r := c.GetTask(uint(id))
	if r.Err != nil {
		return &CreateTaskResult{Err: r.Err}
	}

	if r.State != core.TaskStateLookup[core.TaskDisabled] {
		return &CreateTaskResult{Err: disabledErr}
	}

	sch := &Schedule{Type: r.Schedule.Type, Interval: r.Schedule.Interval}
	ct := c.CreateTask(sch, r.Workflow, r.Name, true)
	if ct.Err != nil {
		return &CreateTaskResult{Err: r.Err}
	}
	return ct
}

type CreateTaskResult struct {
	*rbody.AddScheduledTask
	Err error
}

type WatchTasksResult struct {
	count     int
	Err       error
	EventChan chan *rbody.StreamedTaskEvent
	DoneChan  chan struct{}
	killChan  chan struct{}
}

func (w *WatchTasksResult) Close() {
	close(w.killChan)
	// We do this as a way to ensure the signal gets to the server
	<-w.DoneChan
}

type GetTasksResult struct {
	*rbody.ScheduledTaskListReturned
	Err error
}

type GetTaskResult struct {
	*rbody.ScheduledTaskReturned
	Err error
}

type StartTasksResult struct {
	*rbody.ScheduledTaskStarted
	Err error
}

type StopTasksResult struct {
	*rbody.ScheduledTaskStopped
	Err error
}

type RemoveTasksResult struct {
	*rbody.ScheduledTaskRemoved
	Err error
}
