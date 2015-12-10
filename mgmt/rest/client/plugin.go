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

package client

import (
	"fmt"
	"net/url"
	"time"

	"github.com/intelsdi-x/snap/core/serror"
	"github.com/intelsdi-x/snap/mgmt/rest/rbody"
)

// LoadPlugin loads plugins for the given plugin names.
// A slide of loaded plugins returns if succeeded. Otherwise, an error is returned.
func (c *Client) LoadPlugin(p []string) *LoadPluginResult {
	r := new(LoadPluginResult)
	resp, err := c.pluginUploadRequest(p)
	if err != nil {
		r.Err = serror.New(err)
		return r
	}

	switch resp.Meta.Type {
	case rbody.PluginsLoadedType:
		pl := resp.Body.(*rbody.PluginsLoaded)
		r.LoadedPlugins = convertLoadedPlugins(pl.LoadedPlugins)
	case rbody.ErrorType:
		f := resp.Body.(*rbody.Error).Fields
		fields := make(map[string]interface{})
		for k, v := range f {
			fields[k] = v
		}
		r.Err = serror.New(resp.Body.(*rbody.Error), fields)
	default:
		r.Err = serror.New(ErrAPIResponseMetaType)
	}
	return r
}

// UnloadPlugin unloads a plugin given plugin type, name, and version through an HTTP DELETE request.
// The unloaded plugin returns if succeeded. Otherwise, an error is returned.
func (c *Client) UnloadPlugin(pluginType, name string, version int) *UnloadPluginResult {
	r := &UnloadPluginResult{}
	resp, err := c.do("DELETE", fmt.Sprintf("/plugins/%s/%s/%d", pluginType, url.QueryEscape(name), version), ContentTypeJSON)
	if err != nil {
		r.Err = err
		return r
	}

	switch resp.Meta.Type {
	case rbody.PluginUnloadedType:
		// Success
		up := resp.Body.(*rbody.PluginUnloaded)
		r = &UnloadPluginResult{up, nil}
	case rbody.ErrorType:
		r.Err = resp.Body.(*rbody.Error)
	default:
		r.Err = ErrAPIResponseMetaType
	}
	return r
}

// GetPlugins returns the loaded and available plugins through an HTTP GET request.
// By specifying the details flag to tweak output info. An error returns if it failed.
func (c *Client) GetPlugins(details bool) *GetPluginsResult {
	r := &GetPluginsResult{}

	var path string
	if details {
		path = "/plugins?details"
	} else {
		path = "/plugins"
	}

	resp, err := c.do("GET", path, ContentTypeJSON)
	if err != nil {
		r.Err = err
		return r
	}

	switch resp.Meta.Type {
	// TODO change this to concrete const type when Joel adds it
	case rbody.PluginListType:
		// Success
		b := resp.Body.(*rbody.PluginList)
		r.LoadedPlugins = convertLoadedPlugins(b.LoadedPlugins)
		r.AvailablePlugins = convertAvailablePlugins(b.AvailablePlugins)
		return r
	case rbody.ErrorType:
		r.Err = resp.Body.(*rbody.Error)
	default:
		r.Err = ErrAPIResponseMetaType
	}
	return r
}

// GetPluginsResult is the response from snap/client on a GetPlugins call.
type GetPluginsResult struct {
	LoadedPlugins    []LoadedPlugin
	AvailablePlugins []AvailablePlugin
	Err              error
}

// LoadPluginResult is the response from snap/client on a LoadPlugin call.
type LoadPluginResult struct {
	LoadedPlugins []LoadedPlugin
	Err           serror.SnapError
}

// UnloadPluginResponse is the response from snap/client on an UnloadPlugin call.
type UnloadPluginResult struct {
	*rbody.PluginUnloaded
	Err error
}

// We wrap this so we can provide some functionality (like LoadedTime)
type LoadedPlugin struct {
	*rbody.LoadedPlugin
}

// LoadedTime returns a unix time.
func (l *LoadedPlugin) LoadedTime() time.Time {
	return time.Unix(l.LoadedTimestamp, 0)
}

// The wrapper for AvailablePlugin struct defined inside rbody package.
type AvailablePlugin struct {
	*rbody.AvailablePlugin
}

func convertLoadedPlugins(r []rbody.LoadedPlugin) []LoadedPlugin {
	lp := make([]LoadedPlugin, len(r))
	for i := range r {
		lp[i] = LoadedPlugin{&r[i]}
	}
	return lp
}

func convertAvailablePlugins(r []rbody.AvailablePlugin) []AvailablePlugin {
	lp := make([]AvailablePlugin, len(r))
	for i := range r {
		lp[i] = AvailablePlugin{&r[i]}
	}
	return lp
}
