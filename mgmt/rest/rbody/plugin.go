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

package rbody

import (
	"fmt"
	"strings"
)

const (
	PluginsLoadedType  = "plugins_loaded"
	PluginUnloadedType = "plugin_unloaded"
	PluginListType     = "plugin_list_returned"
	PluginReturnedType = "plugin_returned"
)

// Successful response to the loading of a plugins
type PluginsLoaded struct {
	LoadedPlugins []LoadedPlugin `json:"loaded_plugins"`
}

func (p *PluginsLoaded) ResponseBodyMessage() string {
	s := "Plugins loaded: "
	l := make([]string, len(p.LoadedPlugins))
	for i, pl := range p.LoadedPlugins {
		l[i] = fmt.Sprintf("%s(%s v%d)", pl.Name, pl.Type, pl.Version)
	}
	s += strings.Join(l, ", ")
	return s
}

func (p *PluginsLoaded) ResponseBodyType() string {
	return PluginsLoadedType
}

// Successful response to the unloading of a plugin
type PluginUnloaded struct {
	Name    string `json:"name"`
	Version int    `json:"version"`
	Type    string `json:"type"`
}

func (u *PluginUnloaded) ResponseBodyMessage() string {
	return fmt.Sprintf("Plugin successfully unloaded (%sv%d)", u.Name, u.Version)
}

func (u *PluginUnloaded) ResponseBodyType() string {
	return PluginUnloadedType
}

type PluginList struct {
	LoadedPlugins    []LoadedPlugin    `json:"loaded_plugins,omitempty"`
	AvailablePlugins []AvailablePlugin `json:"available_plugins,omitempty"`
}

func (p *PluginList) ResponseBodyMessage() string {
	return "Plugin list returned"
}

func (p *PluginList) ResponseBodyType() string {
	return PluginListType
}

type PluginReturned LoadedPlugin

func (p *PluginReturned) ResponseBodyMessage() string {
	return "Plugin returned"
}

func (p *PluginReturned) ResponseBodyType() string {
	return PluginReturnedType
}

type LoadedPlugin struct {
	Name            string        `json:"name"`
	Version         int           `json:"version"`
	Type            string        `json:"type"`
	Signed          bool          `json:"signed"`
	Status          string        `json:"status"`
	LoadedTimestamp int64         `json:"loaded_timestamp"`
	Href            string        `json:"href"`
	ConfigPolicy    []PolicyTable `json:"policy,omitempty"`
}

type AvailablePlugin struct {
	Name             string `json:"name"`
	Version          int    `json:"version"`
	Type             string `json:"type"`
	HitCount         int    `json:"hitcount"`
	LastHitTimestamp int64  `json:"last_hit_timestamp"`
	ID               uint32 `json:"id"`
	Href             string `json:"href"`
	PprofPort        string `json:"pprof_port"`
}
