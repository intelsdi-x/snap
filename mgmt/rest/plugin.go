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
		var files []string
		var i int
		for {
			var b []byte
			files = append(files, "")
			p, err := mr.NextPart()
			if err == io.EOF {
				files = files[:len(files)-1]
				break
			}
			if err != nil {
				respond(500, rbody.FromError(err), w)
				return
			}
			if r.Header.Get("Plugin-Compression") == "gzip" {
				g, err := gzip.NewReader(p)
				defer g.Close()
				if err != nil {
					respond(500, rbody.FromError(err), w)
					return
				}
				b, err = ioutil.ReadAll(g)
				if err != nil {
					respond(500, rbody.FromError(err), w)
					return
				}
			} else {
				b, err = ioutil.ReadAll(p)
				if err != nil {
					respond(500, rbody.FromError(err), w)
					return
				}
			}
			files[i], err = writeFile(p.FileName(), b)
			if err != nil {
				respond(500, rbody.FromError(err), w)
				return
			}
			i++
		}
		restLogger.Info("Loading plugin: ", files[0])
		pl, err := s.mm.Load(files...)
		if err != nil {
			var ec int
			restLogger.Error(err)
			for _, f := range files {
				restLogger.Debugf("Removing file (%s) after failure to load plugin (%s)", f, files[0])
				err := os.Remove(f)
				if err != nil {
					restLogger.Error(err)
				}
			}
			rb := rbody.FromPulseError(err)
			switch rb.ResponseBodyMessage() {
			case PluginAlreadyLoaded:
				ec = 409
			default:
				ec = 500
			}
			respond(ec, rb, w)
			return
		}
		lp.LoadedPlugins = append(lp.LoadedPlugins, *catalogedPluginToLoaded(pl))
		respond(201, lp, w)
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
	n, err := f.Write(b)
	log.Debugf("wrote %v to %v", n, f.Name())
	if err != nil {
		return "", err
	}
	err = f.Chmod(0700)
	if err != nil {
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

	plugins := new(rbody.PluginList)

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

func (s *Server) getPlugin(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
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

	pluginCatalog := s.mm.PluginCatalog()
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
		pe := perror.New(ErrPluginNotFound, f)
		respond(404, rbody.FromPulseError(pe), w)
		return
	}

	rd := r.FormValue("download")
	d, _ := strconv.ParseBool(rd)
	if d {
		b, err := ioutil.ReadFile(plugin.PluginPath())
		if err != nil {
			f["plugin-path"] = plugin.PluginPath()
			pe := perror.New(err, f)
			respond(500, rbody.FromPulseError(pe), w)
			return
		}

		w.Header().Set("Content-Encoding", "gzip")
		gz := gzip.NewWriter(w)
		defer gz.Close()
		_, err = gz.Write(b)
		if err != nil {
			f["plugin-path"] = plugin.PluginPath()
			pe := perror.New(err, f)
			respond(500, rbody.FromPulseError(pe), w)
			return
		}
		return
	}
}
