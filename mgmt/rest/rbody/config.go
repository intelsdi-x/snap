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

import "github.com/intelsdi-x/snap/core/cdata"

// const a list of constants
const (
	PluginConfigItemType       = "config_plugin_item_returned"
	SetPluginConfigItemType    = "config_plugin_item_created"
	DeletePluginConfigItemType = "config_plugin_item_deleted"
)

// DeletePluginConfigItem has PluginConfigItem type
type DeletePluginConfigItem PluginConfigItem

// ResponseBodyMessage returns a string message
func (t *DeletePluginConfigItem) ResponseBodyMessage() string {
	return "Plugin config item field(s) deleted"
}

// ResponseBodyType returns the DeletePluginConfigItemType
func (t *DeletePluginConfigItem) ResponseBodyType() string {
	return DeletePluginConfigItemType
}

// SetPluginConfigItem type
type SetPluginConfigItem PluginConfigItem

// ResponseBodyMessage returns a string response message
func (t *SetPluginConfigItem) ResponseBodyMessage() string {
	return "Plugin config item(s) set"
}

// ResponseBodyType returns a stringg type
func (t *SetPluginConfigItem) ResponseBodyType() string {
	return SetPluginConfigItemType
}

// PluginConfigItem struct type
type PluginConfigItem struct {
	cdata.ConfigDataNode
}

// ResponseBodyMessage returns a string message
func (t *PluginConfigItem) ResponseBodyMessage() string {
	return "Plugin config item retrieved"
}

// ResponseBodyType returns a string message
func (t *PluginConfigItem) ResponseBodyType() string {
	return PluginConfigItemType
}
