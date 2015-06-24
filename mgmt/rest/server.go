package rest

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"

	log "github.com/Sirupsen/logrus"
	"github.com/codegangsta/negroni"
	"github.com/julienschmidt/httprouter"

	"github.com/intelsdi-x/pulse/core"
	cschedule "github.com/intelsdi-x/pulse/pkg/schedule"
	"github.com/intelsdi-x/pulse/scheduler/wmap"
)

type managesMetrics interface {
	MetricCatalog() ([]core.Metric, error)
	Load(string) error
	PluginCatalog() core.PluginCatalog
	AvailablePlugins() []core.AvailablePlugin
	GetAutodiscoverPaths() []string
}

type managesTasks interface {
	CreateTask(cschedule.Schedule, *wmap.WorkflowMap, ...core.TaskOption) (core.Task, core.TaskErrors)
	GetTasks() map[uint64]core.Task
	StartTask(id uint64) error
	StopTask(id uint64) error
	RemoveTask(id uint64) error
}

type Server struct {
	mm managesMetrics
	mt managesTasks
	n  *negroni.Negroni
	r  *httprouter.Router
}

func New() *Server {

	n := negroni.New(
		NewLogger(),
		// TODO a recovery logger
	)
	return &Server{
		r: httprouter.New(),
		n: n,
	}

}

func (s *Server) Start(addrString string) {
	go s.start(addrString)
}

func (s *Server) run(addrString string) {
	log.Printf("[pulse-rest] listening on %s\n", addrString)
	http.ListenAndServe(addrString, s.n)
}

func (s *Server) BindMetricManager(m managesMetrics) {
	s.mm = m
}

func (s *Server) BindTaskManager(t managesTasks) {
	s.mt = t
}

func (s *Server) start(addrString string) {
	// plugin routes
	s.r.GET("/v1/plugins", s.getPlugins)
	s.r.GET("/v1/plugins/:name", s.getPluginsByName)
	s.r.GET("/v1/plugins/:name/:version", s.getPlugin)
	s.r.POST("/v1/plugins", s.loadPlugin)

	// metric routes
	s.r.GET("/v1/metrics", s.getMetrics)
	s.r.GET("/v1/metrics/*namespace", s.getMetricsFromTree)

	// task routes
	s.r.GET("/v1/tasks", s.getTasks)
	s.r.POST("/v1/tasks", s.addTask)
	s.r.PUT("/v1/tasks/:id/start", s.startTask)
	s.r.PUT("/v1/tasks/:id/stop", s.stopTask)
	s.r.DELETE("/v1/tasks/:id", s.removeTask)

	// set negroni router to the server's router (httprouter)
	s.n.UseHandler(s.r)
	// start http handling
	s.run(addrString)
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

func parseNamespace(ns string) []string {
	if strings.Index(ns, "/") == 0 {
		return strings.Split(ns[1:], "/")
	}
	return strings.Split(ns, "/")
}

func joinNamespace(ns []string) string {
	return "/" + strings.Join(ns, "/")
}
