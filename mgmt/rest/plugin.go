package rest

import (
	"errors"
	"io"
	"io/ioutil"
	"mime"
	"mime/multipart"
	"net/http"
	"os"
	"path"
	"strconv"
	"strings"

	log "github.com/Sirupsen/logrus"
	"github.com/julienschmidt/httprouter"

	"github.com/intelsdi-x/pulse/core"
	"github.com/intelsdi-x/pulse/core/perror"
	"github.com/intelsdi-x/pulse/mgmt/rest/rbody"
)

var (
	ErrMissingPluginName = errors.New("missing plugin name")
)

type plugin struct {
	name    string
	version int
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
	mediaType, params, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
	if err != nil {
		log.Fatal(err)
		respond(500, rbody.FromError(err), w)
		return
	}
	if strings.HasPrefix(mediaType, "multipart/") {
		lp := &rbody.PluginsLoaded{}
		lp.LoadedPlugins = make([]rbody.LoadedPlugin, 0)
		mr := multipart.NewReader(r.Body, params["boundary"])
		for {
			p, err := mr.NextPart()
			if err == io.EOF {
				respond(201, lp, w)
				return
			}
			if err != nil {
				log.Fatal(err)
				respond(500, rbody.FromError(err), w)
				return
			}
			b, err := ioutil.ReadAll(p)
			if err != nil {
				log.Fatal(err)
				respond(500, rbody.FromError(err), w)
				return
			}
			autoPaths := s.mm.GetAutodiscoverPaths()
			var f *os.File
			if len(autoPaths) > 0 {
				// write to first autoPath
				f, err = os.Create(path.Join(autoPaths[0], p.FileName()))
			} else {
				// write to temp location
				f, err = ioutil.TempFile("", p.FileName())
			}
			if err != nil {
				log.Fatal(err)
				respond(500, rbody.FromError(err), w)
				return
			}
			n, err := f.Write(b)
			log.Debugf("wrote %v to %v", n, f.Name())
			if err != nil {
				log.Fatal(err)
				respond(500, rbody.FromError(err), w)
				return
			}
			err = f.Chmod(0700)
			if err != nil {
				log.Fatal(err)
				respond(500, rbody.FromError(err), w)
				return
			}

			pl, err := s.mm.Load(f.Name())
			if err != nil {
				log.Fatal(err)
				respond(500, rbody.FromError(err), w)
				return
			}
			lp.LoadedPlugins = append(lp.LoadedPlugins, *catalogedPluginToLoaded(pl))
		}
	}
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
		plugins.LoadedPlugins[i] = *catalogedPluginToLoaded(p)
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

func catalogedPluginToLoaded(c core.CatalogedPlugin) *rbody.LoadedPlugin {
	return &rbody.LoadedPlugin{
		Name:            c.Name(),
		Version:         c.Version(),
		Type:            c.TypeName(),
		Status:          c.Status(),
		LoadedTimestamp: c.LoadedTimestamp().Unix(),
	}
}

func (s *Server) getPluginsByName(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
}

func (s *Server) getPlugin(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
}
