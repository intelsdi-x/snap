package api

import (
	"github.com/intelsdi-x/snap/core"
	"github.com/intelsdi-x/snap/core/cdata"
)

type Config interface {
	GetPluginConfigDataNode(core.PluginType, string, int) cdata.ConfigDataNode
	GetPluginConfigDataNodeAll() cdata.ConfigDataNode
	MergePluginConfigDataNode(pluginType core.PluginType, name string, ver int, cdn *cdata.ConfigDataNode) cdata.ConfigDataNode
	MergePluginConfigDataNodeAll(cdn *cdata.ConfigDataNode) cdata.ConfigDataNode
	DeletePluginConfigDataNodeField(pluginType core.PluginType, name string, ver int, fields ...string) cdata.ConfigDataNode
	DeletePluginConfigDataNodeFieldAll(fields ...string) cdata.ConfigDataNode
}
