package plugin

import (
	"time"

	"github.com/intelsdilabs/pulse/core/cdata"
)

// Represents a metric type. Only used within plugins and across plugin calls.
// Converted to core.MetricType before being used within modules.
type PluginMetricType struct {
	Namespace_          []string
	LastAdvertisedTime_ time.Time
	Version_            int
	Config_             *cdata.ConfigDataNode //map[string]ctypes.ConfigValue
	Data_               interface{}
}

// // PluginMetricType Constructor
func NewPluginMetricType(namespace []string, data interface{}) *PluginMetricType {
	return &PluginMetricType{
		Namespace_: namespace,
		Data_:      data,
	}
}

// Returns the namespace.
func (p PluginMetricType) Namespace() []string {
	return p.Namespace_
}

// Returns the last time this metric type was received from the plugin.
func (p PluginMetricType) LastAdvertisedTime() time.Time {
	return p.LastAdvertisedTime_
}

// Returns the namespace.
func (p PluginMetricType) Version() int {
	return p.Version_
}

// Config returns the map of config data for this metric
func (p PluginMetricType) Config() *cdata.ConfigDataNode {
	return p.Config_
}

func (p PluginMetricType) Data() interface{} {
	return p.Data_
}

func (p *PluginMetricType) AddData(data interface{}) {
	p.Data_ = data
}

/*

core.Metric(i) (used by pulse modules)
core.MetricType(i) (used by pulse modules)

PluginMetric (created by plugins and returned, goes over RPC)
PLuginMetricType (created by plugins and returned, goes over RPC)

Get

agent - call Get
client - call Get
plugin - builds array of PluginMetricTypes
plugin - return array of PluginMetricTypes
client - returns array of PluginMetricTypes through MetricType interface
agent - receives MetricTypes

Collect

agent - call Collect pass MetricTypes
client - call Collect, convert MetricTypes into new (plugin) PluginMetricTypes struct
plugin - build array of PluginMetric based on (plugin) MetricTypes
plugin - return array of PluginMetrics
client - return array of PluginMetrics through core.Metrics interface
agent - receive array of core.Metrics


*/
