package rbody

import (
	"fmt"
)

// Successful response to the loading of a plugin
type LoadPlugin struct {
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
	return "plugin_unloaded"
}

type PluginListReturned struct {
	LoadedPlugins    []LoadedPlugin    `json:"loaded_plugins,omitempty"`
	AvailablePlugins []AvailablePlugin `json:"available_plugins,omitempty"`
}

func (p *PluginListReturned) ResponseBodyMessage() string {
	return "Plugin list retrieved"
}

func (p *PluginListReturned) ResponseBodyType() string {
	return "plugin_list_returned"
}

type LoadedPlugin struct {
	Name            string `json:"name"`
	Version         int    `json:"version"`
	Type            string `json:"type"`
	Status          string `json:"status"`
	LoadedTimestamp int    `json:"loaded_timestamp"`
}

type AvailablePlugin struct {
	Name             string `json:"name"`
	Version          int    `json:"version"`
	Type             string `json:"type"`
	HitCount         int    `json:"hitcount"`
	LastHitTimestamp int    `json:"last_hit_timestamp"`
	ID               int    `json:"id"`
}
