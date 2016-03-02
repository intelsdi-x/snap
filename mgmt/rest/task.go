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

package rest

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strings"
	"time"

	"golang.org/x/net/context"
	"google.golang.org/grpc"

	log "github.com/Sirupsen/logrus"
	"github.com/julienschmidt/httprouter"

	"github.com/intelsdi-x/snap/core"
	"github.com/intelsdi-x/snap/mgmt/rest/rbody"
	"github.com/intelsdi-x/snap/mgmt/rest/request"
	cschedule "github.com/intelsdi-x/snap/pkg/schedule"
	"github.com/intelsdi-x/snap/scheduler/rpc"
	"github.com/intelsdi-x/snap/scheduler/wmap"
)

var (
	// The amount of time to buffer streaming events before flushing in seconds
	StreamingBufferWindow = 0.1

	ErrStreamingUnsupported    = errors.New("Streaming unsupported")
	ErrTaskNotFound            = errors.New("Task not found")
	ErrTaskDisabledNotRunnable = errors.New("Task is disabled. Cannot be started")
)

type configItem struct {
	Key   string      `json:"key"`
	Value interface{} `json:"value"`
}

type task struct {
	ID                 uint64                  `json:"id"`
	Config             map[string][]configItem `json:"config"`
	Name               string                  `json:"name"`
	Deadline           string                  `json:"deadline"`
	Workflow           wmap.WorkflowMap        `json:"workflow"`
	Schedule           cschedule.Schedule      `json:"schedule"`
	CreationTime       time.Time               `json:"creation_timestamp,omitempty"`
	LastRunTime        time.Time               `json:"last_run_timestamp,omitempty"`
	HitCount           uint                    `json:"hit_count,omitempty"`
	MissCount          uint                    `json:"miss_count,omitempty"`
	FailedCount        uint                    `json:"failed_count,omitempty"`
	LastFailureMessage string                  `json:"last_failure_message,omitempty"`
	State              string                  `json:"task_state"`
}

func (s *Server) addTask(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {

	tr, err := marshalTask(r.Body)
	if err != nil {
		respond(500, rbody.FromError(err), w)
		return
	}
	sch, err := json.Marshal(tr.Schedule)
	if err != nil {
		respond(400, rbody.FromError(err), w)
		return
	}
	wm, err := json.Marshal(tr.Workflow)
	if err != nil {
		respond(400, rbody.FromError(err), w)
		return
	}
	arg := &rpc.CreateTaskArg{
		ScheduleJson: sch,
		WmapJson:     wm,
		Start:        tr.Start,
		Opts: &rpc.CreateTaskOpts{
			Deadline: tr.Deadline,
			TaskName: tr.Name,
		},
	}
	t, err := s.mt.CreateTask(context.Background(), arg)
	if err != nil {
		respond(500, rbody.FromError(err), w)
		return
	}
	if t.Errors != nil && len(t.Errors) > 0 {
		var errMsg string
		for _, e := range t.Errors {
			errMsg = errMsg + e.ErrorString + " -- "
		}
		respond(500, rbody.FromError(errors.New(errMsg[:len(errMsg)-4])), w)
		return
	}

	taskB := rbody.AddSchedulerTaskFromTask(t.Task)
	taskB.Href = taskURI(r.Host, t.Task)
	respond(201, taskB, w)
}

func (s *Server) getTasks(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	reply, err := s.mt.GetTasks(context.Background(), &rpc.Empty{})
	if err != nil {
		respond(500, rbody.FromError(err), w)
		return
	}

	tasks := &rbody.ScheduledTaskListReturned{}
	tasks.ScheduledTasks = make([]rbody.ScheduledTask, len(reply.Tasks))

	i := 0
	for _, t := range reply.Tasks {
		tasks.ScheduledTasks[i] = *rbody.SchedulerTaskFromTask(t)
		tasks.ScheduledTasks[i].Href = taskURI(r.Host, t)
		i++
	}
	sort.Sort(tasks)
	respond(200, tasks, w)
}

func (s *Server) getTask(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	id := p.ByName("id")
	t, err := s.mt.GetTask(context.Background(), &rpc.GetTaskArg{Id: id})
	if err != nil {
		respond(404, rbody.FromError(err), w)
		return
	}
	task := &rbody.ScheduledTaskReturned{}
	task.AddScheduledTask = *rbody.AddSchedulerTaskFromTask(t)
	task.Href = taskURI(r.Host, t)
	respond(200, task, w)
}

func (s *Server) watchTask(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	logger := log.WithFields(log.Fields{
		"_module": "api",
		"_block":  "watch-task",
		"client":  r.RemoteAddr,
	})

	id := p.ByName("id")

	logger.WithFields(log.Fields{
		"task-id": id,
	}).Debug("request to watch task")

	stream, err := s.mt.WatchTask(context.Background(), &rpc.WatchTaskArg{Id: id})
	if err != nil {
		respond(404, rbody.FromError(err), w)
		return
	}

	// Make this Server Sent Events compatible
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	// get a flusher type
	flusher, ok := w.(http.Flusher)
	if !ok {
		// This only works on ResponseWriters that support streaming
		respond(500, rbody.FromError(ErrStreamingUnsupported), w)
		return
	}
	// send initial stream open event
	so := rbody.StreamedTaskEvent{
		EventType: rpc.Watch_TASK_STREAM_OPEN,
		Message:   "Stream opened",
	}
	fmt.Fprintf(w, "%s\n", so.ToJSON())
	flusher.Flush()

	// Get a channel for if the client notifies us it is closing the connection
	n := w.(http.CloseNotifier).CloseNotify()

	t := time.Now()
	for {
		select {
		case <-n:
			respond(200, &rbody.ScheduledTaskWatchingEnded{}, w)
			return
		default:
			msg, err := stream.Recv()
			if err == io.EOF {
				break
			}
			if err != nil {
				se := rbody.StreamedTaskEvent{
					EventType: rpc.Watch_ERROR,
					Message:   grpc.ErrorDesc(err),
				}
				fmt.Fprintf(w, "%s\n", se.ToJSON())
				flusher.Flush()
				break
			}
			if time.Now().Sub(t).Seconds() < StreamingBufferWindow {
				break
			}
			t = time.Now()
			switch msg.EventType {
			case rpc.Watch_METRICS_COLLECTED, rpc.Watch_TASK_STARTED:
				logger.WithField("msg_type", msg.EventType.String()).Debug(msg.ToJson())
				fmt.Fprintf(w, "%s\n", msg.ToJson())
				flusher.Flush()
			case rpc.Watch_TASK_DISABLED, rpc.Watch_TASK_STOPPED:
				logger.WithField("msg_type", msg.EventType.String()).Debug(msg.ToJson())
				fmt.Fprintf(w, "%s\n", msg.ToJson())
				flusher.Flush()
			}
		}
	}
}

func (s *Server) startTask(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	id := p.ByName("id")
	reply, err := s.mt.StartTask(context.Background(), &rpc.StartTaskArg{Id: id})
	if err != nil {
		respond(500, rbody.FromError(err), w)
		return
	}

	if reply.Errors != nil {
		if strings.Contains(reply.Errors[0].ErrorString, ErrTaskNotFound.Error()) {
			respond(404, rbody.FromSnapErrors(rpc.ConvertSnapErrors(reply.Errors)), w)
			return
		}
		if strings.Contains(reply.Errors[0].ErrorString, ErrTaskDisabledNotRunnable.Error()) {
			respond(409, rbody.FromSnapErrors(rpc.ConvertSnapErrors(reply.Errors)), w)
			return
		}
		respond(500, rbody.FromSnapErrors(rpc.ConvertSnapErrors(reply.Errors)), w)
		return
	}
	// TODO should return resource
	respond(200, &rbody.ScheduledTaskStarted{ID: id}, w)
}

func (s *Server) stopTask(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	id := p.ByName("id")
	reply, err := s.mt.StopTask(context.Background(), &rpc.StopTaskArg{Id: id})
	if err != nil {
		respond(404, rbody.FromError(errors.New(grpc.ErrorDesc(err))), w)
		return
	}
	if reply.Errors != nil {
		if strings.Contains(reply.Errors[0].Error(), ErrTaskNotFound.Error()) {
			respond(404, rbody.FromSnapErrors(rpc.ConvertSnapErrors(reply.Errors)), w)
			return
		}
		respond(500, rbody.FromSnapErrors(rpc.ConvertSnapErrors(reply.Errors)), w)
		return
	}
	respond(200, &rbody.ScheduledTaskStopped{ID: id}, w)
}

func (s *Server) removeTask(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	id := p.ByName("id")
	_, err := s.mt.RemoveTask(context.Background(), &rpc.RemoveTaskArg{Id: id})
	if err != nil {
		if strings.Contains(err.Error(), ErrTaskNotFound.Error()) {
			respond(404, rbody.FromError(errors.New(grpc.ErrorDesc(err))), w)
			return
		}
		respond(500, rbody.FromError(errors.New(grpc.ErrorDesc(err))), w)
		return
	}
	respond(200, &rbody.ScheduledTaskRemoved{ID: id}, w)
}

//enableTask changes the task state from Disabled to Stopped
func (s *Server) enableTask(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	id := p.ByName("id")
	tsk, err := s.mt.EnableTask(context.Background(), &rpc.EnableTaskArg{Id: id})
	if err != nil {
		if strings.Contains(err.Error(), ErrTaskNotFound.Error()) {
			respond(404, rbody.FromError(errors.New(grpc.ErrorDesc(err))), w)
			return
		}
		respond(500, rbody.FromError(errors.New(grpc.ErrorDesc(err))), w)
		return
	}
	task := &rbody.ScheduledTaskEnabled{}
	task.AddScheduledTask = *rbody.AddSchedulerTaskFromTask(tsk)
	respond(200, task, w)
}

func marshalTask(body io.ReadCloser) (*request.TaskCreationRequest, error) {
	var tr request.TaskCreationRequest
	errCode, err := marshalBody(&tr, body)
	if errCode != 0 && err != nil {
		return nil, err
	}
	return &tr, nil
}

func taskURI(host string, t core.Task) string {
	return fmt.Sprintf("%s://%s/v1/tasks/%s", protocolPrefix, host, t.ID())
}
