package client

import (
	"github.com/intelsdilabs/pulse/control/plugin"
)

// A client providing common plugin method calls.
type PluginClient interface {
	Ping() error
	Kill(string) error
}

// A client providing collector specific plugin method calls.
type PluginCollectorClient interface {
	PluginClient
	CollectMetrics([]plugin.MetricType) ([]plugin.Metric, error)
	GetMetricTypes() ([]plugin.MetricType, error)
}

// A client providing processor specific plugin method calls.
type PluginProcessorClient interface {
	PluginClient
	ProcessMetricData()
}
