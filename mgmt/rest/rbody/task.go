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

package rbody

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/intelsdi-x/snap/core"
	"github.com/intelsdi-x/snap/mgmt/rest/request"
	"github.com/intelsdi-x/snap/pkg/schedule"
	"github.com/intelsdi-x/snap/scheduler/wmap"
)

// const A list of constants
const (
	ScheduledTaskListReturnedType  = "scheduled_task_list_returned"
	ScheduledTaskReturnedType      = "scheduled_task_returned"
	AddScheduledTaskType           = "scheduled_task_created"
	ScheduledTaskType              = "scheduled_task"
	ScheduledTaskStartedType       = "scheduled_task_started"
	ScheduledTaskStoppedType       = "scheduled_task_stopped"
	ScheduledTaskRemovedType       = "scheduled_task_removed"
	ScheduledTaskWatchingEndedType = "schedule_task_watch_ended"
	ScheduledTaskEnabledType       = "scheduled_task_enabled"

	// Event types for task watcher streaming
	TaskWatchStreamOpen   = "stream-open"
	TaskWatchMetricEvent  = "metric-event"
	TaskWatchTaskDisabled = "task-disabled"
	TaskWatchTaskStarted  = "task-started"
	TaskWatchTaskStopped  = "task-stopped"
)

// ScheduledTaskListReturned Array of ScheduledTasks
type ScheduledTaskListReturned struct {
	ScheduledTasks []ScheduledTask
}

// Len The length of scheduled tasks
func (s *ScheduledTaskListReturned) Len() int {
	return len(s.ScheduledTasks)
}

// Less The bool result of comprison of two scheduled tasks
func (s *ScheduledTaskListReturned) Less(i, j int) bool {
	return s.ScheduledTasks[i].Name < s.ScheduledTasks[j].Name
}

// Swap Swapping two scheduled tasks
func (s *ScheduledTaskListReturned) Swap(i, j int) {
	s.ScheduledTasks[i], s.ScheduledTasks[j] = s.ScheduledTasks[j], s.ScheduledTasks[i]
}

// ResponseBodyMessage returns a string message
func (s *ScheduledTaskListReturned) ResponseBodyMessage() string {
	return "Scheduled tasks retrieved"
}

// ResponseBodyType returns a string response body type
func (s *ScheduledTaskListReturned) ResponseBodyType() string {
	return ScheduledTaskListReturnedType
}

// ScheduledTaskReturned struct type is same as AddScheduledTask
type ScheduledTaskReturned struct {
	AddScheduledTask
}

// ResponseBodyMessage returns a string of response with task id
func (s *ScheduledTaskReturned) ResponseBodyMessage() string {
	return fmt.Sprintf("Scheduled task (%s) returned", s.ID)
}

// ResponseBodyType returns a string response body type
func (s *ScheduledTaskReturned) ResponseBodyType() string {
	return ScheduledTaskReturnedType
}

// AddScheduledTask data type
type AddScheduledTask ScheduledTask

// ResponseBodyMessage returns a string of response with task id
func (s *AddScheduledTask) ResponseBodyMessage() string {
	return fmt.Sprintf("Scheduled task created (%s)", s.ID)
}

// ResponseBodyType returns a string of response body type
func (s *AddScheduledTask) ResponseBodyType() string {
	return AddScheduledTaskType
}

// AddSchedulerTaskFromTask returns the added schedule task
func AddSchedulerTaskFromTask(t core.Task) *AddScheduledTask {
	st := &AddScheduledTask{
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
		Workflow:           t.WMap(),
	}
	assertSchedule(t.Schedule(), st)
	if st.LastRunTimestamp < 0 {
		st.LastRunTimestamp = -1
	}
	return st
}

// ScheduledTask struct type
type ScheduledTask struct {
	ID                 string            `json:"id"`
	Name               string            `json:"name"`
	Deadline           string            `json:"deadline"`
	Workflow           *wmap.WorkflowMap `json:"workflow,omitempty"`
	Schedule           *request.Schedule `json:"schedule,omitempty"`
	CreationTimestamp  int64             `json:"creation_timestamp,omitempty"`
	LastRunTimestamp   int64             `json:"last_run_timestamp,omitempty"`
	HitCount           int               `json:"hit_count,omitempty"`
	MissCount          int               `json:"miss_count,omitempty"`
	FailedCount        int               `json:"failed_count,omitempty"`
	LastFailureMessage string            `json:"last_failure_message,omitempty"`
	State              string            `json:"task_state"`
	Href               string            `json:"href"`
}

// CreationTime returns the unix time of a scheduled task
func (s *ScheduledTask) CreationTime() time.Time {
	return time.Unix(s.CreationTimestamp, 0)
}

// ResponseBodyMessage returns a string response with task id
func (s *ScheduledTask) ResponseBodyMessage() string {
	return fmt.Sprintf("Scheduled task created (%s)", s.ID)
}

// ResponseBodyType returns a string response body type
func (s *ScheduledTask) ResponseBodyType() string {
	return ScheduledTaskType
}

// SchedulerTaskFromTask transforms core.Task to the ScheduledTask
func SchedulerTaskFromTask(t core.Task) *ScheduledTask {
	st := &ScheduledTask{
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

// ScheduledTaskStarted struct type
type ScheduledTaskStarted struct {
	// TODO return resource
	ID string `json:"id"`
}

// ResponseBodyMessage returns task started response with the task id
func (s *ScheduledTaskStarted) ResponseBodyMessage() string {
	return fmt.Sprintf("Scheduled task (%s) started", s.ID)
}

// ResponseBodyType returns a string respose body message
func (s *ScheduledTaskStarted) ResponseBodyType() string {
	return ScheduledTaskStartedType
}

// ScheduledTaskStopped struct type
type ScheduledTaskStopped struct {
	// TODO return resource
	ID string `json:"id"`
}

// ResponseBodyMessage returns the scheduled task stopped response
// with the task id
func (s *ScheduledTaskStopped) ResponseBodyMessage() string {
	return fmt.Sprintf("Scheduled task (%s) stopped", s.ID)
}

// ResponseBodyType returns the string response of scheduled task stopped
func (s *ScheduledTaskStopped) ResponseBodyType() string {
	return ScheduledTaskStoppedType
}

// ScheduledTaskRemoved struct type
type ScheduledTaskRemoved struct {
	// TODO return resource
	ID string `json:"id"`
}

// ResponseBodyMessage returns the scheduled task removed response
func (s *ScheduledTaskRemoved) ResponseBodyMessage() string {
	return fmt.Sprintf("Scheduled task (%s) removed", s.ID)
}

// ResponseBodyType returns a string of the scheduled task removed response
func (s *ScheduledTaskRemoved) ResponseBodyType() string {
	return ScheduledTaskRemovedType
}

// ScheduledTaskEnabled struct typeß
type ScheduledTaskEnabled struct {
	AddScheduledTask
}

// ResponseBodyMessage returns a string of disabled task enabled response
func (s *ScheduledTaskEnabled) ResponseBodyMessage() string {
	return fmt.Sprintf("Disabled task (%s) enabled", s.AddScheduledTask.ID)
}

// ResponseBodyType returns a string of task enabled response type
func (s *ScheduledTaskEnabled) ResponseBodyType() string {
	return ScheduledTaskEnabledType
}

func assertSchedule(s schedule.Schedule, t *AddScheduledTask) {
	switch v := s.(type) {
	case *schedule.SimpleSchedule:
		t.Schedule = &request.Schedule{
			Type:     "simple",
			Interval: v.Interval.String(),
		}
		return
	}
	t.Schedule = &request.Schedule{}
}

// ScheduledTaskWatchingEnded struct type
type ScheduledTaskWatchingEnded struct {
}

// ResponseBodyMessage string response
func (s *ScheduledTaskWatchingEnded) ResponseBodyMessage() string {
	return "Task watching ended"
}

// ResponseBodyType returns a string response of ScheduledTaskWatchingEndedType
func (s *ScheduledTaskWatchingEnded) ResponseBodyType() string {
	return ScheduledTaskWatchingEndedType
}

// StreamedTaskEvent struct type
type StreamedTaskEvent struct {
	// Used to describe the event
	EventType string          `json:"type"`
	Message   string          `json:"message"`
	Event     StreamedMetrics `json:"event,omitempty"`
}

// ToJSON returns JSON string of the stream task event
func (s *StreamedTaskEvent) ToJSON() string {
	j, _ := json.Marshal(s)
	return string(j)
}

// StreamedMetric struct type
type StreamedMetric struct {
	Namespace string      `json:"namespace"`
	Data      interface{} `json:"data"`
	Source    string      `json:"source"`
	Timestamp time.Time   `json:"timestamp"`
}

// StreamedMetrics Array of streamed metrics
type StreamedMetrics []StreamedMetric

// Len The length of streamed metrics
func (s StreamedMetrics) Len() int {
	return len(s)
}

// Less The bool comparison result of two stream metrics
func (s StreamedMetrics) Less(i, j int) bool {
	return fmt.Sprintf("%s:%s", s[i].Source, s[i].Namespace) < fmt.Sprintf("%s:%s", s[j].Source, s[j].Namespace)
}

// Swap Swapping two streamed metrics
func (s StreamedMetrics) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}
