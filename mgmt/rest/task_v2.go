/*
http://www.apache.org/licenses/LICENSE-2.0.txt


Copyright 2015,2016 Intel Corporation

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
	"fmt"
	"net/http"
	"sort"

	"errors"
	"github.com/intelsdi-x/snap/core"
	"github.com/intelsdi-x/snap/core/serror"
	"github.com/intelsdi-x/snap/mgmt/rest/v2/response"
	"github.com/julienschmidt/httprouter"
	"strings"
	"time"
)

var (
	// The amount of time to buffer streaming events before flushing in seconds
	StreamingBufferWindow = 0.1

	ErrStreamingUnsupported    = errors.New("Streaming unsupported")
	ErrTaskNotFound            = errors.New("Task not found")
	ErrTaskDisabledNotRunnable = errors.New("Task is disabled. Cannot be started")
	ErrNoActionSpecified       = errors.New("No action was specified in the request")
	ErrWrongAction             = errors.New("Wrong action requested")
)

func (s *Server) addTaskV2(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	task, err := core.CreateTaskFromContent(r.Body, nil, s.mt.CreateTask)
	if err != nil {
		response.Write(500, response.FromError(err), w)
		return
	}
	taskB := response.AddSchedulerTaskFromTask(task)
	taskB.Href = taskURI(r.Host, "v2", task)
	response.Write(201, taskB, w)
}

func (s *Server) getTasksV2(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	// get tasks from the task manager
	sts := s.mt.GetTasks()

	// create the task list response
	tasks := make(response.Tasks, len(sts))
	i := 0
	for _, t := range sts {
		tasks[i] = response.SchedulerTaskFromTask(t)
		tasks[i].Href = taskURI(r.Host, "v2", t)
		i++
	}
	sort.Sort(tasks)

	response.Write(200, tasks, w)
}

func (s *Server) getTaskV2(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	id := p.ByName("id")
	t, err := s.mt.GetTask(id)
	if err != nil {
		response.Write(404, response.FromError(err), w)
		return
	}
	task := response.AddSchedulerTaskFromTask(t)
	task.Href = taskURI(r.Host, "v2", t)
	response.Write(200, task, w)
}

func (s *Server) watchTaskV2(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	s.wg.Add(1)
	defer s.wg.Done()

	id := p.ByName("id")

	tw := &TaskWatchHandlerV2{
		alive: true,
		mChan: make(chan response.StreamedTaskEvent),
	}
	tc, err1 := s.mt.WatchTask(id, tw)
	if err1 != nil {
		if strings.Contains(err1.Error(), ErrTaskNotFound.Error()) {
			response.Write(404, response.FromError(err1), w)
			return
		}
		response.Write(500, response.FromError(err1), w)
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
		response.Write(500, response.FromError(ErrStreamingUnsupported), w)
		return
	}
	// send initial stream open event
	so := response.StreamedTaskEvent{
		EventType: response.TaskWatchStreamOpen,
		Message:   "Stream opened",
	}
	fmt.Fprintf(w, "data: %s\n\n", so.ToJSON())
	flusher.Flush()

	// Get a channel for if the client notifies us it is closing the connection
	n := w.(http.CloseNotifier).CloseNotify()
	t := time.Now()
	for {
		// Write to the ResponseWriter
		select {
		case e := <-tw.mChan:
			switch e.EventType {
			case response.TaskWatchMetricEvent, response.TaskWatchTaskStarted:
				// The client can decide to stop receiving on the stream on Task Stopped.
				// We write the event to the buffer
				fmt.Fprintf(w, "data: %s\n\n", e.ToJSON())
			case response.TaskWatchTaskDisabled, response.TaskWatchTaskStopped:
				// A disabled task should end the streaming and close the connection
				fmt.Fprintf(w, "data: %s\n\n", e.ToJSON())
				// Flush since we are sending nothing new
				flusher.Flush()
				// Close out watcher removing it from the scheduler
				tc.Close()
				// exit since this client is no longer listening
				response.Write(204, nil, w)
			}
			// If we are at least above our minimum buffer time we flush to send
			if time.Now().Sub(t).Seconds() > StreamingBufferWindow {
				flusher.Flush()
				t = time.Now()
			}
		case <-n:
			// Flush since we are sending nothing new
			flusher.Flush()
			// Close out watcher removing it from the scheduler
			tc.Close()
			// exit since this client is no longer listening
			response.Write(204, nil, w)
			return
		case <-s.killChan:
			// Flush since we are sending nothing new
			flusher.Flush()
			// Close out watcher removing it from the scheduler
			tc.Close()
			// exit since this client is no longer listening
			response.Write(204, nil, w)
			return
		}
	}
}

func (s *Server) updateTask(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	errs := make([]serror.SnapError, 0, 1)
	id := p.ByName("id")
	action, exist := r.URL.Query()["action"]
	if !exist && len(action) > 0 {
		errs = append(errs, serror.New(ErrNoActionSpecified))
	} else {
		switch action[0] {
		case "enable":
			_, err := s.mt.EnableTask(id)
			if err != nil {
				errs = append(errs, serror.New(err))
			}
		case "start":
			errs = s.mt.StartTask(id)
		case "stop":
			errs = s.mt.StopTask(id)
		default:
			errs = append(errs, serror.New(ErrWrongAction))
		}
	}

	if len(errs) > 0 {
		statusCode := 500
		switch errs[0].Error() {
		case ErrNoActionSpecified.Error():
			statusCode = 400
		case ErrWrongAction.Error():
			statusCode = 400
		case ErrTaskNotFound.Error():
			statusCode = 404
		case ErrTaskDisabledNotRunnable.Error():
			statusCode = 409
		}
		response.Write(statusCode, response.FromSnapErrors(errs), w)
		return
	}
	response.Write(204, nil, w)
}

func (s *Server) removeTaskV2(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	id := p.ByName("id")
	err := s.mt.RemoveTask(id)
	if err != nil {
		if strings.Contains(err.Error(), ErrTaskNotFound.Error()) {
			response.Write(404, response.FromError(err), w)
			return
		}
		response.Write(500, response.FromError(err), w)
		return
	}
	response.Write(204, nil, w)
}

func taskURI(host, version string, t core.Task) string {
	return fmt.Sprintf("%s://%s/%s/tasks/%s", protocolPrefix, host, version, t.ID())
}

type TaskWatchHandlerV2 struct {
	streamCount int
	alive       bool
	mChan       chan response.StreamedTaskEvent
}

func (t *TaskWatchHandlerV2) CatchCollection(m []core.Metric) {
	sm := make([]response.StreamedMetric, len(m))
	for i := range m {
		sm[i] = response.StreamedMetric{
			Namespace: m[i].Namespace().String(),
			Data:      m[i].Data(),
			Timestamp: m[i].Timestamp(),
			Tags:      m[i].Tags(),
		}
	}
	t.mChan <- response.StreamedTaskEvent{
		EventType: response.TaskWatchMetricEvent,
		Message:   "",
		Event:     sm,
	}
}

func (t *TaskWatchHandlerV2) CatchTaskStarted() {
	t.mChan <- response.StreamedTaskEvent{
		EventType: response.TaskWatchTaskStarted,
	}
}

func (t *TaskWatchHandlerV2) CatchTaskStopped() {
	t.mChan <- response.StreamedTaskEvent{
		EventType: response.TaskWatchTaskStopped,
	}
}

func (t *TaskWatchHandlerV2) CatchTaskDisabled(why string) {
	t.mChan <- response.StreamedTaskEvent{
		EventType: response.TaskWatchTaskDisabled,
		Message:   why,
	}
}
