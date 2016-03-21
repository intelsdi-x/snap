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
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/intelsdi-x/snap/core"
	"github.com/intelsdi-x/snap/core/serror"
	"github.com/intelsdi-x/snap/pkg/schedule"
	"github.com/intelsdi-x/snap/scheduler/wmap"
	"github.com/prometheus/common/log"
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
		TimeLastRun: &Time{Sec: t.LastRunTime().Unix(), Nsec: int64(t.LastRunTime().Nanosecond())},
		TimeCreated: &Time{Sec: t.CreationTime().Unix(), Nsec: int64(t.CreationTime().Nanosecond())},
		Deadline:    t.DeadlineDuration().String(),
		WmapJson:    wmap,
		StopOnFail:  uint64(t.GetStopOnFailure()),
	}
	return tsk, err
}

type SnapErrors []*SnapError

func ConvertSnapErrors(s []*SnapError) []serror.SnapError {
	rerrs := make([]serror.SnapError, len(s))
	for i, err := range s {
		rerrs[i] = serror.New(errors.New(err.ErrorString), err.Fields())
	}
	return rerrs
}

func NewErrors(errs []serror.SnapError) []*SnapError {
	errors := make([]*SnapError, len(errs))
	for i, err := range errs {
		fields := make(map[string]string)
		for k, v := range err.Fields() {
			switch t := v.(type) {
			case string:
				fields[k] = t
			case int:
				fields[k] = strconv.Itoa(t)
			case float64:
				fields[k] = strconv.FormatFloat(t, 'f', -1, 64)
			default:
				log.Errorf("Unexpected type %v\n", t)
			}
		}
		errors[i] = &SnapError{ErrorFields: fields, ErrorString: err.Error()}
	}
	return errors
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

func (s *SnapError) Error() string {
	return s.ErrorString
}

func (s *SnapError) Fields() map[string]interface{} {
	fields := make(map[string]interface{}, len(s.ErrorFields))
	for key, value := range s.ErrorFields {
		fields[key] = value
	}
	return fields
}

func NewMetrics(ms []core.Metric) []*Metric {
	metrics := make([]*Metric, len(ms))
	for i, m := range ms {
		metrics[i] = &Metric{
			Namespace: m.Namespace(),
			Version:   uint64(m.Version()),
			Source:    m.Source(),
			Tags:      m.Tags(),
			Timestamp: &Time{
				Sec:  m.Timestamp().Unix(),
				Nsec: int64(m.Timestamp().Nanosecond()),
			},
		}
		metrics[i].Labels = make([]*Label, len(m.Labels()))
		for y, label := range m.Labels() {
			metrics[i].Labels[y] = &Label{
				Index: uint64(label.Index),
				Name:  label.Name,
			}
		}
		var b bytes.Buffer
		enc := gob.NewEncoder(&b)
		switch t := m.Data().(type) {
		case string:
			enc.Encode(t)
			metrics[i].Data = b.Bytes()
			metrics[i].DataType = "string"
		case float64:
			enc.Encode(t)
			metrics[i].Data = b.Bytes()
			metrics[i].DataType = "float64"
		case float32:
			enc.Encode(t)
			metrics[i].Data = b.Bytes()
			metrics[i].DataType = "float32"
		case int32:
			enc.Encode(t)
			metrics[i].Data = b.Bytes()
			metrics[i].DataType = "int32"
		case int:
			enc.Encode(t)
			metrics[i].Data = b.Bytes()
			metrics[i].DataType = "int"
		case int64:
			enc.Encode(t)
			metrics[i].Data = b.Bytes()
			metrics[i].DataType = "int64"
		default:
			panic(t)
		}
	}
	return metrics
}

func (w *Watch) ToJson() string {
	type M struct {
		Namespace string `json:"namespace"`
		Timestamp string `json:"timestamp"`
		DataType  string `json:"data_type"`
		Data      string `json:"data"`
		Source    string `json:"source"`
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
			Namespace: strings.Join(i.Namespace, "/"),
			Timestamp: time.Unix(i.Timestamp.Sec, i.Timestamp.Nsec).String(),
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
