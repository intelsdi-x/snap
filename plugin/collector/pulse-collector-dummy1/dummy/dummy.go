package dummy

import (
	"fmt"

	"github.com/intelsdi-x/pulse/control/plugin"
	"github.com/intelsdi-x/pulse/control/plugin/cpolicy"
)

const (
	// Name of plugin
	Name = "dummy1"
	// Version of plugin
	Version = 1
	// Type of plugin
	Type = plugin.CollectorPluginType
)

// Dummy collector implementation used for testing
type Dummy struct {
}

// CollectMetrics collects metrics for testing
func (f *Dummy) CollectMetrics(mts []plugin.PluginMetricType) ([]plugin.PluginMetricType, error) {
	metrics := make([]plugin.PluginMetricType, len(mts))
	for i, p := range mts {
		data := fmt.Sprintf("The dummy collected data! config data: user=%s password=%s", p.Config().Table()["name"], p.Config().Table()["name"])
		metrics[i] = plugin.PluginMetricType{
			Namespace_: p.Namespace(),
			Data_:      data,
		}
	}
	return metrics, nil
}

//GetMetricTypes returns metric types for testing
func (f *Dummy) GetMetricTypes() ([]plugin.PluginMetricType, error) {
	m1 := &plugin.PluginMetricType{Namespace_: []string{"intel", "dummy", "foo"}}
	m2 := &plugin.PluginMetricType{Namespace_: []string{"intel", "dummy", "bar"}}
	return []plugin.PluginMetricType{*m1, *m2}, nil
}

//GetConfigPolicyTree returns a ConfigPolicyTree for testing
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

//Meta returns meta data for testing
func Meta() *plugin.PluginMeta {
	return plugin.NewPluginMeta(Name, Version, Type)
}
