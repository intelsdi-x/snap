package rest

import (
	"errors"
	"io"
	"net/http"
	"sort"
	"strconv"
	"time"

	"github.com/julienschmidt/httprouter"

	"github.com/intelsdi-x/pulse/core"
	"github.com/intelsdi-x/pulse/mgmt/rest/rbody"
	"github.com/intelsdi-x/pulse/mgmt/rest/request"
	cschedule "github.com/intelsdi-x/pulse/pkg/schedule"
	"github.com/intelsdi-x/pulse/scheduler/wmap"
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

	task, errs := s.mt.CreateTask(sch, tr.Workflow, opts...)
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
	// rmap := make(map[string]interface{})
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

func (s *Server) startTask(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	id, err := strconv.ParseUint(p.ByName("id"), 0, 64)
	if err != nil {
		respond(500, rbody.FromError(err), w)
		return
	}
	err = s.mt.StartTask(id)
	if err != nil {
		respond(404, rbody.FromError(err), w)
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
	var sch cschedule.Schedule
	switch s.Type {
	case "simple":
		d, err := time.ParseDuration(s.Interval)
		if err != nil {
			return nil, err
		}
		sch = cschedule.NewSimpleSchedule(d)
	default:
		return nil, errors.New("unknown schedule type " + s.Type)
	}
	err := sch.Validate()
	if err != nil {
		return nil, err
	}
	return sch, nil
}
