package dummy

import (
	"log"

	"github.com/intelsdilabs/pulse/control/plugin"
	"github.com/intelsdilabs/pulse/control/plugin/cpolicy"
)

const (
	Name    = "dummy1"
	Version = 1
	Type    = plugin.CollectorPluginType
)

// Dummy collector implementation used for testing
type Dummy struct {
}

func (f *Dummy) CollectMetrics(mts []plugin.PluginMetricType) ([]plugin.PluginMetric, error) {
	for _, p := range mts {
		log.Println("collecting", p)
	}
	m := plugin.PluginMetric{Namespace_: []string{"intel", "dummy", "foo"}, Data_: 1}
	ms := []plugin.PluginMetric{m}
	return ms, nil
}

func (f *Dummy) GetMetricTypes() ([]plugin.PluginMetricType, error) {
	m1 := plugin.NewPluginMetricType([]string{"intel", "dummy", "foo"})
	m2 := plugin.NewPluginMetricType([]string{"intel", "dummy", "bar"})
	return []plugin.PluginMetricType{*m1, *m2}, nil
}

func (f *Dummy) GetConfigPolicyTree() (cpolicy.ConfigPolicyTree, error) {
	c := cpolicy.NewTree()
	rule, _ := cpolicy.NewStringRule("name", false, "bob")
	rule2, _ := cpolicy.NewStringRule("password", true)
	p := cpolicy.NewPolicyNode()
	p.Add(rule)
	p.Add(rule2)
	c.Add([]string{"intel", "dummy", "foo"}, p)
	return *c, nil
}

func Meta() *plugin.PluginMeta {
	return plugin.NewPluginMeta(Name, Version, Type)
}
