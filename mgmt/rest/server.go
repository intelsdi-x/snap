package rest

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/codegangsta/negroni"
	"github.com/julienschmidt/httprouter"

	"github.com/intelsdilabs/pulse/core"
	"github.com/intelsdilabs/pulse/core/cdata"
)

type managesMetrics interface {
	MetricCatalog() ([]core.MetricType, error)
	Load(string) error
}

type managesTasks interface {
	CreateTask([]core.MetricType, core.Schedule, *cdata.ConfigDataTree, core.Workflow, ...core.TaskOption) (core.Task, core.TaskErrors)
}

type Server struct {
	mm managesMetrics
	mt managesTasks
	n  *negroni.Negroni
	r  *httprouter.Router
}

func New() *Server {
	return &Server{
		n: negroni.Classic(),
		r: httprouter.New(),
	}
}

func (s *Server) Start(addrString string) {
	go s.start(addrString)
}

func (s *Server) BindMetricManager(m managesMetrics) {
	s.mm = m
}

func (s *Server) BindTaskManager(t managesTasks) {
	s.mt = t
}

func (s *Server) start(addrString string) {

	// plugin routes
	s.r.GET("/v1/plugin", s.getPlugins)
	s.r.GET("/v1/plugin/:name", s.getPluginsByName)
	s.r.GET("/v1/plugin/:name/:version", s.getPlugin)
	s.r.POST("/v1/plugin", s.loadPlugin)

	// metric routes
	s.r.GET("/v1/metric", s.getMetrics)
	s.r.GET("/v1/metric/*namespace", s.getMetricsFromTree)

	// task routes
	s.r.GET("/v1/task", s.getTasks)
	s.r.POST("/v1/task", s.addTask)

	// set negroni router to the server's router (httprouter)
	s.n.UseHandler(s.r)
	// start http handling
	s.n.Run(addrString)
}

type response struct {
	Meta *responseMeta          `json:"meta"`
	Data map[string]interface{} `json:"data"`
}

type responseMeta struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

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

func parsens(ns string) []string {
	if strings.Index(ns, "/") == 0 {
		return strings.Split(ns[1:], "/")
	} else {
		return strings.Split(ns, "/")
	}
}

func joinns(ns []string) string {
	return "/" + strings.Join(ns, "/")
}
