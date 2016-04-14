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
	"crypto/sha256"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"mime"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"

	"golang.org/x/net/context"

	"github.com/julienschmidt/httprouter"

	"github.com/intelsdi-x/snap/control"
	"github.com/intelsdi-x/snap/control/rpc"
	"github.com/intelsdi-x/snap/core/serror"
	"github.com/intelsdi-x/snap/internal/common"
	"github.com/intelsdi-x/snap/mgmt/rest/rbody"
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
		var pluginName string
		var signature []byte
		var checkSum [sha256.Size]byte
		lp := &rbody.PluginsLoaded{}
		lp.LoadedPlugins = make([]rbody.LoadedPlugin, 0)
		mr := multipart.NewReader(r.Body, params["boundary"])
		var pluginFile []byte
		i := 0
		for {
			var b []byte
			p, err := mr.NextPart()
			if err == io.EOF {
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
					respond(500, rbody.FromError(e), w)
					return
				}
				// Get filename, bytes, and checksum for plugin to pass over gRPC to load
				pluginName = p.FileName()
				pluginFile = b
				checkSum = sha256.Sum256(b)
			case i == 1:
				if filepath.Ext(p.FileName()) == ".asc" {
					signature = b
				} else {
					e := errors.New("Error: second file passed was not a signature file")
					respond(500, rbody.FromError(e), w)
					return
				}
			case i == 2:
				e := errors.New("Error: More than two files passed to the load plugin api")
				respond(500, rbody.FromError(e), w)
				return
			}
			i++
		}
		restLogger.Info("Loading plugin: ", pluginName)
		arg := &rpc.PluginRequest{
			Name:       pluginName,
			CheckSum:   checkSum[:],
			Signature:  signature,
			PluginFile: pluginFile,
		}
		reply, err := s.mm.Load(context.Background(), arg)
		if err != nil {
			var code int
			switch err.Error() {
			case control.ErrPluginAlreadyLoaded.Error():
				code = 409
			default:
				code = 500
			}
			respond(code, rbody.FromError(err), w)
			return
		}
		plugin, _ := rpc.ReplyToLoadedPlugin(reply)
		loadedPlugin := rbody.LoadedPlugin{
			Name:            plugin.Name,
			Version:         plugin.Version,
			Type:            plugin.TypeName,
			Signed:          plugin.IsSigned,
			Status:          plugin.Status,
			LoadedTimestamp: plugin.LoadedTimestamp.Unix(),
			Href:            pluginURI(r.Host, plugin.TypeName, plugin.Name, plugin.Version),
		}
		lp.LoadedPlugins = append(lp.LoadedPlugins, loadedPlugin)
		respond(201, lp, w)
	}
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
		se := serror.New(errors.New("invalid version"))
		se.SetFields(f)
		respond(400, rbody.FromSnapError(se), w)
		return
	}

	if plName == "" {
		se := serror.New(errors.New("missing plugin name"))
		se.SetFields(f)
		respond(400, rbody.FromSnapError(se), w)
		return
	}
	if plType == "" {
		se := serror.New(errors.New("missing plugin type"))
		se.SetFields(f)
		respond(400, rbody.FromSnapError(se), w)
		return
	}

	restLogger.Info("Unloading plugin: ", plName, plVersion, plType)
	arg := &rpc.UnloadPluginRequest{
		Name:       plName,
		Version:    int32(plVersion),
		PluginType: plType,
	}
	reply, err := s.mm.Unload(context.Background(), arg)
	//rpc error
	if err != nil {
		respond(500, rbody.FromError(err), w)
		return
	}
	pr := &rbody.PluginUnloaded{
		Name:    reply.Name,
		Version: int(reply.Version),
		Type:    reply.TypeName,
	}
	respond(200, pr, w)
}

func (s *Server) getPlugins(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	var detail bool
	for k := range r.URL.Query() {
		if k == "details" {
			detail = true
		}
	}
	plName := params.ByName("name")
	plType := params.ByName("type")

	plugins, err := getPlugins(s.mm, detail, r.Host, plName, plType)
	if err != nil {
		respond(500, rbody.FromError(err), w)
		return
	}
	respond(200, plugins, w)
}

func getPlugins(mm managesMetrics, detail bool, host, plName, plType string) (*rbody.PluginList, error) {

	plugins := new(rbody.PluginList)

	restLogger.Info("Getting plugin catalog")
	arg := &common.Empty{}
	reply, err := mm.PluginCatalog(context.Background(), arg)
	if err != nil {
		return nil, err
	}
	plugins.LoadedPlugins = make([]rbody.LoadedPlugin, len(reply.Plugins))
	for idx, plugin := range reply.Plugins {
		lp, err := rpc.ReplyToLoadedPlugin(plugin)
		if err != nil {
			return nil, err
		}
		plugins.LoadedPlugins[idx] = rbody.LoadedPlugin{
			Name:            lp.Name,
			Version:         int(lp.Version),
			Type:            lp.TypeName,
			Signed:          lp.IsSigned,
			Status:          lp.Status,
			LoadedTimestamp: lp.LoadedTimestamp.Unix(),
			Href:            pluginURI(host, lp.TypeName, lp.Name, lp.Version),
		}
	}
	if detail {
		reply, err := mm.AvailablePlugins(context.Background(), arg)
		if err != nil {
			return nil, err
		}
		plugins.AvailablePlugins = make([]rbody.AvailablePlugin, len(reply.Plugins))
		for i, plugin := range reply.Plugins {
			p := rpc.ReplyToAvailablePlugin(plugin)
			plugins.AvailablePlugins[i] = rbody.AvailablePlugin{
				Name:             p.Name,
				Version:          p.Version,
				Type:             p.TypeName,
				HitCount:         p.HitCount,
				LastHitTimestamp: p.LastHit.Unix(),
				ID:               p.ID,
				Href:             pluginURI(host, p.TypeName, p.Name, p.Version),
			}
		}
	}

	filteredPlugins := new(rbody.PluginList)

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

	filteredPlugins = new(rbody.PluginList)

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

	return plugins, nil
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
		se := serror.New(errors.New("invalid version"))
		se.SetFields(f)
		respond(400, rbody.FromSnapError(se), w)
		return
	}

	if plName == "" {
		se := serror.New(errors.New("missing plugin name"))
		se.SetFields(f)
		respond(400, rbody.FromSnapError(se), w)
		return
	}
	if plType == "" {
		se := serror.New(errors.New("missing plugin type"))
		se.SetFields(f)
		respond(400, rbody.FromSnapError(se), w)
		return
	}
	rd := r.FormValue("download")
	plDownload, _ := strconv.ParseBool(rd)
	arg := &rpc.GetPluginRequest{
		Name:     plName,
		Version:  int32(plVersion),
		Type:     plType,
		Download: plDownload,
	}
	reply, err := s.mm.GetPlugin(context.Background(), arg)
	if err != nil {
		se := serror.New(ErrPluginNotFound, f)
		respond(404, rbody.FromSnapError(se), w)
		return
	}
	lp, err := rpc.ReplyToLoadedPlugin(reply.Plugin)
	if err != nil {
		respond(500, rbody.FromError(err), w)
		return
	}
	var configPolicy []rbody.PolicyTable
	if lp.TypeName == "processor" || lp.TypeName == "publisher" {
		rules := lp.ConfigPolicy.Get([]string{""}).RulesAsTable()
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

	if plDownload {

		w.Header().Set("Content-Encoding", "gzip")
		gz := gzip.NewWriter(w)
		defer gz.Close()
		_, err := gz.Write(reply.PluginBytes)
		if err != nil {
			se := serror.New(err, f)
			respond(500, rbody.FromSnapError(se), w)
			return
		}
		return
	} else {
		pluginRet := &rbody.PluginReturned{
			Name:            lp.Name,
			Version:         int(lp.Version),
			Type:            lp.TypeName,
			Signed:          lp.IsSigned,
			Status:          lp.Status,
			LoadedTimestamp: lp.LoadedTimestamp.Unix(),
			Href:            pluginURI(r.Host, lp.TypeName, lp.Name, lp.Version),
			ConfigPolicy:    configPolicy,
		}
		respond(200, pluginRet, w)
	}
}

func pluginURI(host, typeName, name string, version int) string {
	return fmt.Sprintf("%s://%s/v1/plugins/%s/%s/%d", protocolPrefix, host, typeName, name, version)
}
