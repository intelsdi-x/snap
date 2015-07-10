package rest

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strconv"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/julienschmidt/httprouter"

	"github.com/intelsdi-x/pulse/core"
	"github.com/intelsdi-x/pulse/mgmt/rest/rbody"
	"github.com/intelsdi-x/pulse/mgmt/rest/request"
	cschedule "github.com/intelsdi-x/pulse/pkg/schedule"
	"github.com/intelsdi-x/pulse/scheduler/wmap"
)

var (
	// The amount of time to buffer streaming events before flushing in seconds
	StreamingBufferWindow = 0.1

	ErrStreamingUnsupported = errors.New("Streaming unsupported")
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

	sch, err := makeSchedule(tr.Schedule)
	if err != nil {
		respond(500, rbody.FromError(err), w)
		return
	}

	var opts []core.TaskOption
	if tr.Deadline != "" {
		dl, err := time.ParseDuration(tr.Deadline)
		if err != nil {
			respond(500, rbody.FromError(err), w)
			return
		}
		opts = append(opts, core.TaskDeadlineDuration(dl))
	}

	if tr.Name != "" {
		opts = append(opts, core.SetTaskName(tr.Name))
	}
	opts = append(opts, core.OptionStopOnFailure(10))

	task, errs := s.mt.CreateTask(sch, tr.Workflow, tr.Start, opts...)
	if errs != nil && len(errs.Errors()) != 0 {
		var errMsg string
		for _, e := range errs.Errors() {
			errMsg = errMsg + e.Error() + " -- "
		}
		respond(500, rbody.FromError(errors.New(errMsg[:len(errMsg)-4])), w)
		return
	}

	taskB := rbody.AddSchedulerTaskFromTask(task)
	respond(201, taskB, w)
}

func (s *Server) getTasks(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	sts := s.mt.GetTasks()

	tasks := &rbody.ScheduledTaskListReturned{}
	tasks.ScheduledTasks = make([]rbody.ScheduledTask, len(sts))

	i := 0
	for _, t := range sts {
		tasks.ScheduledTasks[i] = *rbody.SchedulerTaskFromTask(t)
		i++
	}
	sort.Sort(tasks)
	respond(200, tasks, w)
}

func (s *Server) getTask(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	id, err := strconv.ParseUint(p.ByName("id"), 0, 64)
	if err != nil {
		respond(500, rbody.FromError(err), w)
		return
	}
	t, err1 := s.mt.GetTask(id)
	if err1 != nil {
		respond(404, rbody.FromError(err1), w)
		return
	}
	task := &rbody.ScheduledTaskReturned{}
	task.AddScheduledTask = *rbody.AddSchedulerTaskFromTask(t)
	respond(200, task, w)
}

func (s *Server) watchTask(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	logger := log.WithFields(log.Fields{
		"_module": "api",
		"_block":  "watch-task",
		"client":  r.RemoteAddr,
	})

	id, err := strconv.ParseUint(p.ByName("id"), 0, 64)
	if err != nil {
		respond(500, rbody.FromError(err), w)
		return
	}
	logger.WithFields(log.Fields{
		"task-id": id,
	}).Debug("request to watch task")
	tw := &TaskWatchHandler{
		alive: true,
		mChan: make(chan rbody.StreamedTaskEvent),
	}
	tc, err1 := s.mt.WatchTask(id, tw)
	if err1 != nil {
		respond(404, rbody.FromError(err1), w)
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
		EventType: rbody.TaskWatchStreamOpen,
		Message:   "Stream opended",
	}
	fmt.Fprintf(w, "%s\n", so.ToJSON())
	flusher.Flush()

	// Get a channel for if the client notifies us it is closing the connection
	n := w.(http.CloseNotifier).CloseNotify()
	t := time.Now()
	for {
		// Write to the ResponseWriter
		select {
		case e := <-tw.mChan:
			logger.WithFields(log.Fields{
				"task-id":            id,
				"task-watcher-event": e.EventType,
			}).Debug("new event")
			switch e.EventType {
			case rbody.TaskWatchMetricEvent, rbody.TaskWatchTaskStarted, rbody.TaskWatchTaskStopped:
				// The client can decide to stop receiving on the stream on Task Stopped.
				// We write the event to the buffer
				fmt.Fprintf(w, "%s\n", e.ToJSON())
			case rbody.TaskWatchTaskDisabled:
				// A disabled task should end the streaming and close the connection
				fmt.Fprintf(w, "%s\n", e.ToJSON())
				// Flush since we are sending nothing new
				flusher.Flush()
				// Close out watcher removing it from the scheduler
				tc.Close()
				// exit since this client is no longer listening
				respond(200, &rbody.ScheduledTaskWatchingEnded{}, w)
			}
			// If we are at least above our minimum buffer time we flush to send
			if time.Now().Sub(t).Seconds() > StreamingBufferWindow {
				flusher.Flush()
				t = time.Now()
			}
		case <-n:
			logger.WithFields(log.Fields{
				"task-id": id,
			}).Debug("client disconnecting")
			// Flush since we are sending nothing new
			flusher.Flush()
			// Close out watcher removing it from the scheduler
			tc.Close()
			// exit since this client is no longer listening
			respond(200, &rbody.ScheduledTaskWatchingEnded{}, w)
		}

	}
}

func (s *Server) startTask(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	id, err := strconv.ParseUint(p.ByName("id"), 0, 64)
	if err != nil {
		respond(500, rbody.FromError(err), w)
		return
	}
	errs := s.mt.StartTask(id)
	if errs != nil {
		respond(404, rbody.FromPulseErrors(errs), w)
		return
	}
	// TODO should return resource
	respond(200, &rbody.ScheduledTaskStarted{ID: int(id)}, w)
}

func (s *Server) stopTask(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	id, err := strconv.ParseUint(p.ByName("id"), 0, 64)
	if err != nil {
		respond(500, rbody.FromError(err), w)
		return
	}
	err = s.mt.StopTask(id)
	if err != nil {
		respond(404, rbody.FromError(err), w)
		return
	}
	respond(200, &rbody.ScheduledTaskStopped{ID: int(id)}, w)
}

func (s *Server) removeTask(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	id, err := strconv.ParseUint(p.ByName("id"), 0, 64)
	if err != nil {
		respond(500, rbody.FromError(err), w)
		return
	}
	err = s.mt.RemoveTask(id)
	if err != nil {
		respond(404, rbody.FromError(err), w)
		return
	}
	respond(200, &rbody.ScheduledTaskRemoved{ID: int(id)}, w)
}

func marshalTask(body io.ReadCloser) (*request.TaskCreationRequest, error) {
	var tr request.TaskCreationRequest
	errCode, err := marshalBody(&tr, body)
	if errCode != 0 && err != nil {
		return nil, err
	}
	return &tr, nil
}

func makeSchedule(s request.Schedule) (cschedule.Schedule, error) {
	switch s.Type {
	case "simple":
		d, err := time.ParseDuration(s.Interval)
		if err != nil {
			return nil, err
		}
		sch := cschedule.NewSimpleSchedule(d)

		err = sch.Validate()
		if err != nil {
			return nil, err
		}
		return sch, nil
	case "windowed":
		d, err := time.ParseDuration(s.Interval)
		if err != nil {
			return nil, err
		}

		var start, stop *time.Time
		if s.StartTimestamp != nil {
			t := time.Unix(*s.StartTimestamp, 0)
			start = &t
		}
		if s.StopTimestamp != nil {
			t := time.Unix(*s.StopTimestamp, 0)
			stop = &t
		}
		sch := cschedule.NewWindowedSchedule(
			d,
			start,
			stop,
		)

		err = sch.Validate()
		if err != nil {
			return nil, err
		}
		return sch, nil
	default:
		return nil, errors.New("unknown schedule type " + s.Type)
	}
}

type TaskWatchHandler struct {
	streamCount int
	alive       bool
	mChan       chan rbody.StreamedTaskEvent
}

func (t *TaskWatchHandler) CatchCollection(m []core.Metric) {
	sm := make([]rbody.StreamedMetric, len(m))
	for i, _ := range m {
		sm[i] = rbody.StreamedMetric{
			Namespace: joinNamespace(m[i].Namespace()),
			Data:      m[i].Data(),
		}
	}
	t.mChan <- rbody.StreamedTaskEvent{
		EventType: rbody.TaskWatchMetricEvent,
		Message:   "",
		Event:     sm,
	}
}

func (t *TaskWatchHandler) CatchTaskStarted() {
	t.mChan <- rbody.StreamedTaskEvent{
		EventType: rbody.TaskWatchTaskStarted,
	}
}

func (t *TaskWatchHandler) CatchTaskStopped() {
	t.mChan <- rbody.StreamedTaskEvent{
		EventType: rbody.TaskWatchTaskStopped,
	}
}

func (t *TaskWatchHandler) CatchTaskDisabled(why string) {
	t.mChan <- rbody.StreamedTaskEvent{
		EventType: rbody.TaskWatchTaskDisabled,
		Message:   why,
	}
}
