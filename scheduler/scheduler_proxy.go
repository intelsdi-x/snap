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

package scheduler

import (
	"encoding/json"
	"errors"
	"time"

	log "github.com/Sirupsen/logrus"
	"golang.org/x/net/context"

	"github.com/intelsdi-x/snap/core"
	"github.com/intelsdi-x/snap/internal/common"
	"github.com/intelsdi-x/snap/pkg/schedule"
	"github.com/intelsdi-x/snap/scheduler/rpc"
	"github.com/intelsdi-x/snap/scheduler/wmap"
)

type schedulerProxy struct {
	scheduler *scheduler
}

func (s *schedulerProxy) GetTask(ctx context.Context, arg *rpc.GetTaskArg) (*rpc.Task, error) {
	task, err := s.scheduler.getTask(arg.Id)
	if err != nil {
		return nil, err
	}
	return rpc.NewTask(task)
}

func (s *schedulerProxy) GetTasks(ctx context.Context, arg *common.Empty) (*rpc.GetTasksReply, error) {
	tasks := s.scheduler.GetTasks()
	reply := &rpc.GetTasksReply{Tasks: make(map[string]*rpc.Task)}
	for id, task := range tasks {
		t, err := rpc.NewTask(task)
		if err != nil {
			return nil, err
		}
		reply.Tasks[id] = t
	}
	return reply, nil
}

func (s *schedulerProxy) StartTask(ctx context.Context, arg *rpc.StartTaskArg) (*rpc.StartTaskReply, error) {
	errs := s.scheduler.StartTask(arg.Id)
	return &rpc.StartTaskReply{Errors: rpc.NewErrors(errs)}, nil
}

func (s *schedulerProxy) StopTask(ctx context.Context, arg *rpc.StopTaskArg) (*rpc.StopTaskReply, error) {
	errs := s.scheduler.StopTask(arg.Id)
	return &rpc.StopTaskReply{Errors: rpc.NewErrors(errs)}, nil
}

func (s *schedulerProxy) RemoveTask(ctx context.Context, arg *rpc.RemoveTaskArg) (*common.Empty, error) {
	return &common.Empty{}, s.scheduler.RemoveTask(arg.Id)
}

func (s *schedulerProxy) EnableTask(ctx context.Context, arg *rpc.EnableTaskArg) (*rpc.Task, error) {
	task, err := s.scheduler.EnableTask(arg.Id)
	if err != nil {
		return nil, err
	}
	return rpc.NewTask(task)
}

func (s *schedulerProxy) WatchTask(arg *rpc.WatchTaskArg, stream rpc.TaskManager_WatchTaskServer) error {
	twh := newTaskWatchHandler()
	tw, err := s.scheduler.WatchTask(arg.Id, twh)
	if err != nil {
		return err
	}
	for {
		msg := <-twh.mChan
		if msg.EventType == rpc.Watch_TASK_STOPPED || msg.EventType == rpc.Watch_TASK_DISABLED {
			err := stream.Send(msg)
			if err != nil {
				log.Error(err)
				tw.Close()
				return err
			}
			tw.Close()
			return nil
		}
		if err := stream.Send(msg); err != nil {
			tw.Close()
			log.Error(err)
			return err
		}
	}
}

func (s *schedulerProxy) CreateTask(ctx context.Context, arg *rpc.CreateTaskArg) (*rpc.CreateTaskReply, error) {
	var w *wmap.WorkflowMap
	var sch schedule.Schedule

	json.Unmarshal(arg.WmapJson, &w)

	st := struct {
		Type string `json:"type"`
	}{}
	json.Unmarshal(arg.ScheduleJson, &st)
	switch st.Type {
	case "simple":
		var simpleSch *schedule.SimpleSchedule
		json.Unmarshal(arg.ScheduleJson, &simpleSch)
		sch = simpleSch
	case "windowed":
		var windowedSch *schedule.WindowedSchedule
		json.Unmarshal(arg.ScheduleJson, &windowedSch)
		sch = windowedSch
	default:
		return nil, errors.New("unknown schedule type " + st.Type)
	}

	var opts []core.TaskOption
	if arg.Opts != nil {
		if arg.Opts.Deadline != "" {
			dl, err := time.ParseDuration(arg.Opts.Deadline)
			if err != nil {
				return nil, err
			}
			opts = append(opts, core.TaskDeadlineDurationOption(dl))
		}
		if arg.Opts.TaskName != "" {
			opts = append(opts, core.SetTaskNameOption(arg.Opts.TaskName))
		}
		if arg.Opts.StopOnFail > 0 {
			opts = append(opts, core.StopOnFailureOption(uint(arg.Opts.StopOnFail)))
		}
		if arg.Opts.TaskId != "" {
			opts = append(opts, core.SetTaskIdOption(arg.Opts.TaskId))
		}
	}

	reply := &rpc.CreateTaskReply{}
	t, errs := s.scheduler.CreateTask(sch, w, arg.Start, opts...)
	if errs != nil {
		reply.Errors = rpc.NewErrors(errs.Errors())
	}
	if t != nil {
		task, err := rpc.NewTask(t)
		if err != nil {
			return nil, err
		}
		reply.Task = task
	}
	return reply, nil
}

type taskWatchHandler struct {
	mChan chan *rpc.Watch
}

func newTaskWatchHandler() *taskWatchHandler {
	return &taskWatchHandler{
		mChan: make(chan *rpc.Watch),
	}
}

func (t *taskWatchHandler) CatchCollection(m []core.Metric) {
	t.mChan <- &rpc.Watch{
		EventType: rpc.Watch_METRICS_COLLECTED,
		Events:    rpc.NewMetrics(m),
	}
}

func (t *taskWatchHandler) CatchTaskStarted() {
	t.mChan <- &rpc.Watch{
		EventType: rpc.Watch_TASK_STARTED,
	}
}

func (t *taskWatchHandler) CatchTaskStopped() {
	t.mChan <- &rpc.Watch{
		EventType: rpc.Watch_TASK_STOPPED,
	}
}

func (t *taskWatchHandler) CatchTaskDisabled(why string) {
	t.mChan <- &rpc.Watch{
		EventType: rpc.Watch_TASK_DISABLED,
		Message:   why,
	}
}
