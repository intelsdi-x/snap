/*
http://www.apache.org/licenses/LICENSE-2.0.txt


Copyright 2017 Intel Corporation

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

package v2

import (
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"mime"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/intelsdi-x/snap/control"
	"github.com/intelsdi-x/snap/core"
	"github.com/intelsdi-x/snap/core/serror"
	"github.com/julienschmidt/httprouter"
)

const (
	// TLSCertPrefix defines a prefix for file fragment carrying path to TLS certificate
	TLSCertPrefix = "crt."
	// TLSKeyPrefix defines a prefix for file fragment carrying path to TLS private key
	TLSKeyPrefix = "key."
	// TLSCACertsPrefix defines a prefix for file fragment carrying paths to TLS CA certificates
	TLSCACertsPrefix = "cacerts."
)

// PluginResponse represents the response from plugin operations.
//
// swagger:response PluginResponse
type PluginResponse struct {
	// List of a plugin
	//
	// in: body
	Plugin Plugin `json:"plugin,omitempty"`
}

// PluginsResp represents the response from plugins operations.
//
// swagger:response PluginsResponse
type PluginsResp struct {
	// List of plugins
	//
	// in: body
	Body struct {
		Plugins []Plugin `json:"plugins,omitempty"`
	}
}

type PluginsResponse struct {
	Plugins []Plugin `json:"plugins,omitempty"`
}

// Plugin represents a plugin type definition.
type Plugin struct {
	Name             string        `json:"name"`
	Version          int           `json:"version"`
	Type             string        `json:"type"`
	Signed           bool          `json:"signed"`
	Status           string        `json:"status"`
	LoadedTimestamp  int64         `json:"loaded_timestamp,omitempty"`
	Href             string        `json:"href,omitempty"`
	ConfigPolicy     []PolicyTable `json:"config_policy,omitempty"`
	HitCount         int           `json:"hitcount,omitempty"`
	LastHitTimestamp int64         `json:"last_hit_timestamp,omitempty"`
	ID               uint32        `json:"id,omitempty"`
	PprofPort        string        `json:"pprof_port,omitempty"`
}

// PluginParams represents the request path plugin name, version and type.
//
// swagger:parameters getPlugin unloadPlugin getPluginConfigItem setPluginConfigItem
type PluginParams struct {
	// required: true
	// in: path
	PName string `json:"pname"`
	// required: true
	// in: path
	PVersion int `json:"pversion"`
	// required: true
	// in: path
	// enum: collector, processor, publisher
	PType string `json:"ptype"`
}

// PluginsParams represents the query parameters for getting a list of plugins.
//
// swagger:parameters getPlugins
type PluginsParams struct {
	// in: query
	Name string `json:"name"`
	// in: query
	// enum: collector, processor, publisher
	Type string `json:"type"`
	// in: query
	Running bool `json:"running"`
}

// PluginPostParams defines type for loading a plugin.
//
// swagger:parameters loadPlugin
type PluginPostParams struct {
	// loads a plugin.
	//
	// in:formData
	//
	// swagger:file
	PluginData *bytes.Buffer `json:"plugin_data"`
	// Plugin GRPC TLS server key
	//
	// in: formData
	//
	PluginKey string `json:"plugin_key"`
	// Plugin GRPC TLS server certification
	//
	// in: formData
	//
	PluginCert string `json:"plugin_cert"`
	// CA root certification
	//
	// in: formData
	//
	CACerts string `json:"ca_certs"`
	// Stand-alone plugin URI
	//
	// in: formData
	//
	PluginURI string `json:"plugin_uri"`
}

// Map for collecting HTTP form field data
type formFieldMap map[string]formField
type formField struct {
	fileName string
	data     []byte
}

// Name plugin name string
func (p *PluginParams) Name() string {
	return p.PName
}

// Version plugin version integer
func (p *PluginParams) Version() int {
	return p.PVersion
}

// TypeName plugin type string.
// They are collector, processor and publisher.
func (p *PluginParams) TypeName() string {
	return p.PType
}

func (s *apiV2) loadPlugin(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	var rp *core.RequestedPlugin
	mediaType, params, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
	if err != nil {
		Write(415, FromError(err), w)
		return
	}

	if strings.HasPrefix(mediaType, "multipart/") {
		var certPath, keyPath, caCertPaths string
		var signature []byte
		var checkSum [sha256.Size]byte

		// Go OpenAPI sends URL-encoded forms (without boundary data) if no file fields were passed.
		// In standalone plugin mode, only plugin address is passed, so because of this behavior,
		// multipart reader does not work and data needs to be processed using url.ParseQuery().
		// Otherwise, incoming data is parsed using multipart reader.

		// Reading HTTP form fields
		formFields := make(formFieldMap)
		if params["boundary"] == "" {
			b, err := ioutil.ReadAll(r.Body)
			if err != nil {
				Write(500, FromError(err), w)
				return
			}
			data, err := url.ParseQuery(string(b))
			if err != nil {
				Write(500, FromError(err), w)
				return
			}
			for key, value := range data {
				formFields[key] = formField{fileName: "", data: []byte(value[0])}
			}
		} else {
			mr := multipart.NewReader(r.Body, params["boundary"])

			var b []byte
			for {
				p, err := mr.NextPart()
				if err == io.EOF {
					break
				}
				if err != nil {
					Write(500, FromError(err), w)
					return
				}
				if r.Header.Get("Plugin-Compression") == "gzip" {
					g, err := gzip.NewReader(p)
					defer g.Close()
					if err != nil {
						Write(500, FromError(err), w)
						return
					}
					b, err = ioutil.ReadAll(g)
					if err != nil {
						Write(500, FromError(err), w)
						return
					}
				} else {
					b, err = ioutil.ReadAll(p)
					if err != nil {
						Write(500, FromError(err), w)
						return
					}
				}
				formFields[p.FormName()] = formField{fileName: p.FileName(), data: b}
			}
		}

		// Handle actions for received form fields
		for fieldName, field := range formFields {
			switch fieldName {
			case "ca_certs":
				caCertPaths = string(field.data)
				handleError(caCertPaths, w)
			case "plugin_key":
				keyPath = string(field.data)
				handleError(keyPath, w)
			case "plugin_cert":
				certPath = string(field.data)
				handleError(certPath, w)
			//plugin_data is from REST API and snap-plugins is from rest_v2_test.go.
			case "plugin_data", "snap-plugins":
				rp, err = core.NewRequestedPlugin(field.fileName, s.metricManager.GetTempDir(), field.data)
				if err != nil {
					Write(500, FromError(err), w)
					return
				}
				checkSum = sha256.Sum256(field.data)
			case "plugin_uri":
				pluginURI := string(field.data)
				rp, err = core.NewRequestedPlugin(pluginURI, "", nil)
				if err != nil {
					Write(500, FromError(err), w)
					return
				}
			default:
				if filepath.Ext(field.fileName) == ".asc" {
					signature = field.data
				} else {
					e := errors.New("Error: An unknown file found " + field.fileName)
					Write(400, FromError(e), w)
					return
				}
			}
		}

		// Sanity check, verify the checkSum on the file sent is the same
		// as after it is written to disk.
		if rp.CheckSum() != checkSum {
			e := errors.New("Error: CheckSum mismatch on requested plugin to load")
			Write(400, FromError(e), w)
			return
		}

		// check if one of TLS params (cert or key) has been provided; if not, skip the part related to TLS
		if hasTLS(certPath, keyPath) {
			// check if both of required params have been provided to setup TLS connection;
			// if not, return an appropriate error
			if isTLSEnabled(certPath, keyPath) {
				rp.SetTLSEnabled(true)
				rp.SetCACertPaths(caCertPaths)
				rp.SetCertPath(certPath)
				rp.SetKeyPath(keyPath)
			} else {
				e := errors.New("Error: TLS setup incomplete - Both plugin TLS certificate and the key are required")
				Write(500, FromError(e), w)
				return
			}
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
			rb := FromError(err)
			switch rb.ErrorMessage {
			case ErrPluginAlreadyLoaded:
				ec = 409
			default:
				ec = 500
			}
			Write(ec, rb, w)
			return
		}
		Write(201, catalogedPluginBody(r.Host, pl), w)
	}
}

func handleError(p string, w http.ResponseWriter) {
	if _, err := os.Stat(p); os.IsNotExist(err) {
		Write(500, FromError(err), w)
		return
	}
}

func hasTLS(cert, key string) bool {
	if cert == "" && key == "" {
		return false
	}
	return true
}

func isTLSEnabled(cert, key string) bool {
	if cert != "" && key != "" {
		return true
	}
	return false
}

func pluginParameters(p httprouter.Params) (string, string, int, map[string]interface{}, serror.SnapError) {
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
		return "", "", 0, nil, se
	}
	return plType, plName, int(plVersion), f, nil
}

func (s *apiV2) unloadPlugin(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	plType, plName, plVersion, f, se := pluginParameters(p)
	if se != nil {
		Write(400, FromSnapError(se), w)
		return
	}

	_, se = s.metricManager.Unload(&PluginParams{
		PName:    plName,
		PVersion: plVersion,
		PType:    plType,
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
		Write(statusCode, FromSnapError(se), w)
		return
	}
	Write(204, nil, w)
}

func (s *apiV2) getPlugins(w http.ResponseWriter, r *http.Request, params httprouter.Params) {

	// filter by plugin name or plugin type
	q := r.URL.Query()
	plName := q.Get("name")
	plType := q.Get("type")
	nbFilter := Btoi(plName != "") + Btoi(plType != "")

	var plugins []Plugin
	if _, detail := r.URL.Query()["running"]; detail {
		// get running plugins
		plugins = runningPluginsBody(r.Host, s.metricManager.AvailablePlugins())
	} else {
		// get plugins from the plugin catalog
		plugins = pluginCatalogBody(r.Host, s.metricManager.PluginCatalog())
	}

	filteredPlugins := []Plugin{}
	if nbFilter > 0 {
		for _, p := range plugins {
			if nbFilter == 1 && (p.Name == plName || p.Type == plType) || nbFilter == 2 && (p.Name == plName && p.Type == plType) {
				filteredPlugins = append(filteredPlugins, p)
			}
		}
	} else {
		filteredPlugins = plugins
	}
	Write(200, PluginsResponse{Plugins: filteredPlugins}, w)
}

func Btoi(b bool) int {
	if b {
		return 1
	}
	return 0
}

func pluginCatalogBody(host string, c []core.CatalogedPlugin) []Plugin {
	plugins := make([]Plugin, len(c))
	for i, p := range c {
		plugins[i] = catalogedPluginBody(host, p)
	}
	return plugins
}

func catalogedPluginBody(host string, c core.CatalogedPlugin) Plugin {
	return Plugin{
		Name:            c.Name(),
		Version:         c.Version(),
		Type:            c.TypeName(),
		Signed:          c.IsSigned(),
		Status:          c.Status(),
		LoadedTimestamp: c.LoadedTimestamp().Unix(),
		Href:            pluginURI(host, c),
	}
}

func runningPluginsBody(host string, c []core.AvailablePlugin) []Plugin {
	plugins := make([]Plugin, len(c))
	for i, p := range c {
		plugins[i] = Plugin{
			Name:             p.Name(),
			Version:          p.Version(),
			Type:             p.TypeName(),
			HitCount:         p.HitCount(),
			LastHitTimestamp: p.LastHit().Unix(),
			ID:               p.ID(),
			Href:             pluginURI(host, p),
			PprofPort:        p.Port(),
		}
	}
	return plugins
}

func pluginURI(host string, c core.Plugin) string {
	return fmt.Sprintf("%s://%s/%s/plugins/%s/%s/%d", protocolPrefix, host, version, c.TypeName(), c.Name(), c.Version())
}

func (s *apiV2) getPlugin(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	plType, plName, plVersion, f, se := pluginParameters(p)
	if se != nil {
		Write(400, FromSnapError(se), w)
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
		Write(404, FromSnapError(se), w)
		return
	}

	rd := r.FormValue("download")
	d, _ := strconv.ParseBool(rd)
	var configPolicy []PolicyTable
	if plugin.TypeName() == "processor" || plugin.TypeName() == "publisher" {
		rules := plugin.Policy().Get([]string{""}).RulesAsTable()
		configPolicy = make([]PolicyTable, 0, len(rules))
		for _, r := range rules {
			configPolicy = append(configPolicy, PolicyTable{
				Name:     r.Name,
				Type:     r.Type,
				Default:  r.Default,
				Required: r.Required,
				Minimum:  r.Minimum,
				Maximum:  r.Maximum,
			})
		}

	}

	if d {
		b, err := ioutil.ReadFile(plugin.PluginPath())
		if err != nil {
			f["plugin-path"] = plugin.PluginPath()
			se := serror.New(err, f)
			Write(500, FromSnapError(se), w)
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
			Write(500, FromSnapError(se), w)
			return
		}
		w.WriteHeader(200)
		return
	}
	pluginRet := Plugin{
		Name:            plugin.Name(),
		Version:         plugin.Version(),
		Type:            plugin.TypeName(),
		Signed:          plugin.IsSigned(),
		Status:          plugin.Status(),
		LoadedTimestamp: plugin.LoadedTimestamp().Unix(),
		Href:            pluginURI(r.Host, plugin),
		ConfigPolicy:    configPolicy,
	}
	Write(200, pluginRet, w)
}
