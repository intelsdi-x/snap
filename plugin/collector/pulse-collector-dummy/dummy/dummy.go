package dummy

import (
	"time"

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

func (f *Dummy) Collect(args plugin.CollectorArgs, reply *plugin.CollectorReply) error {
	return nil
}

func (f *Dummy) GetMetricTypes(_ plugin.GetMetricTypesArgs, reply *plugin.GetMetricTypesReply) error {
	reply.MetricTypes = []*plugin.MetricType{
		plugin.NewMetricType([]string{"foo", "bar"}, time.Now()),
	}
	return nil
}

func Meta() *plugin.PluginMeta {
	return plugin.NewPluginMeta(Name, Version, Type)
}

func ConfigPolicy() *plugin.ConfigPolicy {
	c := new(plugin.ConfigPolicy)
	return c
}
