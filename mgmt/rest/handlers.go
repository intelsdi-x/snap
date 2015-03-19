/*
   Handlers for incoming rest requests are implemented here. The handlers
   have a Server pointer receiver so the 'state' of the server can be
   injected in to a function whose signature is hard-typed to
   httrouter.Handle: https://godoc.org/github.com/julienschmidt/httprouter#Handle
*/
package rest

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/julienschmidt/httprouter"

	"github.com/intelsdilabs/pulse/core"
	"github.com/intelsdilabs/pulse/core/cdata"
	"github.com/intelsdilabs/pulse/core/ctypes"
	sched "github.com/intelsdilabs/pulse/schedule"
)

type response struct {
	Meta *responseMeta          `json:"meta"`
	Data map[string]interface{} `json:"data"`
}

type responseMeta struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

/* ------------------ begin metric handlers ------------------ */

type metricType struct {
	Ns  []string `json:"namespace"`
	Ver int      `json:"version"`
	LAT int64    `json:"last_advertised_timestamp,omitempty"`
}

func (m *metricType) Namespace() []string           { return m.Ns }
func (m *metricType) Version() int                  { return m.Ver }
func (m *metricType) LastAdvertisedTime() time.Time { return time.Unix(m.LAT, 0) }
func (m *metricType) Config() *cdata.ConfigDataNode { return nil }

func (s *Server) getMetrics(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	rmets := []metricType{}
	mets := s.mm.MetricCatalog()
	for _, m := range mets {
		rmets = append(rmets, metricType{
			Ns:  m.Namespace(),
			Ver: m.Version(),
			LAT: m.LastAdvertisedTime().Unix(),
		})
	}
	mtsmap := make(map[string]interface{})
	mtsmap["metric_types"] = rmets
	replySuccess(200, w, mtsmap)
}

func (s *Server) getMetricsFromTree(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
}

/* ------------------ begin plugin handlers ------------------ */

func (s *Server) loadPlugin(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	loadRequest := make(map[string]string)
	errCode, err := marshalBody(&loadRequest, r.Body)
	if errCode != 0 && err != nil {
		replyError(errCode, w, err)
		return
	}
	err = s.mm.Load(loadRequest["path"])
	if err != nil {
		replyError(500, w, err)
		return
	}
	replySuccess(200, w, nil)
}

func (s *Server) getPlugins(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
}

func (s *Server) getPluginsByName(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
}

func (s *Server) getPlugin(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
}

/* ------------------ begin task handlers ------------------ */

type schedule struct {
	Interval time.Duration
}

type workflow struct {
	MTs        []*metricType `json:"metric_types"`
	Publishers []string      `json:"publishers"`
}

type taskConfig map[string]map[string]interface{}

type task struct {
	ID           string     `json:"id"`
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
	// TODO (dpittman): pass in publisher keys
	wf := sched.NewWorkflow()
	sch := sched.NewSimpleSchedule(tr.Schedule.Interval)
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
	mts := make([]core.MetricType, len(tr.Workflow.MTs))
	for i, mt := range tr.Workflow.MTs {
		mts[i] = mt
	}
	task, errs := s.mt.CreateTask(mts, sch, cdtree, wf)
	if errs != nil && len(errs.Errors()) != 0 {
		var errMsg string
		for _, e := range errs.Errors() {
			errMsg = errMsg + e.Error() + " -- "
		}
		replyError(500, w, errors.New(errMsg[:len(errMsg)-4]))
		return
	}
	// set timestamp
	tr.CreationTime = task.CreationTime.Unix()
	// set task id
	tr.ID = task.ID
	// create return map
	rmap := make(map[string]interface{})
	rmap["task"] = tr
	replySuccess(200, w, rmap)
}

func (s *Server) getTasks(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
}

/* ------------------ begin helper functions ------------------ */

func replyError(code int, w http.ResponseWriter, err error) {
	w.WriteHeader(code)
	resp := &response{
		Meta: &responseMeta{
			Code:    code,
			Message: err.Error(),
		},
	}
	jerr, _ := json.MarshalIndent(resp, "", "  ")
	fmt.Fprint(w, string(jerr))
}

func replySuccess(code int, w http.ResponseWriter, data map[string]interface{}) {
	w.WriteHeader(code)
	resp := &response{
		Meta: &responseMeta{
			Code: code,
		},
		Data: data,
	}
	j, err := json.MarshalIndent(resp, "", "  ")
	if err != nil {
		replyError(500, w, err)
		return
	}
	fmt.Fprint(w, string(j))
}

func marshalBody(in interface{}, body io.ReadCloser) (int, error) {
	b, err := ioutil.ReadAll(body)
	if err != nil {
		return 500, err
	}
	err = json.Unmarshal(b, in)
	if err != nil {
		return 400, err
	}
	return 0, nil
}
