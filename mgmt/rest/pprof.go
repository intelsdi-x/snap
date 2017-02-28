package rest

import (
	"net/http"
	"net/http/pprof"

	"github.com/julienschmidt/httprouter"
)

func (s *Server) addPprofRoutes() {
	if s.pprof {
		s.r.GET("/debug/pprof/", s.index)
		s.r.GET("/debug/pprof/block", s.index)
		s.r.GET("/debug/pprof/goroutine", s.index)
		s.r.GET("/debug/pprof/heap", s.index)
		s.r.GET("/debug/pprof/threadcreate", s.index)
		s.r.GET("/debug/pprof/cmdline", s.cmdline)
		s.r.GET("/debug/pprof/profile", s.profile)
		s.r.GET("/debug/pprof/symbol", s.symbol)
		s.r.GET("/debug/pprof/trace", s.trace)
	}
}

// profiling tools handlers

func (s *Server) index(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	pprof.Index(w, r)
}

func (s *Server) cmdline(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	pprof.Cmdline(w, r)
}

func (s *Server) profile(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	pprof.Profile(w, r)
}

func (s *Server) symbol(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	pprof.Symbol(w, r)
}

func (s *Server) trace(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	pprof.Trace(w, r)
}
