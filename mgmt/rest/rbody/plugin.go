package rbody

import (
	"fmt"
	"strings"
)

const (
	PluginsLoadedType      = "plugins_loaded"
	PluginUnloadedType     = "plugin_unloaded"
	PluginListReturnedType = "plugin_list_returned"
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
	return fmt.Sprintf("Plugin successfuly unloaded (%sv%d)", u.Name, u.Version)
}

func (u *PluginUnloaded) ResponseBodyType() string {
	return PluginUnloadedType
}

type PluginListReturned struct {
	LoadedPlugins    []LoadedPlugin    `json:"loaded_plugins,omitempty"`
	AvailablePlugins []AvailablePlugin `json:"available_plugins,omitempty"`
}

func (p *PluginListReturned) ResponseBodyMessage() string {
	return "Plugin list retrieved"
}

func (p *PluginListReturned) ResponseBodyType() string {
	return PluginListReturnedType
}

type LoadedPlugin struct {
	Name            string `json:"name"`
	Version         int    `json:"version"`
	Type            string `json:"type"`
	Status          string `json:"status"`
	LoadedTimestamp int64  `json:"loaded_timestamp"`
}

type AvailablePlugin struct {
	Name             string `json:"name"`
	Version          int    `json:"version"`
	Type             string `json:"type"`
	HitCount         int    `json:"hitcount"`
	LastHitTimestamp int64  `json:"last_hit_timestamp"`
	ID               int    `json:"id"`
}
