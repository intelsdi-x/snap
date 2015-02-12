package facter

import (
	"github.com/intelsdilabs/pulse/control/plugin"
)

const (
	Name    = "facter"
	Version = 1
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
	m := new(plugin.PluginMeta)
	m.Name = Name
	m.Version = Version
	return m
}

func ConfigPolicy() *plugin.ConfigPolicy {
	c := new(plugin.ConfigPolicy)
	return c
}
