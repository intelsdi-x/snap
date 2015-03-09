package facter

import (
	"github.com/intelsdilabs/pulse/control/plugin"
	"github.com/intelsdilabs/pulse/control/plugin/cpolicy"
)

const (
	Name    = "facter"
	Version = 1
	Type    = plugin.CollectorPluginType
)

type Facter struct {
}

func (f *Facter) CollectMetrics([]plugin.PluginMetricType) ([]plugin.PluginMetric, error) {
	m := plugin.PluginMetric{Namespace_: []string{"intel", "facter", "foo"}, Data_: 1}
	ms := []plugin.PluginMetric{m}
	return ms, nil
}

func (f *Facter) GetMetricTypes() ([]plugin.PluginMetricType, error) {
	m := plugin.NewPluginMetricType([]string{"intel", "facter", "foo"})
	return []plugin.PluginMetricType{*m}, nil
}

func Meta() *plugin.PluginMeta {
	return plugin.NewPluginMeta(Name, Version, Type)
}

func ConfigPolicyTree() *cpolicy.ConfigPolicyTree {
	c := cpolicy.NewTree()
	return c
}
