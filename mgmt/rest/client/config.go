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
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/intelsdi-x/snap/core/ctypes"
	"github.com/intelsdi-x/snap/mgmt/rest/v1/rbody"
)

// GetPluginConfig retrieves the merged plugin config given the type of plugin,
// name and version.  If plugin type, name and version are all empty strings
// the plugin config for "all" plugins will be returned.  If the plugin type is
// provided and the name and version are empty strings the config for that plugin
// type will be returned.  So on and so forth for the rest of the arguments.
func (c *Client) GetPluginConfig(pluginType, name, version string) *GetPluginConfigResult {
	r := &GetPluginConfigResult{}
	resp, err := c.do("GET", fmt.Sprintf("/plugins/%s/%s/%s/config", pluginType, url.QueryEscape(name), version), ContentTypeJSON)
	if err != nil {
		r.Err = err
		return r
	}

	switch resp.Meta.Type {
	case rbody.PluginConfigItemType:
		// Success
		config := resp.Body.(*rbody.PluginConfigItem)
		r = &GetPluginConfigResult{config, nil}
	case rbody.ErrorType:
		r.Err = resp.Body.(*rbody.Error)
	default:
		r.Err = ErrAPIResponseMetaType
	}
	return r
}

// SetPluginConfig sets the plugin config given the type, name and version
// of a plugin.  Like GetPluginConfig if the type, name and version are all
// empty strings the plugin config is set for all plugins.  When config data
// is set it is merged with the existing data if present.
func (c *Client) SetPluginConfig(pluginType, name, version string, key string, value ctypes.ConfigValue) *SetPluginConfigResult {
	r := &SetPluginConfigResult{}
	b, err := json.Marshal(map[string]ctypes.ConfigValue{key: value})
	if err != nil {
		r.Err = err
		return r
	}
	resp, err := c.do("PUT", fmt.Sprintf("/plugins/%s/%s/%s/config", pluginType, url.QueryEscape(name), version), ContentTypeJSON, b)
	if err != nil {
		r.Err = err
		return r
	}

	switch resp.Meta.Type {
	case rbody.SetPluginConfigItemType:
		// Success
		config := resp.Body.(*rbody.SetPluginConfigItem)
		r = &SetPluginConfigResult{config, nil}
	case rbody.ErrorType:
		r.Err = resp.Body.(*rbody.Error)
	default:
		r.Err = ErrAPIResponseMetaType
	}
	return r
}

// DeletePluginConfig removes the plugin config item given the plugin type, name
// and version.
func (c *Client) DeletePluginConfig(pluginType, name, version string, key string) *DeletePluginConfigResult {
	r := &DeletePluginConfigResult{}
	b, err := json.Marshal([]string{key})
	if err != nil {
		r.Err = err
		return r
	}
	resp, err := c.do("DELETE", fmt.Sprintf("/plugins/%s/%s/%s/config", pluginType, url.QueryEscape(name), version), ContentTypeJSON, b)
	if err != nil {
		r.Err = err
		return r
	}

	switch resp.Meta.Type {
	case rbody.DeletePluginConfigItemType:
		// Success
		config := resp.Body.(*rbody.DeletePluginConfigItem)
		r = &DeletePluginConfigResult{config, nil}
	case rbody.ErrorType:
		r.Err = resp.Body.(*rbody.Error)
	default:
		r.Err = ErrAPIResponseMetaType
	}
	return r
}

// GetPluginConfigResult is the response from snap/client on a GetPluginConfig call.
type GetPluginConfigResult struct {
	*rbody.PluginConfigItem
	Err error
}

// SetPluginConfigResult is the response from snap/client on a SetPluginConfig call.
type SetPluginConfigResult struct {
	*rbody.SetPluginConfigItem
	Err error
}

// DeletePluginConfigResult is the response from snap/client on a DeletePluginConfig call.
type DeletePluginConfigResult struct {
	*rbody.DeletePluginConfigItem
	Err error
}
