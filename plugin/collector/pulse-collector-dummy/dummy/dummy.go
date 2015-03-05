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

func (f *Dummy) CollectMetrics([]plugin.MetricType) ([]plugin.Metric, error) {
	m := plugin.Metric{Namespace: []string{"intel", "dummy", "foo"}, Data: 1}
	ms := []plugin.Metric{m}
	return ms, nil
}

func (f *Dummy) GetMetricTypes() ([]plugin.MetricType, error) {
	m := plugin.NewMetricType([]string{"intel", "dummy", "foo"})
	return []plugin.MetricType{*m}, nil
}

func Meta() *plugin.PluginMeta {
	return plugin.NewPluginMeta(Name, Version, Type)
}

func ConfigPolicy() *plugin.ConfigPolicy {
	c := new(plugin.ConfigPolicy)
	return c
}
