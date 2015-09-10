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
	"github.com/intelsdi-x/pulse/core/perror"
	"github.com/intelsdi-x/pulse/mgmt/rest/rbody"
	cschedule "github.com/intelsdi-x/pulse/pkg/schedule"
	"github.com/intelsdi-x/pulse/scheduler/wmap"
)

/*
REST API (ALPHA, It may CHANGE!)

module specific date or error <= internal call

REST API response

Response (JSON encoded)
	meta <Response Meta (common across all responses)>
		code <HTTP response code duplciated from header>
		body_type <keyword for structure type in body>
		version <API version for future version switching>
	body <Response Body>
		<Happy Path>
		Action specific version of struct matching type in Meta.Type. Should include URI if returning a resource or collection of rsources.
		Type should be exposed off rest package for use by clients.
		<Unhappy Path>
		Generic error type with optional fields. Normally converted from perror.PulseError interface types
*/

const (
	APIVersion = 1
)

var (
	restLogger = log.WithField("_module", "_mgmt-rest")
)

type APIResponse struct {
	Meta         *APIResponseMeta `json:"meta"`
	Body         rbody.Body       `json:"body"`
	JSONResponse string           `json:"-"`
}

type apiResponseJSON struct {
	Meta *APIResponseMeta `json:"meta"`
	Body json.RawMessage  `json:"body"`
}

type APIResponseMeta struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Type    string `json:"type"`
	Version int    `json:"version"`
}

type managesMetrics interface {
	MetricCatalog() ([]core.CatalogedMetric, error)
	FetchMetrics([]string, int) ([]core.CatalogedMetric, error)
	GetMetric([]string, int) (core.CatalogedMetric, error)
	Load(string) (core.CatalogedPlugin, perror.PulseError)
	Unload(core.Plugin) (core.CatalogedPlugin, perror.PulseError)
	PluginCatalog() core.PluginCatalog
	AvailablePlugins() []core.AvailablePlugin
	GetAutodiscoverPaths() []string
}

type managesTasks interface {
	CreateTask(cschedule.Schedule, *wmap.WorkflowMap, bool, ...core.TaskOption) (core.Task, core.TaskErrors)
	GetTasks() map[uint64]core.Task
	GetTask(uint64) (core.Task, error)
	StartTask(uint64) []perror.PulseError
	StopTask(uint64) []perror.PulseError
	RemoveTask(uint64) error
	WatchTask(uint64, core.TaskWatcherHandler) (core.TaskWatcherCloser, error)
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
		negroni.NewRecovery(),
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
	s.r.GET("/v1/plugins/:type", s.getPluginsByType)
	s.r.GET("/v1/plugins/:type/:name", s.getPluginsByName)
	s.r.GET("/v1/plugins/:type/:name/:version", s.getPlugin)
	s.r.POST("/v1/plugins", s.loadPlugin)
	s.r.DELETE("/v1/plugins/:type/:name/:version", s.unloadPlugin)

	// metric routes
	s.r.GET("/v1/metrics", s.getMetrics)
	s.r.GET("/v1/metrics/*namespace", s.getMetricsFromTree)

	// task routes
	s.r.GET("/v1/tasks", s.getTasks)
	s.r.GET("/v1/tasks/:id", s.getTask)
	s.r.GET("/v1/tasks/:id/watch", s.watchTask)
	s.r.POST("/v1/tasks", s.addTask)
	s.r.PUT("/v1/tasks/:id/start", s.startTask)
	s.r.PUT("/v1/tasks/:id/stop", s.stopTask)
	s.r.DELETE("/v1/tasks/:id", s.removeTask)

	// set negroni router to the server's router (httprouter)
	s.n.UseHandler(s.r)
	// start http handling
	s.run(addrString)
}

func respond(code int, b rbody.Body, w http.ResponseWriter) {
	resp := &APIResponse{
		Meta: &APIResponseMeta{
			Code:    code,
			Message: b.ResponseBodyMessage(),
			Type:    b.ResponseBodyType(),
			Version: APIVersion,
		},
		Body: b,
	}

	w.WriteHeader(code)

	j, err := json.MarshalIndent(resp, "", "  ")
	if err != nil {
		panic(err)
	}
	fmt.Fprint(w, string(j))
}

func (a *APIResponse) UnmarshalJSON(b []byte) error {
	ar := &apiResponseJSON{}
	err := json.Unmarshal(b, ar)
	if err != nil {
		panic(err)
	}
	body, err := rbody.UnmarshalBody(ar.Meta.Type, ar.Body)
	if err != nil {
		return err
	}
	// Assign
	a.Meta = ar.Meta
	a.Body = body
	return nil
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
		ns = ns[1:]
	}
	if ns[len(ns)-1] == '/' {
		ns = ns[:len(ns)-1]
	}
	return strings.Split(ns, "/")
}
