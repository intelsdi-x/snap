/*
http://www.apache.org/licenses/LICENSE-2.0.txt


Copyright 2017 Intel Corporation

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

package v2

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/intelsdi-x/snap/core"
	"github.com/julienschmidt/httprouter"
)

const (
	// Event types for task watcher streaming
	TaskWatchStreamOpen   = "stream-open"
	TaskWatchMetricEvent  = "metric-event"
	TaskWatchTaskDisabled = "task-disabled"
	TaskWatchTaskStarted  = "task-started"
	TaskWatchTaskStopped  = "task-stopped"
	TaskWatchTaskEnded    = "task-ended"
)

// The amount of time to buffer streaming events before flushing in seconds
var StreamingBufferWindow = 0.1

func (s *apiV2) watchTask(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	s.wg.Add(1)
	defer s.wg.Done()

	id := p.ByName("id")

	tw := &TaskWatchHandler{
		alive: true,
		mChan: make(chan StreamedTaskEvent),
	}
	tc, err1 := s.taskManager.WatchTask(id, tw)
	if err1 != nil {
		if strings.Contains(err1.Error(), ErrTaskNotFound) {
			Write(404, FromError(err1), w)
			return
		}
		Write(500, FromError(err1), w)
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
		Write(500, FromError(ErrStreamingUnsupported), w)
		return
	}
	// send initial stream open event
	so := StreamedTaskEvent{
		EventType: TaskWatchStreamOpen,
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
			case TaskWatchMetricEvent, TaskWatchTaskStarted:
				// The client can decide to stop receiving on the stream on Task Stopped.
				// We write the event to the buffer
				fmt.Fprintf(w, "data: %s\n\n", e.ToJSON())
			case TaskWatchTaskDisabled, TaskWatchTaskStopped, TaskWatchTaskEnded:
				// A disabled task should end the streaming and close the connection
				fmt.Fprintf(w, "data: %s\n\n", e.ToJSON())
				// Flush since we are sending nothing new
				flusher.Flush()
				// Close out watcher removing it from the scheduler
				tc.Close()
				// exit since this client is no longer listening
				Write(204, nil, w)
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
			Write(204, nil, w)
			return
		case <-s.killChan:
			// Flush since we are sending nothing new
			flusher.Flush()
			// Close out watcher removing it from the scheduler
			tc.Close()
			// exit since this client is no longer listening
			Write(204, nil, w)
			return
		}
	}
}

type TaskWatchHandler struct {
	streamCount int
	alive       bool
	mChan       chan StreamedTaskEvent
}

func (t *TaskWatchHandler) CatchCollection(m []core.Metric) {
	sm := make([]StreamedMetric, len(m))
	for i := range m {
		sm[i] = StreamedMetric{
			Namespace: m[i].Namespace().String(),
			Data:      m[i].Data(),
			Timestamp: m[i].Timestamp(),
			Tags:      m[i].Tags(),
		}
	}
	t.mChan <- StreamedTaskEvent{
		EventType: TaskWatchMetricEvent,
		Message:   "",
		Event:     sm,
	}
}

func (t *TaskWatchHandler) CatchTaskStarted() {
	t.mChan <- StreamedTaskEvent{
		EventType: TaskWatchTaskStarted,
	}
}

func (t *TaskWatchHandler) CatchTaskStopped() {
	t.mChan <- StreamedTaskEvent{
		EventType: TaskWatchTaskStopped,
	}
}

func (t *TaskWatchHandler) CatchTaskEnded() {
	t.mChan <- StreamedTaskEvent{
		EventType: TaskWatchTaskEnded,
	}
}

func (t *TaskWatchHandler) CatchTaskDisabled(why string) {
	t.mChan <- StreamedTaskEvent{
		EventType: TaskWatchTaskDisabled,
		Message:   why,
	}
}

// TaskWatchResponse defines the response of the task watching stream.
//
// swagger:response TaskWatchResponse
type TaskWatchResponse struct {
	// in: body
	Body struct {
		TaskWatch StreamedTaskEvent `json:"task_watch"`
	}
}

// StreamedTaskEvent defines the task watching data type.
type StreamedTaskEvent struct {
	EventType string          `json:"type"`
	Message   string          `json:"message"`
	Event     StreamedMetrics `json:"event,omitempty"`
}

func (s *StreamedTaskEvent) ToJSON() string {
	j, _ := json.Marshal(s)
	return string(j)
}

type StreamedMetric struct {
	Namespace string            `json:"namespace"`
	Data      interface{}       `json:"data"`
	Timestamp time.Time         `json:"timestamp"`
	Tags      map[string]string `json:"tags"`
}

// StreamedMetrics defines a slice of streamed metrics.
type StreamedMetrics []StreamedMetric

func (s StreamedMetrics) Len() int {
	return len(s)
}

func (s StreamedMetrics) Less(i, j int) bool {
	return fmt.Sprintf("%s", s[i].Namespace) < fmt.Sprintf("%s", s[j].Namespace)
}

func (s StreamedMetrics) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}
