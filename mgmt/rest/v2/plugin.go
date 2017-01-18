package v2

import (
	"compress/gzip"
	"crypto/sha256"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"mime"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"path"
	"runtime"

	log "github.com/Sirupsen/logrus"
	"github.com/intelsdi-x/snap/control"
	"github.com/intelsdi-x/snap/core"
	"github.com/intelsdi-x/snap/core/serror"
	"github.com/intelsdi-x/snap/mgmt/rest/v2/response"
	"github.com/julienschmidt/httprouter"
)

const PluginAlreadyLoaded = "plugin is already loaded"

var (
	ErrMissingPluginName = errors.New("missing plugin name")
	ErrPluginNotFound    = errors.New("plugin not found")
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

func (s *V2) loadPlugin(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	mediaType, params, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
	if err != nil {
		response.Write(415, response.FromError(err), w)
		return
	}
	if strings.HasPrefix(mediaType, "multipart/") {
		var pluginPath string
		var signature []byte
		var checkSum [sha256.Size]byte
		mr := multipart.NewReader(r.Body, params["boundary"])
		var i int
		for {
			var b []byte
			p, err := mr.NextPart()
			if err == io.EOF {
				break
			}
			if err != nil {
				response.Write(500, response.FromError(err), w)
				return
			}
			if r.Header.Get("Plugin-Compression") == "gzip" {
				g, err := gzip.NewReader(p)
				defer g.Close()
				if err != nil {
					response.Write(500, response.FromError(err), w)
					return
				}
				b, err = ioutil.ReadAll(g)
				if err != nil {
					response.Write(500, response.FromError(err), w)
					return
				}
			} else {
				b, err = ioutil.ReadAll(p)
				if err != nil {
					response.Write(500, response.FromError(err), w)
					return
				}
			}

			// A little sanity checking for files being passed into the API server.
			// First file passed in should be the plugin. If the first file is a signature
			// file, an error is returned. The signature file should be the second
			// file passed to the API server. If the second file does not have the ".asc"
			// extension, an error is returned.
			// If we loop around more than twice before receiving io.EOF, then
			// an error is returned.

			switch {
			case i == 0:
				if filepath.Ext(p.FileName()) == ".asc" {
					e := errors.New("Error: first file passed to load plugin api can not be signature file")
					response.Write(400, response.FromError(e), w)
					return
				}
				if pluginPath, err = writeFile(p.FileName(), b); err != nil {
					response.Write(500, response.FromError(err), w)
					return
				}
				checkSum = sha256.Sum256(b)
			case i == 1:
				if filepath.Ext(p.FileName()) == ".asc" {
					signature = b
				} else {
					e := errors.New("Error: second file passed was not a signature file")
					response.Write(400, response.FromError(e), w)
					return
				}
			case i == 2:
				e := errors.New("Error: More than two files passed to the load plugin api")
				response.Write(400, response.FromError(e), w)
				return
			}
			i++
		}
		rp, err := core.NewRequestedPlugin(pluginPath)
		if err != nil {
			response.Write(500, response.FromError(err), w)
			return
		}
		rp.SetAutoLoaded(false)
		// Sanity check, verify the checkSum on the file sent is the same
		// as after it is written to disk.
		if rp.CheckSum() != checkSum {
			e := errors.New("Error: CheckSum mismatch on requested plugin to load")
			response.Write(400, response.FromError(e), w)
			return
		}
		rp.SetSignature(signature)
		restLogger.Info("Loading plugin: ", rp.Path())
		pl, err := s.metricManager.Load(rp)
		if err != nil {
			var ec int
			restLogger.Error(err)
			restLogger.Debugf("Removing file (%s)", rp.Path())
			err2 := os.RemoveAll(filepath.Dir(rp.Path()))
			if err2 != nil {
				restLogger.Error(err2)
			}
			rb := response.FromError(err)
			switch rb.ErrorMessage {
			case PluginAlreadyLoaded:
				ec = 409
			default:
				ec = 500
			}
			response.Write(ec, rb, w)
			return
		}
		response.Write(201, catalogedPluginBody(r.Host, pl), w)
	}
}

func writeFile(filename string, b []byte) (string, error) {
	// Create temporary directory
	dir, err := ioutil.TempDir("", "")
	if err != nil {
		return "", err
	}
	f, err := os.Create(path.Join(dir, filename))
	if err != nil {
		return "", err
	}
	// Close before load
	defer f.Close()

	n, err := f.Write(b)
	log.Debugf("wrote %v to %v", n, f.Name())
	if err != nil {
		return "", err
	}
	if runtime.GOOS != "windows" {
		err = f.Chmod(0700)
		if err != nil {
			return "", err
		}
	}
	return f.Name(), nil
}

func (s *V2) unloadPlugin(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	plName := p.ByName("name")
	plType := p.ByName("type")
	plVersion, err := strconv.ParseInt(p.ByName("version"), 10, 0)
	f := map[string]interface{}{
		"plugin-name":    plName,
		"plugin-version": plVersion,
		"plugin-type":    plType,
	}

	if err != nil || plName == "" || plType == "" {
		se := serror.New(errors.New("missing or invalid parameter(s)"))
		se.SetFields(f)
		response.Write(400, response.FromSnapError(se), w)
		return
	}

	_, se := s.metricManager.Unload(&plugin{
		name:       plName,
		version:    int(plVersion),
		pluginType: plType,
	})

	// 404 - plugin not found
	// 409 - plugin state is not plugin loaded
	// 500 - removing plugin from /tmp failed
	if se != nil {
		se.SetFields(f)
		statusCode := 500
		switch se.Error() {
		case control.ErrPluginNotFound.Error():
			statusCode = 404
		case control.ErrPluginNotInLoadedState.Error():
			statusCode = 409
		}
		response.Write(statusCode, response.FromSnapError(se), w)
		return
	}
	response.Write(204, nil, w)
}

func (s *V2) getPlugins(w http.ResponseWriter, r *http.Request, params httprouter.Params) {

	// filter by plugin name or plugin type
	plName := params.ByName("name")
	plType := params.ByName("type")
	nbFilter := Btoi(plName != "") + Btoi(plType != "")

	if _, detail := r.URL.Query()["running"]; detail {
		// get running plugins
		plugins := runningPluginsBody(r.Host, s.metricManager.AvailablePlugins())
		filteredPlugins := []response.RunningPlugin{}
		if nbFilter > 0 {
			for _, p := range plugins {
				if nbFilter == 1 && (p.Name == plName || p.Type == plType) || nbFilter == 2 && (p.Name == plName && p.Type == plType) {
					filteredPlugins = append(filteredPlugins, p)
				}
			}
		} else {
			filteredPlugins = plugins
		}
		response.Write(200, filteredPlugins, w)
	} else {
		// get plugins from the plugin catalog
		plugins := pluginCatalogBody(r.Host, s.metricManager.PluginCatalog())
		filteredPlugins := []response.Plugin{}

		if nbFilter > 0 {
			for _, p := range plugins {
				if nbFilter == 1 && (p.Name == plName || p.Type == plType) || nbFilter == 2 && (p.Name == plName && p.Type == plType) {
					filteredPlugins = append(filteredPlugins, p)
				}
			}
		} else {
			filteredPlugins = plugins
		}
		response.Write(200, filteredPlugins, w)
	}
}

func Btoi(b bool) int {
	if b {
		return 1
	}
	return 0
}

func pluginCatalogBody(host string, c []core.CatalogedPlugin) []response.Plugin {
	plugins := make([]response.Plugin, len(c))
	for i, p := range c {
		plugins[i] = catalogedPluginBody(host, p)
	}
	return plugins
}

func catalogedPluginBody(host string, c core.CatalogedPlugin) response.Plugin {
	return response.Plugin{
		Name:            c.Name(),
		Version:         c.Version(),
		Type:            c.TypeName(),
		Signed:          c.IsSigned(),
		Status:          c.Status(),
		LoadedTimestamp: c.LoadedTimestamp().Unix(),
		Href:            pluginURI(host, version, c),
	}
}

func runningPluginsBody(host string, c []core.AvailablePlugin) []response.RunningPlugin {
	plugins := make([]response.RunningPlugin, len(c))
	for i, p := range c {
		plugins[i] = response.RunningPlugin{
			Name:             p.Name(),
			Version:          p.Version(),
			Type:             p.TypeName(),
			HitCount:         p.HitCount(),
			LastHitTimestamp: p.LastHit().Unix(),
			ID:               p.ID(),
			Href:             pluginURI(host, version, p),
			PprofPort:        p.Port(),
		}
	}
	return plugins
}

func pluginURI(host, version string, c core.Plugin) string {
	return fmt.Sprintf("%s://%s/%s/plugins/%s/%s/%d", protocolPrefix, host, version, c.TypeName(), c.Name(), c.Version())
}

func (s *V2) getPlugin(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	plName := p.ByName("name")
	plType := p.ByName("type")
	plVersion, err := strconv.ParseInt(p.ByName("version"), 10, 0)
	f := map[string]interface{}{
		"plugin-name":    plName,
		"plugin-version": plVersion,
		"plugin-type":    plType,
	}
	if err != nil || plName == "" || plType == "" {
		se := serror.New(errors.New("missing or invalid parameter(s)"))
		se.SetFields(f)
		response.Write(400, response.FromSnapError(se), w)
		return
	}

	pluginCatalog := s.metricManager.PluginCatalog()
	var plugin core.CatalogedPlugin
	for _, item := range pluginCatalog {
		if item.Name() == plName &&
			item.Version() == int(plVersion) &&
			item.TypeName() == plType {
			plugin = item
			break
		}
	}
	if plugin == nil {
		se := serror.New(ErrPluginNotFound, f)
		response.Write(404, response.FromSnapError(se), w)
		return
	}

	rd := r.FormValue("download")
	d, _ := strconv.ParseBool(rd)
	var configPolicy []response.PolicyTable
	if plugin.TypeName() == "processor" || plugin.TypeName() == "publisher" {
		rules := plugin.Policy().Get([]string{""}).RulesAsTable()
		configPolicy = make([]response.PolicyTable, 0, len(rules))
		for _, r := range rules {
			configPolicy = append(configPolicy, response.PolicyTable{
				Name:     r.Name,
				Type:     r.Type,
				Default:  r.Default,
				Required: r.Required,
				Minimum:  r.Minimum,
				Maximum:  r.Maximum,
			})
		}

	} else {
		configPolicy = nil
	}

	if d {
		b, err := ioutil.ReadFile(plugin.PluginPath())
		if err != nil {
			f["plugin-path"] = plugin.PluginPath()
			se := serror.New(err, f)
			response.Write(500, response.FromSnapError(se), w)
			return
		}

		w.Header().Set("Content-Type", "application/octet-stream")
		w.Header().Set("Content-Encoding", "gzip")
		gz := gzip.NewWriter(w)
		defer gz.Close()
		_, err = gz.Write(b)
		if err != nil {
			f["plugin-path"] = plugin.PluginPath()
			se := serror.New(err, f)
			response.Write(500, response.FromSnapError(se), w)
			return
		}
		w.WriteHeader(200)
		return
	} else {
		pluginRet := response.Plugin{
			Name:            plugin.Name(),
			Version:         plugin.Version(),
			Type:            plugin.TypeName(),
			Signed:          plugin.IsSigned(),
			Status:          plugin.Status(),
			LoadedTimestamp: plugin.LoadedTimestamp().Unix(),
			Href:            pluginURI(r.Host, version, plugin),
			ConfigPolicy:    configPolicy,
		}
		response.Write(200, pluginRet, w)
	}
}
