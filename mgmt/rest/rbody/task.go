package rbody

import (
	"fmt"
	"time"

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

func (s *ScheduledTaskListReturned) Len() int {
	return len(s.ScheduledTasks)
}

func (s *ScheduledTaskListReturned) Less(i, j int) bool {
	return s.ScheduledTasks[i].Name < s.ScheduledTasks[j].Name
}

func (s *ScheduledTaskListReturned) Swap(i, j int) {
	s.ScheduledTasks[i], s.ScheduledTasks[j] = s.ScheduledTasks[j], s.ScheduledTasks[i]
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

type ScheduledTask struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	Deadline string `json:"deadline"`
	// Workflow           *wmap.WorkflowMap `json:"workflow"`
	// Schedule           Schedule          `json:"schedule"`
	CreationTimestamp  int64  `json:"creation_timestamp,omitempty"`
	LastRunTimestamp   int64  `json:"last_run_timestamp,omitempty"`
	HitCount           int    `json:"hit_count,omitempty"`
	MissCount          int    `json:"miss_count,omitempty"`
	FailedCount        int    `json:"failed_count,omitempty"`
	LastFailureMessage string `json:"last_failure_message,omitempty"`
	State              string `json:"task_state"`
}

func (s *ScheduledTask) CreationTime() time.Time {
	return time.Unix(s.CreationTimestamp, 0)
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
