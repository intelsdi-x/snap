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
	"time"

	"github.com/intelsdi-x/snap/core"
	"github.com/intelsdi-x/snap/pkg/schedule"
	"github.com/intelsdi-x/snap/scheduler/wmap"
)

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

func AddSchedulerTaskFromTask(t core.Task) *Task {
	st := &Task{
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
	st.assertSchedule(t.Schedule())
	if st.LastRunTimestamp < 0 {
		st.LastRunTimestamp = -1
	}
	return st
}

type Task struct {
	ID                 string            `json:"id"`
	Name               string            `json:"name"`
	Deadline           string            `json:"deadline"`
	Workflow           *wmap.WorkflowMap `json:"workflow,omitempty"`
	Schedule           *core.Schedule    `json:"schedule,omitempty"`
	CreationTimestamp  int64             `json:"creation_timestamp,omitempty"`
	LastRunTimestamp   int64             `json:"last_run_timestamp,omitempty"`
	HitCount           int               `json:"hit_count,omitempty"`
	MissCount          int               `json:"miss_count,omitempty"`
	FailedCount        int               `json:"failed_count,omitempty"`
	LastFailureMessage string            `json:"last_failure_message,omitempty"`
	State              string            `json:"task_state"`
	Href               string            `json:"href"`
}

func (s *Task) CreationTime() time.Time {
	return time.Unix(s.CreationTimestamp, 0)
}

func SchedulerTaskFromTask(t core.Task) *Task {
	st := &Task{
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

type TaskStarted struct {
	// TODO return resource
	ID string `json:"id"`
}

type TaskStopped struct {
	// TODO return resource
	ID string `json:"id"`
}

type TaskRemoved struct {
	// TODO return resource
	ID string `json:"id"`
}

func (t *Task) assertSchedule(s schedule.Schedule) {
	switch v := s.(type) {
	case *schedule.SimpleSchedule:
		t.Schedule = &core.Schedule{
			Type:     "simple",
			Interval: v.Interval.String(),
		}
		return
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

type ScheduledTaskWatchingEnded struct {
}

type StreamedTaskEvent struct {
	// Used to describe the event
	EventType string          `json:"type"`
	Message   string          `json:"message"`
	Event     StreamedMetrics `json:"event,omitempty"`
}

func (s *StreamedTaskEvent) ToJSON() string {
	j, _ := json.Marshal(s)
	return string(j)
}
