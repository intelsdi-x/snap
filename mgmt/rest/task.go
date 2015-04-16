package rest

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"time"

	"github.com/julienschmidt/httprouter"

	"github.com/intelsdilabs/pulse/core"
	"github.com/intelsdilabs/pulse/core/cdata"
	"github.com/intelsdilabs/pulse/core/ctypes"
)

type schedule struct {
	Type     string `json:"type"`
	Interval string `json:"interval"`
}

type collect struct {
	MetricTypes []*metricType `json:"metric_types"`
	Process     []process     `json:"process"`
	Publish     []publish     `json:"publish"`
}

type process struct {
	Plugin  plugin    `json:"plugin"`
	Publish []publish `json:"publish"`
}

type publish struct {
	Plugin plugin `json:"plugin"`
}

type workflow struct {
	St      core.WorkflowState `json:"state,omitempty"`
	Collect collect            `json:"collect"`
}

func (w *workflow) Marshal() ([]byte, error) {
	j, err := json.Marshal(w)
	if err != nil {
		return nil, err
	}
	return j, nil
}

func (w *workflow) Unmarshal(j []byte) error {
	err := json.Unmarshal(j, &w)
	if err != nil {
		return err
	}
	return nil
}

func (w *workflow) State() core.WorkflowState {
	return w.St
}

type configItem struct {
	Key   string      `json:"key"`
	Value interface{} `json:"value"`
}

type task struct {
	ID           uint64                  `json:"id"`
	Config       map[string][]configItem `json:"config"`
	Deadline     string                  `json:"deadline"`
	Workflow     *workflow               `json:"workflow"`
	Schedule     schedule                `json:"schedule"`
	CreationTime int64                   `json:"creation_timestamp"`
	LastRunTime  int64                   `json:"last_run_time,omitempty"`
	HitCount     uint                    `json:"hit_count,omitempty"`
	MissCount    uint                    `json:"miss_count,omitempty"`
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

	cdtree, err := assembleCDTree(tr.Config)
	if err != nil {
		replyError(400, w, err)
		return
	}

	mts := make([]core.MetricType, len(tr.Workflow.Collect.MetricTypes))
	for i, m := range tr.Workflow.Collect.MetricTypes {
		mts[i] = m
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

	task, errs := s.mt.CreateTask(mts, sch, cdtree, tr.Workflow, opts...)
	if errs != nil && len(errs.Errors()) != 0 {
		var errMsg string
		for _, e := range errs.Errors() {
			errMsg = errMsg + e.Error() + " -- "
		}
		replyError(500, w, errors.New(errMsg[:len(errMsg)-4]))
		return
	}

	// set timestamp
	tr.CreationTime = task.CreationTime().Unix()

	// set task id
	tr.ID = task.Id()

	// create return map
	rmap := make(map[string]interface{})
	rmap["task"] = tr
	replySuccess(200, w, rmap)
}

func (s *Server) getTasks(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
}

func marshalTask(body io.ReadCloser) (*task, error) {
	var tr task
	errCode, err := marshalBody(&tr, body)
	if errCode != 0 && err != nil {
		return nil, err
	}
	return &tr, nil
}

func makeSchedule(s schedule) (core.Schedule, error) {
	var sch core.Schedule
	switch s.Type {
	case "simple":
		d, err := time.ParseDuration(s.Interval)
		if err != nil {
			return nil, err
		}
		sch = core.NewSimpleSchedule(d)
	default:
		return nil, errors.New("invalid schedule type: " + s.Type)
	}
	err := sch.Validate()
	if err != nil {
		return nil, err
	}
	return sch, nil
}

func assembleCDTree(m map[string][]configItem) (*cdata.ConfigDataTree, error) {
	// build config data tree
	cdtree := cdata.NewTree()
	// walk through config items
	// ns = namespace ct = config table
	for ns, ct := range m {
		config := make(map[string]ctypes.ConfigValue)
		// walk through key and value for a given namespace
		for _, ci := range ct {
			// assert type and insert into a table (config)
			switch v := ci.Value.(type) {
			case int:
				config[ci.Key] = ctypes.ConfigValueInt{
					Value: v,
				}
			case string:
				config[ci.Key] = ctypes.ConfigValueStr{
					Value: v,
				}
			case float64:
				config[ci.Key] = ctypes.ConfigValueFloat{
					Value: v,
				}
			}
			// unable to assert the type of ci.Value, so return an error.
			if _, ok := config[ci.Key]; !ok {
				return nil, errors.New("unsupported type for config data key " + ci.Key)
			}
		}
		// create config data node
		cdn := cdata.FromTable(config)
		nss := parsens(ns)
		cdtree.Add(nss, cdn)
	}
	return cdtree, nil
}
