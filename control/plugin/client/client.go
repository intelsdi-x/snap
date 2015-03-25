package client

import (
	// "github.com/intelsdilabs/pulse/control/plugin"
	"github.com/intelsdilabs/pulse/core"
)

// A client providing common plugin method calls.
type PluginClient interface {
	Ping() error
	Kill(string) error
}

// A client providing collector specific plugin method calls.
type PluginCollectorClient interface {
	PluginClient
	CollectMetrics([]core.MetricType) ([]core.Metric, error)
	GetMetricTypes() ([]core.MetricType, error)
}

// A client providing processor specific plugin method calls.
type PluginProcessorClient interface {
	PluginClient
	ProcessMetricData()
}

type PluginPublisherClient interface {
	PluginClient
	PublishMetrics([]core.MetricType) error
}
