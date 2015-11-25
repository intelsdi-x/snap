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

const (
	PluginConfigItemType       = "config_plugin_item_returned"
	SetPluginConfigItemType    = "config_plugin_item_created"
	DeletePluginConfigItemType = "config_plugin_item_deleted"
)

type DeletePluginConfigItem PluginConfigItem

func (t *DeletePluginConfigItem) ResponseBodyMessage() string {
	return "Plugin config item field(s) deleted"
}

func (t *DeletePluginConfigItem) ResponseBodyType() string {
	return DeletePluginConfigItemType
}

type SetPluginConfigItem PluginConfigItem

func (t *SetPluginConfigItem) ResponseBodyMessage() string {
	return "Plugin config item(s) set"
}

func (t *SetPluginConfigItem) ResponseBodyType() string {
	return SetPluginConfigItemType
}

type PluginConfigItem struct {
	cdata.ConfigDataNode
}

func (t *PluginConfigItem) ResponseBodyMessage() string {
	return "Plugin config item retrieved"
}

func (t *PluginConfigItem) ResponseBodyType() string {
	return PluginConfigItemType
}
