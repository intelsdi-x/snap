// +build legacy small medium large

/*
http://www.apache.org/licenses/LICENSE-2.0.txt

Copyright 2016 Intel Corporation

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

package mock

import (
	"github.com/intelsdi-x/snap/core"
	"github.com/intelsdi-x/snap/core/cdata"
	"github.com/intelsdi-x/snap/core/ctypes"
)

var mockConfig *cdata.ConfigDataNode

func init() {
	mockConfig = cdata.NewNode()
	mockConfig.AddItem("User", ctypes.ConfigValueStr{Value: "KELLY"})
	mockConfig.AddItem("Port", ctypes.ConfigValueInt{Value: 2})
}

type MockConfigManager struct{}

func (MockConfigManager) GetPluginConfigDataNode(core.PluginType, string, int) cdata.ConfigDataNode {
	return *mockConfig
}
func (MockConfigManager) GetPluginConfigDataNodeAll() cdata.ConfigDataNode {
	return *mockConfig
}
func (MockConfigManager) MergePluginConfigDataNode(
	pluginType core.PluginType, name string, ver int, cdn *cdata.ConfigDataNode) cdata.ConfigDataNode {
	return *cdn
}
func (MockConfigManager) MergePluginConfigDataNodeAll(cdn *cdata.ConfigDataNode) cdata.ConfigDataNode {
	return cdata.ConfigDataNode{}
}
func (MockConfigManager) DeletePluginConfigDataNodeField(
	pluginType core.PluginType, name string, ver int, fields ...string) cdata.ConfigDataNode {
	for _, field := range fields {
		mockConfig.DeleteItem(field)

	}
	return *mockConfig
}

func (MockConfigManager) DeletePluginConfigDataNodeFieldAll(fields ...string) cdata.ConfigDataNode {
	for _, field := range fields {
		mockConfig.DeleteItem(field)

	}
	return *mockConfig
}

// These constants are the expected plugin config responses from running
// rest_v2_test.go on the plugin config routes found in mgmt/rest/server.go
const (
	SET_PLUGIN_CONFIG_ITEM = `{
  "user": "Jane"
}
`

	GET_PLUGIN_CONFIG_ITEM = `{
  "Port": 2,
  "User": "KELLY"
}
`

	DELETE_PLUGIN_CONFIG_ITEM = `{
  "Port": 2,
  "User": "KELLY"
}
`
)
