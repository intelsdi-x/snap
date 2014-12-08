package dummy

import (
	"github.com/intelsdilabs/pulse/control/plugin"
)

const (
	Name    = "dummy"
	Version = 1
	Type    = "collector"
)

// Dummy collector implementation used for testing
type Dummy struct {
}

func (f *Dummy) Collect(args plugin.CollectorArgs, reply *plugin.CollectorReply) error {
	return nil
}

func Meta() *plugin.PluginMeta {
	m := new(plugin.PluginMeta)
	m.Name = Name
	m.Version = Version
	return m
}

func ConfigPolicy() *plugin.ConfigPolicy {
	c := new(plugin.ConfigPolicy)
	return c
}
