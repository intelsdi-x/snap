package control

import (
	"bytes"
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"

	"github.com/intelsdi-x/pulse/control/plugin"
	"github.com/intelsdi-x/pulse/core/cdata"
	"github.com/intelsdi-x/pulse/core/ctypes"
)

// type configData map[string]*cdata.ConfigDataNode

type pluginConfig struct {
	All         *cdata.ConfigDataNode        `json:"all"`
	Collector   map[string]*pluginConfigItem `json:"collector"`
	Publisher   map[string]*pluginConfigItem `json:"publisher"`
	Processor   map[string]*pluginConfigItem `json:"processor"`
	pluginCache map[string]*cdata.ConfigDataNode
}

type pluginConfigItem struct {
	*cdata.ConfigDataNode
	Versions map[int]*cdata.ConfigDataNode `json:"versions"`
}

type config struct {
	Control   *cdata.ConfigDataNode `json:"control"`
	Scheduler *cdata.ConfigDataNode `json:"scheduler"`
	Plugins   *pluginConfig         `json:"plugins"`
}

func NewConfig() *config {
	return &config{
		Control:   cdata.NewNode(),
		Scheduler: cdata.NewNode(),
		Plugins:   newPluginConfig(),
	}
}

func newPluginConfig() *pluginConfig {
	return &pluginConfig{
		All:         cdata.NewNode(),
		Collector:   make(map[string]*pluginConfigItem),
		Processor:   make(map[string]*pluginConfigItem),
		Publisher:   make(map[string]*pluginConfigItem),
		pluginCache: make(map[string]*cdata.ConfigDataNode),
	}
}

// UnmarshalJSON unmarshals valid json into pluginConfig.  An example Config
// github.com/intelsdi-x/pulse/examples/configs/pulse-config-sample.
func (p *pluginConfig) UnmarshalJSON(data []byte) error {
	t := map[string]interface{}{}
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.UseNumber()
	if err := dec.Decode(&t); err != nil {
		return err
	}

	if v, ok := t["all"]; ok {
		jv, err := json.Marshal(v)
		if err != nil {
			return err
		}
		cdn := &cdata.ConfigDataNode{}
		dec = json.NewDecoder(bytes.NewReader(jv))
		dec.UseNumber()
		if err := dec.Decode(&cdn); err != nil {
			return err
		}
		p.All = cdn
	}

	for _, typ := range []string{"collector", "processor", "publisher"} {
		if err := unmarshalPluginConfig(typ, p, t); err != nil {
			return err
		}
	}

	return nil
}

func newPluginConfigItem(opts ...pluginConfigOpt) *pluginConfigItem {
	p := &pluginConfigItem{
		ConfigDataNode: cdata.NewNode(),
		Versions:       make(map[int]*cdata.ConfigDataNode),
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
	p.pluginCache[key].Merge(p.All)

	// check for plugin config
	switch pluginType {
	case plugin.CollectorPluginType:
		if res, ok := p.Collector[name]; ok {
			p.pluginCache[key].Merge(res.ConfigDataNode)
			if res2, ok2 := res.Versions[ver]; ok2 {
				p.pluginCache[key].Merge(res2)
			}
		}
	case plugin.ProcessorPluginType:
		if res, ok := p.Processor[name]; ok {
			p.pluginCache[key].Merge(res.ConfigDataNode)
			if res2, ok2 := res.Versions[ver]; ok2 {
				p.pluginCache[key].Merge(res2)
			}
		}
	case plugin.PublisherPluginType:
		if res, ok := p.Publisher[name]; ok {
			p.pluginCache[key].Merge(res.ConfigDataNode)
			if res2, ok2 := res.Versions[ver]; ok2 {
				p.pluginCache[key].Merge(res2)
			}
		}
	}

	return p.pluginCache[key]
}

func unmarshalPluginConfig(typ string, p *pluginConfig, t map[string]interface{}) error {
	if v, ok := t[typ]; ok {
		switch plugins := v.(type) {
		case map[string]interface{}:
			for name, c := range plugins {
				switch typ {
				case "collector":
					p.Collector[name] = newPluginConfigItem()
				case "processor":
					p.Processor[name] = newPluginConfigItem()
				case "publisher":
					p.Publisher[name] = newPluginConfigItem()
				}
				switch col := c.(type) {
				case map[string]interface{}:
					if v, ok := col["all"]; ok {
						jv, err := json.Marshal(v)
						if err != nil {
							return err
						}
						cdn := cdata.NewNode()
						dec := json.NewDecoder(bytes.NewReader(jv))
						dec.UseNumber()
						if err := dec.Decode(&cdn); err != nil {
							return err
						}
						switch typ {
						case "collector":
							p.Collector[name].ConfigDataNode = cdn
						case "processor":
							p.Processor[name].ConfigDataNode = cdn
						case "publisher":
							p.Publisher[name].ConfigDataNode = cdn
						}
					}
					if vs, ok := col["versions"]; ok {
						switch versions := vs.(type) {
						case map[string]interface{}:
							for ver, version := range versions {
								switch v := version.(type) {
								case map[string]interface{}:
									jv, err := json.Marshal(v)
									if err != nil {
										return err
									}
									cdn := cdata.NewNode()
									dec := json.NewDecoder(bytes.NewReader(jv))
									dec.UseNumber()
									if err := dec.Decode(&cdn); err != nil {
										return err
									}
									ver, err := strconv.Atoi(ver)
									if err != nil {
										return err
									}
									switch typ {
									case "collector":
										p.Collector[name].Versions[ver] = cdn
									case "processor":
										p.Processor[name].Versions[ver] = cdn
									case "publisher":
										p.Publisher[name].Versions[ver] = cdn
									}
								default:
									return fmt.Errorf("Error unmarshalling %v'%v' expected '%v' got '%v'", typ, name, map[string]interface{}{}, reflect.TypeOf(v))
								}
							}

						default:
							return fmt.Errorf("Error unmarshalling %v '%v' expected '%v' got '%v'", typ, name, map[string]interface{}{}, reflect.TypeOf(versions))
						}
					}
				default:
					return fmt.Errorf("Error unmarshalling %v '%v' expected '%v' got '%v'", typ, name, map[string]interface{}{}, reflect.TypeOf(col))
				}
			}
		default:
			return fmt.Errorf("Error unmarshalling %v expected '%v' got '%v'", typ, map[string]interface{}{}, reflect.TypeOf(plugins))
		}
	}
	return nil
}
