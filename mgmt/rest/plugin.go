package rest

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	// log "github.com/Sirupsen/logrus"
	"github.com/julienschmidt/httprouter"

	"github.com/intelsdi-x/pulse/core/perror"
	"github.com/intelsdi-x/pulse/mgmt/rest/rbody"
)

var (
	ErrMissingPluginName = errors.New("missing plugin name")
)

type loadedPlugin struct {
	*plugin
	TypeName        string `json:"type"`
	Status          string `json:"status"`
	LoadedTimestamp int64  `json:"loaded_timestamp"`
}

type availablePlugin struct {
	*plugin
	TypeName string    `json:"type"`
	HitCount int       `json:"hit_count"`
	LastHit  time.Time `json:"last_hit"`
	ID       int       `json:"ID"`
}

type plugin struct {
	name    string `json:"name"`
	version int    `json:"version"`
}

func (p *plugin) Name() string {
	return p.name
}

func (p *plugin) Version() int {
	return p.version
}

func (s *Server) loadPlugin(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	// Disabling in expectation of merging #184

	// loadRequest := make(map[string]string)
	// errCode, err := marshalBody(&loadRequest, r.Body)
	// if errCode != 0 && err != nil {
	// 	replyError(errCode, w, err)
	// 	return
	// }
	// loadErr := s.mm.Load(loadRequest["path"])
	// if loadErr != nil {
	// 	// restLogger.WithFields(log.Fields{
	// 	// 	"method": r.Method,
	// 	// 	"url":    r.URL.Path,
	// 	// }).WithFields(loadErr.Fields()).Warning(err.Error())
	// 	replyError(500, w, loadErr)
	// 	return
	// }
	// replySuccess(200, "", w, nil)
}

func (s *Server) unloadPlugin(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	plName := p.ByName("name")
	plVersion, iErr := strconv.ParseInt(p.ByName("version"), 10, 0)
	f := map[string]interface{}{
		"plugin-name":    plName,
		"plugin-version": plVersion,
	}

	if iErr != nil {
		pe := perror.New(ErrMissingPluginName)
		pe.SetFields(f)
		respond(500, rbody.FromPulseError(pe), w)
		return
	}

	if plName == "" {
		pe := perror.New(errors.New("missing plugin name"))
		pe.SetFields(f)
		respond(500, rbody.FromPulseError(pe), w)
		return
	}
	pe := s.mm.Unload(&plugin{name: plName, version: int(plVersion)})
	if pe != nil {
		pe.SetFields(f)
		respond(500, rbody.FromPulseError(pe), w)
		return
	}
	pr := &rbody.PluginUnloaded{
		Name:    plName,
		Version: int(plVersion),
	}
	respond(200, pr, w)
}

func (s *Server) getPlugins(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	var detail bool
	// make this a function because DRY
	for k, _ := range r.URL.Query() {
		if k == "details" {
			detail = true
		}
	}

	plugins := new(rbody.PluginListReturned)

	// Cache the catalog here to avoid multiple reads
	plCatalog := s.mm.PluginCatalog()
	plugins.LoadedPlugins = make([]rbody.LoadedPlugin, len(plCatalog))
	for i, p := range s.mm.PluginCatalog() {
		plugins.LoadedPlugins[i] = rbody.LoadedPlugin{
			Name:            p.Name(),
			Version:         p.Version(),
			Type:            p.TypeName(),
			Status:          p.Status(),
			LoadedTimestamp: p.LoadedTimestamp(),
		}
	}

	if detail {
		aPlugins := s.mm.AvailablePlugins()
		plugins.AvailablePlugins = make([]rbody.AvailablePlugin, len(aPlugins))
		for i, p := range aPlugins {
			plugins.AvailablePlugins[i] = rbody.AvailablePlugin{
				Name:             p.Name(),
				Version:          p.Version(),
				Type:             p.TypeName(),
				HitCount:         p.HitCount(),
				LastHitTimestamp: p.LastHit().Unix(),
				ID:               p.ID(),
			}
		}
	}

	respond(200, plugins, w)
}

func (s *Server) getPluginsByName(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
}

func (s *Server) getPlugin(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
}
