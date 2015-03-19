package rest

import (
	"github.com/codegangsta/negroni"
	"github.com/julienschmidt/httprouter"

	"github.com/intelsdilabs/pulse/core"
	"github.com/intelsdilabs/pulse/core/cdata"
	sched "github.com/intelsdilabs/pulse/schedule"
)

type managesMetrics interface {
	MetricCatalog() []core.MetricType
	Load(string) error
}

type managesTasks interface {
	CreateTask([]core.MetricType, sched.Schedule, *cdata.ConfigDataTree, sched.Workflow, ...sched.TaskOption) (*sched.Task, sched.TaskErrors)
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
