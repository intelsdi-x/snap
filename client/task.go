package pulse

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/intelsdi-x/pulse/scheduler/wmap"
)

type ErrTaskCreationFailed struct {
	message string
}

func (e *ErrTaskCreationFailed) Error() string {
	return e.message
}

type ErrGetTasksFailed struct {
	message string
}

func (e *ErrGetTasksFailed) Error() string {
	return e.message
}

type ErrStartTask struct {
	message string
}

func (e *ErrStartTask) Error() string {
	return e.message
}

// type TaskSchedule struct {
// 	Interval time.Duration `json:"interval"`
// }

// type ConfigSetting struct {
// 	Namespace []string
// 	Key       string
// 	Value     interface{}
// }

type Schedule struct {
	Type     string `json:"type"`
	Interval string `json:"interval"`
}

type Task struct {

	// A UUID to identify the task. This is set by the Pulse Agent.
	// It is strongly advised not to make changes to this field once
	// the agent has set it.
	ID uint64 `json:"id"`

	Workflow *wmap.WorkflowMap `json:"workflow"`
	// Config       []ConfigSetting   `json:"config"`
	Schedule     *Schedule  `json:"schedule"`
	CreationTime *time.Time `json:"creation_timestamp,omitempty"`
	LastHitTime  *time.Time `json:"last_run_timestamp,omitempty"`
	HitCount     uint       `json:"hit_count"`
	MissCount    uint       `json:"miss_count"`
	State        string     `json:"task_state"`
}

/*
   Unexported type for task is to obfuscate the idea of a fairly
   complex tree data structure from the actual, simple enumeration.
   the data structure represented by Task.Config is roughly
   representational of the internal tree's api: `tree.Add(namespace, node)`
*/
type task struct {
	*Task
	ConfigMap map[string]map[string]interface{} `json:"config"`
}

func (c *Client) NewTask(s *Schedule, wf *wmap.WorkflowMap) *Task {
	return &Task{
		Workflow: wf,
		Schedule: s,
	}
}

// CreateTask takes a pointer to a task structure,
// and POSTs it to Pulse's REST API.
// If an error is encountered during the process it returns
// a generic error.  If an error occured when Pulse attempts
// to create the Task, a type-assertable error is returned.
// Also note that CreateTask modifies the pointed to data
// by adding an ID and a created time.
func (c *Client) CreateTask(t *Task) error {

	// prepare for JSON Marshaling
	jt := &task{
		Task: t,
	}

	// Marshal to JSON for request body
	j, err := json.Marshal(jt)
	if err != nil {
		return err
	}

	resp, err := c.do("POST", "/tasks", j)
	if err != nil {
		return err
	}
	var ctr createTaskReply
	err = json.Unmarshal(resp.body, &ctr)
	if err != nil {
		return err
	}
	if resp.status != 200 {
		return &ErrTaskCreationFailed{ctr.Meta.Message}
	}

	return nil
}

func (c *Client) GetTasks() ([]Task, error) {
	resp, err := c.do("GET", "/tasks", nil)
	if err != nil {
		return nil, err
	}

	var gtr getTasksReply
	err = json.Unmarshal(resp.body, &gtr)
	if err != nil {
		return nil, err
	}
	if resp.status != 200 {
		return nil, &ErrGetTasksFailed{gtr.Meta.Message}
	}
	return gtr.Data.Tasks, nil

}

func (c *Client) StartTask(id uint64) error {
	resp, err := c.do("PUT", fmt.Sprintf("/tasks/%v/start", id))
	if err != nil {
		return err
	}

	var str startTaskReply
	err = json.Unmarshal(resp.body, &str)
	if err != nil {
		return err
	}
	if resp.status != 200 {
		return &ErrStartTask{str.Meta.Message}
	}
	return nil
}

func (c *Client) StopTask(id uint64) error {
	resp, err := c.do("PUT", fmt.Sprintf("/tasks/%v/stop", id))
	if err != nil {
		return err
	}

	var str stopTaskReply
	err = json.Unmarshal(resp.body, &str)
	if err != nil {
		return err
	}
	if resp.status != 200 {
		return &ErrStartTask{str.Meta.Message}
	}
	return nil
}

type createTaskReply struct {
	respBody
	Data createTaskData `json:"data"`
}

type createTaskData struct {
	Task task `json:"task"`
}

type getTasksReply struct {
	respBody
	Data getTasksData `json:"data"`
}

type getTasksData struct {
	Tasks []Task `json:"tasks"`
}

type startTaskReply struct {
	respBody
}

type stopTaskReply struct {
	respBody
}

func makens(ns []string) string {
	return "/" + strings.Join(ns, "/")
}
