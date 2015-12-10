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

// const A list of constants
const (
	PluginsLoadedType  = "plugins_loaded"
	PluginUnloadedType = "plugin_unloaded"
	PluginListType     = "plugin_list_returned"
	PluginReturnedType = "plugin_returned"
)

// PluginsLoaded Successful response to the loading of a plugins
type PluginsLoaded struct {
	LoadedPlugins []LoadedPlugin `json:"loaded_plugins"`
}

// ResponseBodyMessage returns a string of response body message
func (p *PluginsLoaded) ResponseBodyMessage() string {
	s := "Plugins loaded: "
	l := make([]string, len(p.LoadedPlugins))
	for i, pl := range p.LoadedPlugins {
		l[i] = fmt.Sprintf("%s(%s v%d)", pl.Name, pl.Type, pl.Version)
	}
	s += strings.Join(l, ", ")
	return s
}

// ResponseBodyType returns a string of response body type
func (p *PluginsLoaded) ResponseBodyType() string {
	return PluginsLoadedType
}

// PluginUnloaded Successful response to the unloading of a plugin
type PluginUnloaded struct {
	Name    string `json:"name"`
	Version int    `json:"version"`
	Type    string `json:"type"`
}

// ResponseBodyMessage returns a string of body response
func (u *PluginUnloaded) ResponseBodyMessage() string {
	return fmt.Sprintf("Plugin successfuly unloaded (%sv%d)", u.Name, u.Version)
}

// ResponseBodyType returns a string of response type
func (u *PluginUnloaded) ResponseBodyType() string {
	return PluginUnloadedType
}

// PluginList struct type
type PluginList struct {
	LoadedPlugins    []LoadedPlugin    `json:"loaded_plugins,omitempty"`
	AvailablePlugins []AvailablePlugin `json:"available_plugins,omitempty"`
}

// ResponseBodyMessage return a string response body message
func (p *PluginList) ResponseBodyMessage() string {
	return "Plugin list returned"
}

// ResponseBodyType returns a string response of PluginListType
func (p *PluginList) ResponseBodyType() string {
	return PluginListType
}

// PluginReturned type
type PluginReturned LoadedPlugin

// ResponseBodyMessage returns a string response
func (p *PluginReturned) ResponseBodyMessage() string {
	return "Plugin returned"
}

// ResponseBodyType returns a string response of PluginReturnedType
func (p *PluginReturned) ResponseBodyType() string {
	return PluginReturnedType
}

// LoadedPlugin struct type
type LoadedPlugin struct {
	Name            string `json:"name"`
	Version         int    `json:"version"`
	Type            string `json:"type"`
	Signed          bool   `json:"signed"`
	Status          string `json:"status"`
	LoadedTimestamp int64  `json:"loaded_timestamp"`
	Href            string `json:"href"`
}

// AvailablePlugin struct type
type AvailablePlugin struct {
	Name             string `json:"name"`
	Version          int    `json:"version"`
	Type             string `json:"type"`
	HitCount         int    `json:"hitcount"`
	LastHitTimestamp int64  `json:"last_hit_timestamp"`
	ID               uint32 `json:"id"`
	Href             string `json:"href"`
}
