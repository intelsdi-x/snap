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

func (f *Facter) CollectMetrics([]plugin.MetricType) ([]plugin.Metric, error) {
	m := plugin.Metric{Namespace: []string{"intel", "facter", "foo"}, Data: 1}
	ms := []plugin.Metric{m}
	return ms, nil
}

func (f *Facter) GetMetricTypes() ([]plugin.MetricType, error) {
	m := plugin.NewMetricType([]string{"intel", "facter", "foo"})
	return []plugin.MetricType{*m}, nil
}

func Meta() *plugin.PluginMeta {
	return plugin.NewPluginMeta(Name, Version, Type)
}

func ConfigPolicy() *plugin.ConfigPolicy {
	c := new(plugin.ConfigPolicy)
	return c
}
