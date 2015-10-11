package control

import (
	"fmt"

	"github.com/intelsdi-x/pulse/control/plugin"
	"github.com/intelsdi-x/pulse/core/cdata"
	"github.com/intelsdi-x/pulse/core/ctypes"
)

// type configData map[string]*cdata.ConfigDataNode

type pluginConfig struct {
	all         *cdata.ConfigDataNode
	collector   map[string]*pluginConfigItem
	publisher   map[string]*pluginConfigItem
	processor   map[string]*pluginConfigItem
	pluginCache map[string]*cdata.ConfigDataNode
}

type pluginConfigItem struct {
	*cdata.ConfigDataNode
	versions map[int]*cdata.ConfigDataNode
}

type config struct {
	control   *cdata.ConfigDataNode
	scheduler *cdata.ConfigDataNode
	plugins   *pluginConfig
}

func newConfig() *config {
	return &config{
		control:   cdata.NewNode(),
		scheduler: cdata.NewNode(),
		plugins:   newPluginConfig(),
	}
}

func newPluginConfig() *pluginConfig {
	return &pluginConfig{
		all:         cdata.NewNode(),
		collector:   make(map[string]*pluginConfigItem),
		processor:   make(map[string]*pluginConfigItem),
		publisher:   make(map[string]*pluginConfigItem),
		pluginCache: make(map[string]*cdata.ConfigDataNode),
	}
}

func newPluginConfigItem(opts ...pluginConfigOpt) *pluginConfigItem {
	p := &pluginConfigItem{
		ConfigDataNode: cdata.NewNode(),
		versions:       make(map[int]*cdata.ConfigDataNode),
	}

	for _, opt := range opts {
		opt(p)
	}

	return p
}

type pluginConfigOpt func(*pluginConfigItem)

func optAddPluginConfigItem(key string, value ctypes.ConfigValue) pluginConfigOpt {
	return func(p *pluginConfigItem) {
		p.AddItem(key, value)
	}
}

func (p *pluginConfig) get(pluginType plugin.PluginType, name string, ver int) *cdata.ConfigDataNode {
	// check cache
	key := fmt.Sprintf("%d:%s:%d", pluginType, name, ver)
	if res, ok := p.pluginCache[key]; ok {
		return res
	}

	//todo process/interpolate values

	p.pluginCache[key] = cdata.NewNode()
	p.pluginCache[key].Merge(p.all)

	// check for plugin config
	switch pluginType {
	case plugin.CollectorPluginType:
		if res, ok := p.collector[name]; ok {
			p.pluginCache[key].Merge(res.ConfigDataNode)
			if res2, ok2 := res.versions[ver]; ok2 {
				p.pluginCache[key].Merge(res2)
			}
		}
	case plugin.ProcessorPluginType:
		if res, ok := p.processor[name]; ok {
			p.pluginCache[key].Merge(res.ConfigDataNode)
			if res2, ok2 := res.versions[ver]; ok2 {
				p.pluginCache[key].Merge(res2)
			}
		}
	case plugin.PublisherPluginType:
		if res, ok := p.publisher[name]; ok {
			p.pluginCache[key].Merge(res.ConfigDataNode)
			if res2, ok2 := res.versions[ver]; ok2 {
				p.pluginCache[key].Merge(res2)
			}
		}
	}

	return p.pluginCache[key]
}
