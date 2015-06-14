package rest

import (
	"errors"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/julienschmidt/httprouter"

	"github.com/intelsdi-x/pulse/core"
	cschedule "github.com/intelsdi-x/pulse/pkg/schedule"
	"github.com/intelsdi-x/pulse/scheduler/wmap"
)

type schedule struct {
	Type     string `json:"type"`
	Interval string `json:"interval"`
}

type configItem struct {
	Key   string      `json:"key"`
	Value interface{} `json:"value"`
}

type task struct {
	ID uint64 `json:"id"`
	// Config       map[string][]configItem `json:"config"`
	Deadline     string            `json:"deadline"`
	Workflow     *wmap.WorkflowMap `json:"workflow"`
	Schedule     schedule          `json:"schedule"`
	CreationTime *time.Time        `json:"creation_timestamp,omitempty"`
	LastRunTime  *time.Time        `json:"last_run_timestamp,omitempty"`
	HitCount     uint              `json:"hit_count,omitempty"`
	MissCount    uint              `json:"miss_count,omitempty"`
	State        string            `json:"task_state"`
}

func (s *Server) addTask(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {

	tr, err := marshalTask(r.Body)
	if tr == nil {
		replyError(400, w, err)
		return
	}

	sch, err := makeSchedule(tr.Schedule)
	if err != nil {
		replyError(400, w, err)
		return
	}

	var opts []core.TaskOption
	if tr.Deadline != "" {
		dl, err := time.ParseDuration(tr.Deadline)
		if err != nil {
			replyError(400, w, err)
			return
		}
		opts = append(opts, core.TaskDeadlineDuration(dl))
	}

	task, errs := s.mt.CreateTask(sch, tr.Workflow, opts...)
	if errs != nil && len(errs.Errors()) != 0 {
		var errMsg string
		for _, e := range errs.Errors() {
			errMsg = errMsg + e.Error() + " -- "
		}
		replyError(500, w, errors.New(errMsg[:len(errMsg)-4]))
		return
	}

	// set timestamp
	tr.CreationTime = task.CreationTime()

	// set task id
	tr.ID = task.ID()

	// create return map
	rmap := make(map[string]interface{})
	rmap["task"] = tr
	replySuccess(200, w, rmap)
}

func (s *Server) getTasks(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	rmap := make(map[string]interface{})
	sts := s.mt.GetTasks()
	rts := make([]task, len(sts))
	i := 0
	for _, t := range sts {
		rts[i] = task{
			ID:           t.ID(),
			Deadline:     t.DeadlineDuration().String(),
			CreationTime: t.CreationTime(),
			LastRunTime:  t.LastRunTime(),
			HitCount:     t.HitCount(),
			MissCount:    t.MissedCount(),
			State:        t.State().String(),
		}
		i++
	}
	rmap["tasks"] = rts
	replySuccess(200, w, rmap)
}

func (s *Server) startTask(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	id, err := strconv.ParseUint(p.ByName("id"), 0, 64)
	if err != nil {
		replyError(500, w, err)
		return
	}
	err = s.mt.StartTask(id)
	if err != nil {
		replyError(404, w, err)
		return
	}

	replySuccess(200, w, nil)
}

func (s *Server) stopTask(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	id, err := strconv.ParseUint(p.ByName("id"), 0, 64)
	if err != nil {
		replyError(500, w, err)
		return
	}
	err = s.mt.StopTask(id)
	if err != nil {
		replyError(404, w, err)
		return
	}

	replySuccess(200, w, nil)
}

func marshalTask(body io.ReadCloser) (*task, error) {
	var tr task
	errCode, err := marshalBody(&tr, body)
	if errCode != 0 && err != nil {
		return nil, err
	}
	return &tr, nil
}

func makeSchedule(s schedule) (cschedule.Schedule, error) {
	var sch cschedule.Schedule
	switch s.Type {
	case "simple":
		d, err := time.ParseDuration(s.Interval)
		if err != nil {
			return nil, err
		}
		sch = cschedule.NewSimpleSchedule(d)
	default:
		return nil, errors.New("invalid schedule type: " + s.Type)
	}
	err := sch.Validate()
	if err != nil {
		return nil, err
	}
	return sch, nil
}
