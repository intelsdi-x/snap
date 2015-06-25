package pulse

import (
	"fmt"
	"net/url"
	"time"

	"github.com/intelsdi-x/pulse/mgmt/rest/rbody"
)

type Plugin struct {
	Name            string     `json:"name"`
	Version         int        `json:"version,omitempty"`
	TypeName        string     `json:"type"`
	Status          string     `json:"status,omitempty"`
	LoadedTimestamp *time.Time `json:"loaded_timestamp,omitempty"`
	HitCount        int        `json:"hit_count,omitempty"`
	LastHit         *time.Time `json:"last_hit,omitempty"`
}

// TODO, this should RETURN the plugin that was loaded...
func (c *Client) LoadPlugin(p string) error {
	resp, err := c.pluginUploadRequest(p)
	if err != nil {
		return err
	}

	switch resp.Meta.Type {
	// TODO change this to concrete const type when Joel adds it
	case "plugin_load":
		//
	case rbody.ErrorType:
		return resp.Body.(*rbody.Error)
	default:
		return ErrAPIResponseMetaType
	}
	return ErrUnknown
}

func (c *Client) UnloadPlugin(name string, version int) *UnloadPluginResult {
	r := &UnloadPluginResult{
		Name:    name,
		Version: version,
	}
	resp, err := c.do("DELETE", fmt.Sprintf("/plugins/%s/%d", url.QueryEscape(name), version), ContentTypeJSON)
	if err != nil {
		r.Err = err
		return r
	}

	switch resp.Meta.Type {
	// TODO change this to concrete const type when Joel adds it
	case rbody.PluginUnloadedType:
		// Success
	case rbody.ErrorType:
		r.Err = resp.Body.(*rbody.Error)
	default:
		r.Err = ErrAPIResponseMetaType
	}
	return r
}

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
	case rbody.PluginListReturnedType:
		// Success
		b := resp.Body.(*rbody.PluginListReturned)
		r.LoadedPlugins = b.LoadedPlugins
		r.AvailablePlugins = b.AvailablePlugins
		return r
	case rbody.ErrorType:
		r.Err = resp.Body.(*rbody.Error)
	default:
		r.Err = ErrAPIResponseMetaType
	}
	return r
}

type GetPluginsResult struct {
	LoadedPlugins    []rbody.LoadedPlugin
	AvailablePlugins []rbody.AvailablePlugin
	Err              error
}

// UnloadPluginResponse is the response from pulse/client on an UnloadPlugin call.
type UnloadPluginResult struct {
	Name    string
	Version int
	Err     error
}
