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

package v1

import (
	"compress/gzip"
	"crypto/sha256"
	"encoding/json"
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

	"github.com/intelsdi-x/snap/core"
	"github.com/intelsdi-x/snap/core/serror"
	"github.com/intelsdi-x/snap/mgmt/rest/api"
	"github.com/intelsdi-x/snap/mgmt/rest/v1/rbody"
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

func (s *apiV1) loadPlugin(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	lp := &rbody.PluginsLoaded{}
	lp.LoadedPlugins = make([]rbody.LoadedPlugin, 0)
	var rp *core.RequestedPlugin
	mediaType, params, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
	if err != nil {
		rbody.Write(500, rbody.FromError(err), w)
		return
	}
	if strings.HasPrefix(mediaType, "multipart/") {
		var certPath string
		var keyPath string
		var caCertPaths string

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
				rbody.Write(500, rbody.FromError(err), w)
				return
			}
			if r.Header.Get("Plugin-Compression") == "gzip" {
				g, err := gzip.NewReader(p)
				defer g.Close()
				if err != nil {
					rbody.Write(500, rbody.FromError(err), w)
					return
				}
				b, err = ioutil.ReadAll(g)
				if err != nil {
					rbody.Write(500, rbody.FromError(err), w)
					return
				}
			} else {
				b, err = ioutil.ReadAll(p)
				if err != nil {
					rbody.Write(500, rbody.FromError(err), w)
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
			// Reception of TLS security file paths (ceritificate file, private key file, CA certificate files)
			// is also taking place here. Paths are extracted and used to set up a RequestedPlugin object.
			switch {
			case i == 0:
				if filepath.Ext(p.FileName()) == ".asc" {
					e := errors.New("Error: first file passed to load plugin api can not be signature file")
					rbody.Write(500, rbody.FromError(e), w)
					return
				}
				if rp, err = core.NewRequestedPlugin(p.FileName(), s.metricManager.GetTempDir(), b); err != nil {
					rbody.Write(500, rbody.FromError(err), w)
					return
				}
				checkSum = sha256.Sum256(b)
			case i < 5:
				if filepath.Ext(p.FileName()) == ".asc" {
					signature = b
				} else if strings.HasPrefix(p.FileName(), TLSCertPrefix) {
					certPath = string(b)
					if _, err := os.Stat(certPath); os.IsNotExist(err) {
						e := errors.New("Error: given certificate file is not available")
						rbody.Write(500, rbody.FromError(e), w)
						return
					}
				} else if strings.HasPrefix(p.FileName(), TLSKeyPrefix) {
					keyPath = string(b)
					if _, err := os.Stat(keyPath); os.IsNotExist(err) {
						e := errors.New("Error: given key file is not available")
						rbody.Write(500, rbody.FromError(e), w)
						return
					}
				} else if strings.HasPrefix(p.FileName(), TLSCACertsPrefix) {
					caCertPaths = string(b)
					// validation will take place later; take it as it is
				} else {
					e := errors.New("Error: unrecognized file was passed")
					rbody.Write(500, rbody.FromError(e), w)
					return
				}
			case i == 5:
				e := errors.New("Error: More than five files passed to the load plugin API")
				rbody.Write(500, rbody.FromError(e), w)
				return
			}
			i++
		}
		// Sanity check, verify the checkSum on the file sent is the same
		// as after it is written to disk.
		if rp.CheckSum() != checkSum {
			e := errors.New("Error: CheckSum mismatch on requested plugin to load")
			rbody.Write(500, rbody.FromError(e), w)
			return
		}
		rp.SetSignature(signature)
		rp.SetCertPath(certPath)
		rp.SetKeyPath(keyPath)
		rp.SetCACertPaths(caCertPaths)
		if certPath != "" && keyPath != "" {
			rp.SetTLSEnabled(true)
		} else if certPath != "" || keyPath != "" {
			e := errors.New("Error: TLS setup incomplete - missing one of pair: certificate, key files")
			rbody.Write(500, rbody.FromError(e), w)
			return
		}
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
			rb := rbody.FromError(err)
			switch rb.ResponseBodyMessage() {
			case PluginAlreadyLoaded:
				ec = 409
			default:
				ec = 500
			}
			rbody.Write(ec, rb, w)
			return
		}
		lp.LoadedPlugins = append(lp.LoadedPlugins, catalogedPluginToLoaded(r.Host, pl))
		rbody.Write(201, lp, w)
	} else if strings.HasSuffix(mediaType, "json") {
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			rbody.Write(500, rbody.FromError(err), w)
			return
		}
		var resp map[string]string
		err = json.Unmarshal(body, &resp)
		if err != nil {
			rbody.Write(500, rbody.FromError(err), w)
		}
		rp, err := core.NewRequestedPlugin(resp["uri"], "", nil)
		if err != nil {
			rbody.Write(500, rbody.FromError(err), w)
			return
		}
		pl, err := s.metricManager.Load(rp)
		if err != nil {
			// TODO (JC) should return 409 if plugin already loaded
			rbody.Write(500, rbody.FromError(err), w)
			return
		}
		lp.LoadedPlugins = append(lp.LoadedPlugins, catalogedPluginToLoaded(r.Host, pl))
		rbody.Write(201, lp, w)
	}
}

func (s *apiV1) unloadPlugin(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	plName := p.ByName("name")
	plType := p.ByName("type")
	plVersion, iErr := strconv.ParseInt(p.ByName("version"), 10, 0)
	f := map[string]interface{}{
		"plugin-name":    plName,
		"plugin-version": plVersion,
		"plugin-type":    plType,
	}

	if iErr != nil {
		se := serror.New(errors.New("invalid version"))
		se.SetFields(f)
		rbody.Write(400, rbody.FromSnapError(se), w)
		return
	}

	if plName == "" {
		se := serror.New(errors.New("missing plugin name"))
		se.SetFields(f)
		rbody.Write(400, rbody.FromSnapError(se), w)
		return
	}
	if plType == "" {
		se := serror.New(errors.New("missing plugin type"))
		se.SetFields(f)
		rbody.Write(400, rbody.FromSnapError(se), w)
		return
	}
	up, se := s.metricManager.Unload(&plugin{
		name:       plName,
		version:    int(plVersion),
		pluginType: plType,
	})
	if se != nil {
		se.SetFields(f)
		rbody.Write(500, rbody.FromSnapError(se), w)
		return
	}
	pr := &rbody.PluginUnloaded{
		Name:    up.Name(),
		Version: up.Version(),
		Type:    up.TypeName(),
	}
	rbody.Write(200, pr, w)
}

func (s *apiV1) getPlugins(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	var detail bool
	for k := range r.URL.Query() {
		if k == "details" {
			detail = true
		}
	}
	plName := params.ByName("name")
	plType := params.ByName("type")
	rbody.Write(200, getPlugins(s.metricManager, detail, r.Host, plName, plType), w)
}

func getPlugins(mm api.Metrics, detail bool, h string, plName string, plType string) *rbody.PluginList {

	plCatalog := mm.PluginCatalog()

	plugins := rbody.PluginList{}

	plugins.LoadedPlugins = make([]rbody.LoadedPlugin, len(plCatalog))
	for i, p := range plCatalog {
		plugins.LoadedPlugins[i] = catalogedPluginToLoaded(h, p)
	}

	if detail {
		aPlugins := mm.AvailablePlugins()
		plugins.AvailablePlugins = make([]rbody.AvailablePlugin, len(aPlugins))
		for i, p := range aPlugins {
			plugins.AvailablePlugins[i] = rbody.AvailablePlugin{
				Name:             p.Name(),
				Version:          p.Version(),
				Type:             p.TypeName(),
				HitCount:         p.HitCount(),
				LastHitTimestamp: p.LastHit().Unix(),
				ID:               p.ID(),
				Href:             pluginURI(h, version, p),
				PprofPort:        p.Port(),
			}
		}
	}

	filteredPlugins := rbody.PluginList{}

	if plName != "" {
		for _, p := range plugins.LoadedPlugins {
			if p.Name == plName {
				filteredPlugins.LoadedPlugins = append(filteredPlugins.LoadedPlugins, p)
			}
		}
		for _, p := range plugins.AvailablePlugins {
			if p.Name == plName {
				filteredPlugins.AvailablePlugins = append(filteredPlugins.AvailablePlugins, p)
			}
		}
		// update plugins so that further filters consider previous filters
		plugins = filteredPlugins
	}

	filteredPlugins = rbody.PluginList{}

	if plType != "" {
		for _, p := range plugins.LoadedPlugins {
			if p.Type == plType {
				filteredPlugins.LoadedPlugins = append(filteredPlugins.LoadedPlugins, p)
			}
		}
		for _, p := range plugins.AvailablePlugins {
			if p.Type == plType {
				filteredPlugins.AvailablePlugins = append(filteredPlugins.AvailablePlugins, p)
			}
		}
		// filter based on type
		plugins = filteredPlugins
	}

	return &plugins
}

func catalogedPluginToLoaded(host string, c core.CatalogedPlugin) rbody.LoadedPlugin {
	return rbody.LoadedPlugin{
		Name:            c.Name(),
		Version:         c.Version(),
		Type:            c.TypeName(),
		Signed:          c.IsSigned(),
		Status:          c.Status(),
		LoadedTimestamp: c.LoadedTimestamp().Unix(),
		Href:            pluginURI(host, version, c),
	}
}

func (s *apiV1) getPlugin(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	plName := p.ByName("name")
	plType := p.ByName("type")
	plVersion, iErr := strconv.ParseInt(p.ByName("version"), 10, 0)
	f := map[string]interface{}{
		"plugin-name":    plName,
		"plugin-version": plVersion,
		"plugin-type":    plType,
	}

	if iErr != nil {
		se := serror.New(errors.New("invalid version"))
		se.SetFields(f)
		rbody.Write(400, rbody.FromSnapError(se), w)
		return
	}

	if plName == "" {
		se := serror.New(errors.New("missing plugin name"))
		se.SetFields(f)
		rbody.Write(400, rbody.FromSnapError(se), w)
		return
	}
	if plType == "" {
		se := serror.New(errors.New("missing plugin type"))
		se.SetFields(f)
		rbody.Write(400, rbody.FromSnapError(se), w)
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
		rbody.Write(404, rbody.FromSnapError(se), w)
		return
	}

	rd := r.FormValue("download")
	d, _ := strconv.ParseBool(rd)
	var configPolicy []rbody.PolicyTable
	if plugin.TypeName() == "processor" || plugin.TypeName() == "publisher" {
		rules := plugin.Policy().Get([]string{""}).RulesAsTable()
		configPolicy = make([]rbody.PolicyTable, 0, len(rules))
		for _, r := range rules {
			configPolicy = append(configPolicy, rbody.PolicyTable{
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
			rbody.Write(500, rbody.FromSnapError(se), w)
			return
		}

		w.Header().Set("Content-Encoding", "gzip")
		gz := gzip.NewWriter(w)
		defer gz.Close()
		_, err = gz.Write(b)
		if err != nil {
			f["plugin-path"] = plugin.PluginPath()
			se := serror.New(err, f)
			rbody.Write(500, rbody.FromSnapError(se), w)
			return
		}
		return
	} else {
		pluginRet := &rbody.PluginReturned{
			Name:            plugin.Name(),
			Version:         plugin.Version(),
			Type:            plugin.TypeName(),
			Signed:          plugin.IsSigned(),
			Status:          plugin.Status(),
			LoadedTimestamp: plugin.LoadedTimestamp().Unix(),
			Href:            pluginURI(r.Host, version, plugin),
			ConfigPolicy:    configPolicy,
		}
		rbody.Write(200, pluginRet, w)
	}
}

func pluginURI(host, version string, c core.Plugin) string {
	return fmt.Sprintf("%s://%s/%s/plugins/%s/%s/%d", protocolPrefix, host, version, c.TypeName(), c.Name(), c.Version())
}
