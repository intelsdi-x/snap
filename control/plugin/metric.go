package plugin

import (
	"time"

	"github.com/intelsdilabs/pulse/core/cdata"
)

// Represents a collected metric. Only used within plugins and across plugin calls.
// Converted to core.Metric before being used within modules.
type PluginMetric struct {
	Namespace_ []string
	Data_      interface{}
}

func (p PluginMetric) Namespace() []string {
	return p.Namespace_
}

func (p PluginMetric) Data() interface{} {
	return p.Data_
}

// Represents a metric type. Only used within plugins and across plugin calls.
// Converted to core.MetricType before being used within modules.
type PluginMetricType struct {
	Namespace_          []string
	LastAdvertisedTime_ time.Time
	Version_            int
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

// This version of MetricType never implements cdata.ConfigDataNode
func (p PluginMetricType) Config() *cdata.ConfigDataNode {
	return nil
}

func NewPluginMetricType(ns []string) *PluginMetricType {
	return &PluginMetricType{
		Namespace_:          ns,
		LastAdvertisedTime_: time.Now(),
	}
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
