package rest

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
)

type plugin struct {
	Name    string `json:"name"`
	Version int    `json:"version"`
}

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
