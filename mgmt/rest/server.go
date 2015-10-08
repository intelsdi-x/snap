package rest

import (
	"encoding/json"
	"errors"
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
	"github.com/intelsdi-x/pulse/mgmt/tribe/agreement"
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
	ErrBadCert = errors.New("Invalid certificate given")

	restLogger = log.WithField("_module", "_mgmt-rest")
)

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
	GetTasks() map[string]core.Task
	GetTask(string) (core.Task, error)
	StartTask(string) []perror.PulseError
	StopTask(string) []perror.PulseError
	RemoveTask(string) error
	WatchTask(string, core.TaskWatcherHandler) (core.TaskWatcherCloser, error)
	EnableTask(string) (core.Task, error)
}

type managesTribe interface {
	GetAgreement(name string) (*agreement.Agreement, perror.PulseError)
	GetAgreements() map[string]*agreement.Agreement
	AddAgreement(name string) perror.PulseError
	JoinAgreement(agreementName, memberName string) perror.PulseError
	GetMembers() []string
	GetMember(name string) *agreement.Member
}

type Server struct {
	mm  managesMetrics
	mt  managesTasks
	tr  managesTribe
	n   *negroni.Negroni
	r   *httprouter.Router
	tls *tls
}

func New(https bool, cpath, kpath string) (*Server, error) {
	s := &Server{}

	if https {
		var err error
		s.tls, err = newtls(cpath, kpath)
		if err != nil {
			return nil, err
		}
	}

	restLogger.Info(fmt.Sprintf("Loading REST API with HTTPS set to: %v", https))

	s.n = negroni.New(
		NewLogger(),
		negroni.NewRecovery(),
	)
	s.r = httprouter.New()
	return s, nil
}

func (s *Server) Start(addrString string) {
	go s.start(addrString)
}

func (s *Server) run(addrString string) {
	log.Printf("[pulse-rest] listening on %s\n", addrString)
	if s.tls != nil {
		err := http.ListenAndServeTLS(addrString, s.tls.cert, s.tls.key, s.n)
		if err != nil {
			panic(err)
		}
	} else {
		http.ListenAndServe(addrString, s.n)
	}
}

func (s *Server) BindMetricManager(m managesMetrics) {
	s.mm = m
}

func (s *Server) BindTaskManager(t managesTasks) {
	s.mt = t
}

func (s *Server) BindTribeManager(t managesTribe) {
	s.tr = t
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
	s.r.PUT("/v1/tasks/:id/enable", s.enableTask)

	if s.tr != nil {
		s.r.GET("/v1/tribe/agreements", s.getAgreements)
		s.r.POST("/v1/tribe/agreements", s.addAgreement)
		s.r.GET("/v1/tribe/agreements/:name", s.getAgreement)
		s.r.POST("/v1/tribe/agreements/:name/join", s.joinAgreement)
		s.r.GET("/v1/tribe/members", s.getMembers)
		s.r.GET("/v1/tribe/member/:name", s.getMember)
	}

	// set negroni router to the server's router (httprouter)
	s.n.UseHandler(s.r)
	// start http handling
	s.run(addrString)
}

func respond(code int, b rbody.Body, w http.ResponseWriter) {
	resp := &rbody.APIResponse{
		Meta: &rbody.APIResponseMeta{
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
