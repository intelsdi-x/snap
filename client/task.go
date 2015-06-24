package pulse

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

// CreateTask takes a pointer to a task structure,
// and POSTs it to Pulse's REST API.
// If an error is encountered during the process it returns
// a generic error.  If an error occured when Pulse attempts
// to create the Task, a type-assertable error is returned.
// Also note that CreateTask modifies the pointed to data
// by adding an ID and a created time.
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
		return &CreateTaskResult{Error: err}
	}

	resp, err := c.do("POST", "/tasks", j)
	if err != nil {
		return &CreateTaskResult{Error: err}
	}

	switch resp.Meta.Type {
	case rbody.AddScheduledTaskType:
		// Success
		return &CreateTaskResult{resp.Body.(*rbody.AddScheduledTask), nil}
	case rbody.ErrorType:
		return &CreateTaskResult{Error: resp.Body.(*rbody.Error)}
	default:
		return &CreateTaskResult{Error: ErrAPIResponseMetaType}
	}
}

func (c *Client) GetTasks() *GetTasksResult {
	resp, err := c.do("GET", "/tasks", nil)
	if err != nil {
		return &GetTasksResult{Error: err}
	}

	switch resp.Meta.Type {
	case rbody.ScheduledTaskListReturnedType:
		// Success
		return &GetTasksResult{resp.Body.(*rbody.ScheduledTaskListReturned), nil}
	case rbody.ErrorType:
		return &GetTasksResult{Error: resp.Body.(*rbody.Error)}
	default:
		return &GetTasksResult{Error: ErrAPIResponseMetaType}
	}
}

func (c *Client) StartTask(id uint64) *StartTasksResult {
	resp, err := c.do("PUT", fmt.Sprintf("/tasks/%v/start", id))

	if err != nil {
		return &StartTasksResult{Error: err}
	}

	switch resp.Meta.Type {
	case rbody.ScheduledTaskStartedType:
		// Success
		return &StartTasksResult{resp.Body.(*rbody.ScheduledTaskStarted), nil}
	case rbody.ErrorType:
		return &StartTasksResult{Error: resp.Body.(*rbody.Error)}
	default:
		return &StartTasksResult{Error: ErrAPIResponseMetaType}
	}
	return nil
}

func (c *Client) StopTask(id uint64) *StopTasksResult {
	resp, err := c.do("PUT", fmt.Sprintf("/tasks/%v/stop", id))
	if err != nil {
		return &StopTasksResult{Error: err}
	}

	switch resp.Meta.Type {
	case rbody.ScheduledTaskStoppedType:
		// Success
		return &StopTasksResult{resp.Body.(*rbody.ScheduledTaskStopped), nil}
	case rbody.ErrorType:
		return &StopTasksResult{Error: resp.Body.(*rbody.Error)}
	default:
		return &StopTasksResult{Error: ErrAPIResponseMetaType}
	}
	return nil
}

func (c *Client) RemoveTask(id uint64) *RemoveTasksResult {
	resp, err := c.do("DELETE", fmt.Sprintf("/tasks/%v", id))
	if err != nil {
		return &RemoveTasksResult{Error: err}
	}

	switch resp.Meta.Type {
	case rbody.ScheduledTaskRemovedType:
		// Success
		return &RemoveTasksResult{resp.Body.(*rbody.ScheduledTaskRemoved), nil}
	case rbody.ErrorType:
		return &RemoveTasksResult{Error: resp.Body.(*rbody.Error)}
	default:
		return &RemoveTasksResult{Error: ErrAPIResponseMetaType}
	}

	return nil
}

type CreateTaskResult struct {
	*rbody.AddScheduledTask
	Error error
}

type GetTasksResult struct {
	*rbody.ScheduledTaskListReturned
	Error error
}

type StartTasksResult struct {
	*rbody.ScheduledTaskStarted
	Error error
}

type StopTasksResult struct {
	*rbody.ScheduledTaskStopped
	Error error
}

type RemoveTasksResult struct {
	*rbody.ScheduledTaskRemoved
	Error error
}
