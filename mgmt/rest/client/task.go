package client

import (
	"encoding/json"
	"fmt"

	"github.com/intelsdi-x/pulse/mgmt/rest/rbody"
	"github.com/intelsdi-x/pulse/mgmt/rest/request"
	"github.com/intelsdi-x/pulse/scheduler/wmap"
)

type Schedule struct {
	Type     string
	Interval string
}

func (c *Client) CreateTask(s *Schedule, wf *wmap.WorkflowMap, name string) *CreateTaskResult {
	t := request.TaskCreationRequest{
		Schedule: request.Schedule{Type: s.Type, Interval: s.Interval},
		Workflow: wf,
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

func (c *Client) StartTask(id uint64) *StartTasksResult {
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

func (c *Client) StopTask(id uint64) *StopTasksResult {
	resp, err := c.do("PUT", fmt.Sprintf("/tasks/%v/stop", id), ContentTypeJSON)
	if err != nil {
		return &StopTasksResult{Err: err}
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

func (c *Client) RemoveTask(id uint64) *RemoveTasksResult {
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

type CreateTaskResult struct {
	*rbody.AddScheduledTask
	Err error
}

type GetTasksResult struct {
	*rbody.ScheduledTaskListReturned
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
