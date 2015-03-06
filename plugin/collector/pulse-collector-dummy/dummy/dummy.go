package dummy

import (
	"github.com/intelsdilabs/pulse/control/plugin"
)

const (
	Name    = "dummy"
	Version = 1
	Type    = plugin.CollectorPluginType
)

// Dummy collector implementation used for testing
type Dummy struct {
}

func (f *Dummy) CollectMetrics([]plugin.PluginMetricType) ([]plugin.PluginMetric, error) {
	m := plugin.PluginMetric{Namespace_: []string{"intel", "dummy", "foo"}, Data_: 1}
	ms := []plugin.PluginMetric{m}
	return ms, nil
}

func (f *Dummy) GetMetricTypes() ([]plugin.PluginMetricType, error) {
	m1 := plugin.NewPluginMetricType([]string{"intel", "dummy", "foo"})
	m2 := plugin.NewPluginMetricType([]string{"intel", "dummy", "bar"})
	return []plugin.PluginMetricType{*m1, *m2}, nil
}

func Meta() *plugin.PluginMeta {
	return plugin.NewPluginMeta(Name, Version, Type)
}

func ConfigPolicy() *plugin.ConfigPolicy {
	c := new(plugin.ConfigPolicy)
	return c
}
