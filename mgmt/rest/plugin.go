package rest

import (
	"compress/gzip"
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
	name       string
	version    int
	pluginType string
}

func (p *plugin) Name() string {
	return p.name
}

func (p *plugin) Version() int {
	return p.version
}

func (p *plugin) TypeName() string {
	return p.pluginType
}

func (s *Server) loadPlugin(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	mediaType, params, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
	if err != nil {
		respond(500, rbody.FromError(err), w)
		return
	}
	if strings.HasPrefix(mediaType, "multipart/") {
		lp := &rbody.PluginsLoaded{}
		lp.LoadedPlugins = make([]rbody.LoadedPlugin, 0)
		mr := multipart.NewReader(r.Body, params["boundary"])
		p, err := mr.NextPart()
		var fname, fname2 string
		for {
			if err == io.EOF {
				respond(201, lp, w)
				return
			}
			if err != nil {
				respond(500, rbody.FromError(err), w)
				return
			}
			if r.Header.Get("Plugin-Compression") == "gzip" {
				g, err := gzip.NewReader(p)
				if err != nil {
					respond(500, rbody.FromError(err), w)
					return
				}
				b, err := ioutil.ReadAll(g)
				if err != nil {
					respond(500, rbody.FromError(err), w)
					return
				}
				fname, err = writePlugin(s.mm.GetAutodiscoverPaths(), p.FileName(), b, fname)
				if err != nil {
				}
			} else {
				b, err := ioutil.ReadAll(p)
				if err != nil {
					respond(500, rbody.FromError(err), w)
					return
				}
				fname, err = writePlugin(s.mm.GetAutodiscoverPaths(), p.FileName(), b, fname)
			}
			if err != nil {
				respond(500, rbody.FromError(err), w)
				return
			}
			p, err = mr.NextPart()
			if err != nil {
				if fname2 == "" {
					fname2 = fname
				}
				pl, err := s.mm.Load(fname2)
				if err != nil {
					restLogger.Error(err)
					respond(500, rbody.FromPulseError(err), w)
					return
				}
				lp.LoadedPlugins = append(lp.LoadedPlugins, *catalogedPluginToLoaded(pl))
			} else {
				fname2 = fname
			}
		}
	}
}

func writePlugin(autoPaths []string, filename string, b []byte, fqfile string) (string, error) {
	var f *os.File
	var err error
	if len(autoPaths) > 0 {
		// write to first autoPath
		f, err = os.Create(path.Join(autoPaths[0], filename))
	} else {
		// write to temp location for binary
		if fqfile == "" {
			f, err = ioutil.TempFile("", filename)
		} else {
			// write asc to same location as binary
			f, err = os.Create(fqfile + ".asc")
		}
	}
	if err != nil {
		// respond(500, rbody.FromError(err), w)
		return "", err
	}
	n, err := f.Write(b)
	log.Debugf("wrote %v to %v", n, f.Name())
	if err != nil {
		// respond(500, rbody.FromError(err), w)
		return "", err
	}
	err = f.Chmod(0700)
	if err != nil {
		// respond(500, rbody.FromError(err), w)
		return "", err
	}
	// Close before load
	f.Close()
	return f.Name(), nil
}

func (s *Server) unloadPlugin(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	plName := p.ByName("name")
	plType := p.ByName("type")
	plVersion, iErr := strconv.ParseInt(p.ByName("version"), 10, 0)
	f := map[string]interface{}{
		"plugin-name":    plName,
		"plugin-version": plVersion,
		"plugin-type":    plType,
	}

	if iErr != nil {
		pe := perror.New(errors.New("invalid version"))
		pe.SetFields(f)
		respond(400, rbody.FromPulseError(pe), w)
		return
	}

	if plName == "" {
		pe := perror.New(errors.New("missing plugin name"))
		pe.SetFields(f)
		respond(400, rbody.FromPulseError(pe), w)
		return
	}
	if plType == "" {
		pe := perror.New(errors.New("missing plugin type"))
		pe.SetFields(f)
		respond(400, rbody.FromPulseError(pe), w)
		return
	}
	up, pe := s.mm.Unload(&plugin{
		name:       plName,
		version:    int(plVersion),
		pluginType: plType,
	})
	if pe != nil {
		pe.SetFields(f)
		respond(500, rbody.FromPulseError(pe), w)
		return
	}
	pr := &rbody.PluginUnloaded{
		Name:    up.Name(),
		Version: up.Version(),
		Type:    up.TypeName(),
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
		Signed:          c.IsSigned(),
		Status:          c.Status(),
		LoadedTimestamp: c.LoadedTimestamp().Unix(),
	}
}

func (s *Server) getPluginsByType(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
}

func (s *Server) getPluginsByName(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
}

func (s *Server) getPlugin(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
}
