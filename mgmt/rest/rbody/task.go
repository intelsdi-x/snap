package rbody

import (
	"fmt"

	"github.com/intelsdi-x/pulse/core"
)

const (
	ScheduledTaskListReturnedType = "scheduled_task_list_returned"
	AddScheduledTaskType          = "scheduled_task_created"
	ScheduledTaskType             = "scheduled_task"
	ScheduledTaskStartedType      = "scheduled_task_started"
	ScheduledTaskStoppedType      = "scheduled_task_stopped"
	ScheduledTaskRemovedType      = "scheduled_task_removed"
)

type ScheduledTaskListReturned struct {
	ScheduledTasks []ScheduledTask
}

func (s *ScheduledTaskListReturned) ResponseBodyMessage() string {
	return "Scheduled tasks retrieved"
}

func (s *ScheduledTaskListReturned) ResponseBodyType() string {
	return ScheduledTaskListReturnedType
}

type AddScheduledTask ScheduledTask

func (s *AddScheduledTask) ResponseBodyMessage() string {
	return fmt.Sprintf("Scheduled task created (%d)", s.ID)
}

func (s *AddScheduledTask) ResponseBodyType() string {
	return AddScheduledTaskType
}

func AddSchedulerTaskFromTask(t core.Task) *AddScheduledTask {
	// TODO workflow back from core.Task
	// TODO schedule back from core.Task
	st := &AddScheduledTask{
		ID:                 int(t.ID()),
		Name:               t.GetName(),
		Deadline:           t.DeadlineDuration().String(),
		CreationTimestamp:  int(t.CreationTime().Unix()),
		LastRunTimestamp:   int(t.LastRunTime().Unix()),
		HitCount:           int(t.HitCount()),
		MissCount:          int(t.MissedCount()),
		FailedCount:        int(t.FailedCount()),
		LastFailureMessage: t.LastFailureMessage(),
		State:              t.State().String(),
	}
	return st
}

type ScheduledTask struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	Deadline string `json:"deadline"`
	// Workflow           *wmap.WorkflowMap `json:"workflow"`
	// Schedule           Schedule          `json:"schedule"`
	CreationTimestamp  int    `json:"creation_timestamp,omitempty"`
	LastRunTimestamp   int    `json:"last_run_timestamp,omitempty"`
	HitCount           int    `json:"hit_count,omitempty"`
	MissCount          int    `json:"miss_count,omitempty"`
	FailedCount        int    `json:"failed_count,omitempty"`
	LastFailureMessage string `json:"last_failure_message,omitempty"`
	State              string `json:"task_state"`
}

func (s *ScheduledTask) ResponseBodyMessage() string {
	return fmt.Sprintf("Scheduled task created (%d)", s.ID)
}

func (s *ScheduledTask) ResponseBodyType() string {
	return ScheduledTaskType
}

func SchedulerTaskFromTask(t core.Task) *ScheduledTask {
	// TODO workflow back from core.Task
	// TODO schedule back from core.Task
	st := &ScheduledTask{
		ID:                 int(t.ID()),
		Name:               t.GetName(),
		Deadline:           t.DeadlineDuration().String(),
		CreationTimestamp:  int(t.CreationTime().Unix()),
		LastRunTimestamp:   int(t.LastRunTime().Unix()),
		HitCount:           int(t.HitCount()),
		MissCount:          int(t.MissedCount()),
		FailedCount:        int(t.FailedCount()),
		LastFailureMessage: t.LastFailureMessage(),
		State:              t.State().String(),
	}
	return st
}

type ScheduledTaskStarted struct {
	// TODO return resource
	ID int `json:"id"`
}

func (s *ScheduledTaskStarted) ResponseBodyMessage() string {
	return fmt.Sprintf("Scheduled task (%d) started", s.ID)
}

func (s *ScheduledTaskStarted) ResponseBodyType() string {
	return "scheduled_task_started"
}

type ScheduledTaskStopped struct {
	// TODO return resource
	ID int `json:"id"`
}

func (s *ScheduledTaskStopped) ResponseBodyMessage() string {
	return fmt.Sprintf("Scheduled task (%d) stopped", s.ID)
}

func (s *ScheduledTaskStopped) ResponseBodyType() string {
	return "scheduled_task_stopped"
}

type ScheduledTaskRemoved struct {
	// TODO return resource
	ID int `json:"id"`
}

func (s *ScheduledTaskRemoved) ResponseBodyMessage() string {
	return fmt.Sprintf("Scheduled task (%d) removed", s.ID)
}

func (s *ScheduledTaskRemoved) ResponseBodyType() string {
	return "scheduled_task_removed"
}
