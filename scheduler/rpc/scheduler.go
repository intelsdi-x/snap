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

package rpc

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"strconv"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/intelsdi-x/snap/core"
	"github.com/intelsdi-x/snap/internal/common"
	"github.com/intelsdi-x/snap/pkg/schedule"
	"github.com/intelsdi-x/snap/scheduler/wmap"
)

func NewTask(t core.Task) (*Task, error) {
	wmap, err := t.WMap().ToJson()
	tsk := &Task{
		Id:          t.ID(),
		TaskState:   uint64(t.State()),
		Name:        t.GetName(),
		Hit:         uint64(t.HitCount()),
		Misses:      uint64(t.MissedCount()),
		Failed:      uint64(t.FailedCount()),
		LastFailure: t.LastFailureMessage(),
		TimeLastRun: &common.Time{Sec: t.LastRunTime().Unix(), Nsec: int64(t.LastRunTime().Nanosecond())},
		TimeCreated: &common.Time{Sec: t.CreationTime().Unix(), Nsec: int64(t.CreationTime().Nanosecond())},
		Deadline:    t.DeadlineDuration().String(),
		WmapJson:    wmap,
		StopOnFail:  uint64(t.GetStopOnFailure()),
	}
	return tsk, err
}

func (t *Task) ID() string {
	return t.Id
}

func (t *Task) CreationTime() *time.Time {
	ti := time.Unix(t.TimeCreated.Sec, t.TimeCreated.Nsec)
	return &ti
}

func (t *Task) DeadlineDuration() time.Duration {
	d, err := time.ParseDuration(t.Deadline)
	if err != nil {
		log.Error("Failed to parse the deadline duration")
	}
	return d
}

func (t *Task) FailedCount() uint {
	return uint(t.Failed)
}

func (t *Task) MissedCount() uint {
	return uint(t.Misses)
}

func (t *Task) HitCount() uint {
	return uint(t.Hit)
}

func (t *Task) GetName() string {
	return t.Name
}

func (t *Task) GetStopOnFailure() uint {
	return uint(t.StopOnFail)
}

func (t *Task) LastFailureMessage() string {
	return t.LastFailure
}

func (t *Task) LastRunTime() *time.Time {
	ti := time.Unix(t.TimeLastRun.Sec, t.TimeLastRun.Nsec)
	return &ti
}

func (t *Task) State() core.TaskState {
	return core.TaskState(t.TaskState)
}

func (t *Task) WMap() *wmap.WorkflowMap {
	wm := wmap.NewWorkflowMap()
	json.Unmarshal(t.WmapJson, wm)
	return wm
}

func (t *Task) Schedule() schedule.Schedule {
	return schedule.NewSimpleSchedule(1 * time.Second)
}

func (t *Task) SetDeadlineDuration(_ time.Duration) {
}

func (t *Task) SetID(_ string) {
}

func (t *Task) SetTaskID(_ string) {
}

func (t *Task) SetName(_ string) {
}

func (t *Task) SetStopOnFailure(_ uint) {
}

func (w *Watch) ToJSON() string {
	type M struct {
		Namespace string            `json:"namespace"`
		Timestamp time.Time         `json:"timestamp"`
		DataType  string            `json:"data_type"`
		Data      string            `json:"data"`
		Source    string            `json:"source"`
		Tags      map[string]string `json:"tags"`
	}
	wt := &struct {
		EventType Watch_EventType `json:"event_type"`
		Message   string          `json:"message"`
		Events    []M             `json:"events"`
	}{
		EventType: w.EventType,
		Message:   w.Message,
		Events:    make([]M, len(w.Events)),
	}

	for idx, i := range w.Events {
		wt.Events[idx] = M{
			Namespace: core.JoinNamespace(i.Namespace),
			Timestamp: time.Unix(i.Timestamp.Sec, i.Timestamp.Nsec),
			Source:    i.Source,
		}
		switch i.DataType {
		case "int":
			var val int
			buf := bytes.NewBuffer(i.Data)
			decoder := gob.NewDecoder(buf)
			decoder.Decode(&val)
			wt.Events[idx].Data = strconv.Itoa(val)
		case "int32":
			var val int32
			buf := bytes.NewBuffer(i.Data)
			decoder := gob.NewDecoder(buf)
			decoder.Decode(&val)
			wt.Events[idx].Data = strconv.FormatInt(int64(val), 10)
		case "int64":
			var val int64
			buf := bytes.NewBuffer(i.Data)
			decoder := gob.NewDecoder(buf)
			decoder.Decode(&val)
			wt.Events[idx].Data = strconv.FormatInt(val, 10)
		case "float32":
			var val float32
			buf := bytes.NewBuffer(i.Data)
			decoder := gob.NewDecoder(buf)
			decoder.Decode(&val)
			wt.Events[idx].Data = strconv.FormatFloat(float64(val), 'E', -1, 32)
		case "float64":
			var val float64
			buf := bytes.NewBuffer(i.Data)
			decoder := gob.NewDecoder(buf)
			decoder.Decode(&val)
			wt.Events[idx].Data = strconv.FormatFloat(val, 'E', -1, 64)
		case "string":
			var val string
			buf := bytes.NewBuffer(i.Data)
			decoder := gob.NewDecoder(buf)
			decoder.Decode(&val)
			wt.Events[idx].Data = val
		}
		wt.Events[idx].DataType = i.DataType
	}

	j, err := json.Marshal(wt)
	if err != nil {
		log.Error(err)
	}
	return string(j)
}
