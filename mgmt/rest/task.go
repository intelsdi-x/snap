package rest

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/julienschmidt/httprouter"

	"github.com/intelsdilabs/pulse/core"
	"github.com/intelsdilabs/pulse/core/cdata"
	"github.com/intelsdilabs/pulse/core/ctypes"
)

type schedule struct {
	Interval time.Duration
}

type workflow struct {
	St core.WorkflowState `json:"state,omitempty"`
	Mp core.WfMap         `json:"map"`
}

func (w *workflow) State() core.WorkflowState {
	return w.St
}

func (w *workflow) Map() core.WfMap {
	return w.Mp
}

type taskConfig map[string]map[string]interface{}

type task struct {
	ID           uint64     `json:"id"`
	Config       taskConfig `json:"config"`
	Workflow     *workflow  `json:"workflow"`
	Schedule     *schedule  `json:"schedule"`
	CreationTime int64      `json:"creation_timestamp"`
	LastRunTime  int64      `json:"last_run_time,omitempty"`
	HitCount     uint       `json:"hit_count,omitempty"`
	MissCount    uint       `json:"miss_count,omitempty"`
}

func (s *Server) addTask(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	var tr task
	errCode, err := marshalBody(&tr, r.Body)
	if errCode != 0 && err != nil {
		replyError(errCode, w, err)
		return
	}
	// create workflow
	sch := core.NewSimpleSchedule(tr.Schedule.Interval)
	// build config data tree
	cdtree := cdata.NewTree()
	// walk through config items
	// ns = namespace ct = config table
	for ns, ct := range tr.Config {
		config := make(map[string]ctypes.ConfigValue)
		// walk through key and value for a given namespace
		for key, val := range ct {
			// assert type and insert into a table (config)
			switch v := val.(type) {
			case int:
				config[key] = ctypes.ConfigValueInt{
					Value: v,
				}
			case string:
				config[key] = ctypes.ConfigValueStr{
					Value: v,
				}
			case float64:
				config[key] = ctypes.ConfigValueFloat{
					Value: v,
				}
			default:
				replyError(500, w, errors.New("unsupported type for config data key "+key))
				return
			}
		}
		// create config data  node
		cdn := cdata.FromTable(config)
		// namespace creation for adding cdn
		// if given name begins with /, we splice it off
		// so the first element in namespace slice is not empty string.
		var nss []string
		if strings.Index(ns, "/") == 0 {
			nss = strings.Split(ns[1:], "/")
		} else {
			nss = strings.Split(ns, "/")
		}
		cdtree.Add(nss, cdn)
	}
	// cast metricTypes to core.MetricType
	task, errs := s.mt.CreateTask(tr.Workflow.Mp.Collect.MetricTypes, sch, cdtree, tr.Workflow)
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
