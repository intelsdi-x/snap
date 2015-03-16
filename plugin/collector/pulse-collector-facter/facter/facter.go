package facter

import (
	"github.com/intelsdilabs/pulse/control/plugin"
)

const (
	Name    = "facter"
	Version = 1
	Type    = plugin.CollectorPluginType
)

type Facter struct {
}

func (f *Facter) GetMetricTypes(_ plugin.GetMetricTypesArgs, reply *plugin.GetMetricTypesReply) error {
	//reply *[]*plugin.MetricType
	return nil
}

func (f *Facter) Collect(args plugin.CollectorArgs, reply *plugin.CollectorReply) error {
	return nil
}

func Meta() *plugin.PluginMeta {
	return plugin.NewPluginMeta(Name, Version, Type)
}

func ConfigPolicy() *plugin.ConfigPolicy {
	c := new(plugin.ConfigPolicy)
	return c
}
