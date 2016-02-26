/*
http://www.apache.org/licenses/LICENSE-2.0.txt


Copyright 2015 Intel Corporation

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package rest

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/codegangsta/negroni"
	"github.com/julienschmidt/httprouter"

	"github.com/intelsdi-x/snap/core"
	"github.com/intelsdi-x/snap/core/cdata"
	"github.com/intelsdi-x/snap/core/serror"
	"github.com/intelsdi-x/snap/mgmt/rest/rbody"
	"github.com/intelsdi-x/snap/mgmt/tribe/agreement"
	cschedule "github.com/intelsdi-x/snap/pkg/schedule"
	"github.com/intelsdi-x/snap/scheduler/wmap"
)

const (
	APIVersion = 1
)

var (
	ErrBadCert = errors.New("Invalid certificate given")

	restLogger     = log.WithField("_module", "_mgmt-rest")
	protocolPrefix = "http"
)

type managesMetrics interface {
	MetricCatalog() ([]core.CatalogedMetric, error)
	FetchMetrics([]string, int) ([]core.CatalogedMetric, error)
	GetMetricVersions([]string) ([]core.CatalogedMetric, error)
	GetMetric([]string, int) (core.CatalogedMetric, error)
	Load(*core.RequestedPlugin) (core.CatalogedPlugin, serror.SnapError)
	Unload(core.Plugin) (core.CatalogedPlugin, serror.SnapError)
	PluginCatalog() core.PluginCatalog
	AvailablePlugins() []core.AvailablePlugin
	GetAutodiscoverPaths() []string
}

type managesTasks interface {
	CreateTask(cschedule.Schedule, *wmap.WorkflowMap, bool, ...core.TaskOption) (core.Task, core.TaskErrors)
	GetTasks() map[string]core.Task
	GetTask(string) (core.Task, error)
	StartTask(string) []serror.SnapError
	StopTask(string) []serror.SnapError
	RemoveTask(string) error
	WatchTask(string, core.TaskWatcherHandler) (core.TaskWatcherCloser, error)
	EnableTask(string) (core.Task, error)
}

type managesTribe interface {
	GetAgreement(name string) (*agreement.Agreement, serror.SnapError)
	GetAgreements() map[string]*agreement.Agreement
	AddAgreement(name string) serror.SnapError
	RemoveAgreement(name string) serror.SnapError
	JoinAgreement(agreementName, memberName string) serror.SnapError
	LeaveAgreement(agreementName, memberName string) serror.SnapError
	GetMembers() []string
	GetMember(name string) *agreement.Member
}

type managesConfig interface {
	GetPluginConfigDataNode(core.PluginType, string, int) cdata.ConfigDataNode
	GetPluginConfigDataNodeAll() cdata.ConfigDataNode
	MergePluginConfigDataNode(pluginType core.PluginType, name string, ver int, cdn *cdata.ConfigDataNode) cdata.ConfigDataNode
	MergePluginConfigDataNodeAll(cdn *cdata.ConfigDataNode) cdata.ConfigDataNode
	DeletePluginConfigDataNodeField(pluginType core.PluginType, name string, ver int, fields ...string) cdata.ConfigDataNode
	DeletePluginConfigDataNodeFieldAll(fields ...string) cdata.ConfigDataNode
}

type Server struct {
	mm      managesMetrics
	mt      managesTasks
	tr      managesTribe
	mc      managesConfig
	n       *negroni.Negroni
	r       *httprouter.Router
	tls     *tls
	auth    bool
	authpwd string
	addr    net.Addr
	err     chan error
}

func New(https bool, cpath, kpath string) (*Server, error) {
	s := &Server{
		err: make(chan error),
	}

	if https {
		var err error
		s.tls, err = newtls(cpath, kpath)
		if err != nil {
			return nil, err
		}
		protocolPrefix = "https"
	}

	restLogger.Info(fmt.Sprintf("Configuring REST API with HTTPS set to: %v", https))
	s.n = negroni.New(
		NewLogger(),
		negroni.NewRecovery(),
		negroni.HandlerFunc(s.authMiddleware),
	)
	s.r = httprouter.New()
	// Use negroni to handle routes
	s.n.UseHandler(s.r)
	return s, nil
}

// SetAPIAuth sets API authentication to enabled or disabled
func (s *Server) SetAPIAuth(auth bool) {
	s.auth = auth
}

// SetAPIAuthPwd sets the API authentication password from snapd
func (s *Server) SetAPIAuthPwd(pwd string) {
	s.authpwd = pwd
}

// Auth Middleware for REST API
func (s *Server) authMiddleware(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	defer r.Body.Close()
	if s.auth {
		_, password, ok := r.BasicAuth()
		// If we have valid password or going to tribe/agreements endpoint
		// go to next. tribe/agreements endpoint used for populating
		// snapctl help page when tribe mode is turned on.
		if ok && password == s.authpwd {
			next(rw, r)
		} else {
			http.Error(rw, "Not Authorized", 401)
		}
	} else {
		next(rw, r)
	}
}

func (s *Server) Start(addrString string) {
	s.addRoutes()
	s.run(addrString)
}

func (s *Server) Err() <-chan error {
	return s.err
}

func (s *Server) Port() int {
	return s.addr.(*net.TCPAddr).Port
}

func (s *Server) run(addrString string) {
	restLogger.Info("Starting REST API on ", addrString)
	if s.tls != nil {
		go s.serveTLS(addrString)
	} else {
		ln, err := net.Listen("tcp", addrString)
		if err != nil {
			s.err <- err
		}
		s.addr = ln.Addr()
		go s.serve(ln)
	}
}

func (s *Server) serveTLS(addrString string) {
	err := http.ListenAndServeTLS(addrString, s.tls.cert, s.tls.key, s.n)
	if err != nil {
		restLogger.Error(err)
		s.err <- err
	}
}

func (s *Server) serve(ln net.Listener) {
	err := http.Serve(tcpKeepAliveListener{ln.(*net.TCPListener)}, s.n)
	if err != nil {
		restLogger.Error(err)
		s.err <- err
	}
}

// Monkey patch ListenAndServe and TCP alive code from https://golang.org/src/net/http/server.go
// The built in ListenAndServe and ListenAndServeTLS include TCP keepalive
// At this point the Go team is not wanting to provide separate listen and serve methods
// that also provide an exported TCP keepalive per: https://github.com/golang/go/issues/12731
type tcpKeepAliveListener struct {
	*net.TCPListener
}

func (ln tcpKeepAliveListener) Accept() (c net.Conn, err error) {
	tc, err := ln.AcceptTCP()
	if err != nil {
		return
	}
	tc.SetKeepAlive(true)
	tc.SetKeepAlivePeriod(3 * time.Minute)
	return tc, nil
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

func (s *Server) BindConfigManager(c managesConfig) {
	s.mc = c
}

func (s *Server) addRoutes() {
	// plugin routes
	s.r.GET("/v1/plugins", s.getPlugins)
	s.r.GET("/v1/plugins/:type", s.getPluginsByType)
	s.r.GET("/v1/plugins/:type/:name", s.getPluginsByName)
	s.r.GET("/v1/plugins/:type/:name/:version", s.getPlugin)
	s.r.POST("/v1/plugins", s.loadPlugin)
	s.r.DELETE("/v1/plugins/:type/:name/:version", s.unloadPlugin)
	s.r.GET("/v1/plugins/:type/:name/:version/config", s.getPluginConfigItem)
	s.r.PUT("/v1/plugins/:type/:name/:version/config", s.setPluginConfigItem)
	s.r.DELETE("/v1/plugins/:type/:name/:version/config", s.deletePluginConfigItem)

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

	// tribe routes
	if s.tr != nil {
		s.r.GET("/v1/tribe/agreements", s.getAgreements)
		s.r.POST("/v1/tribe/agreements", s.addAgreement)
		s.r.GET("/v1/tribe/agreements/:name", s.getAgreement)
		s.r.DELETE("/v1/tribe/agreements/:name", s.deleteAgreement)
		s.r.PUT("/v1/tribe/agreements/:name/join", s.joinAgreement)
		s.r.DELETE("/v1/tribe/agreements/:name/leave", s.leaveAgreement)
		s.r.GET("/v1/tribe/members", s.getMembers)
		s.r.GET("/v1/tribe/member/:name", s.getMember)
	}
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
	if !w.(negroni.ResponseWriter).Written() {
		w.WriteHeader(code)
	}

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
