package client

import (
	// "github.com/intelsdilabs/pulse/control/plugin"
	"github.com/intelsdilabs/pulse/control/plugin/cpolicy"
	"github.com/intelsdilabs/pulse/core"
	"github.com/intelsdilabs/pulse/core/ctypes"
)

// PluginClient A client providing common plugin method calls.
type PluginClient interface {
	Ping() error
	Kill(string) error
}

// PluginCollectorClient A client providing collector specific plugin method calls.
type PluginCollectorClient interface {
	PluginClient
	CollectMetrics([]core.Metric) ([]core.Metric, error)
	GetMetricTypes() ([]core.Metric, error)
	GetConfigPolicyTree() (cpolicy.ConfigPolicyTree, error)
}

// PluginProcessorClient A client providing processor specific plugin method calls.
type PluginProcessorClient interface {
	PluginClient
	ProcessMetricData([]core.Metric) ([]core.Metric, error)
}

// PluginPublisherClient A client providing publishing specific plugin method calls.
type PluginPublisherClient interface {
	PluginClient
	Publish(contentType string, content []byte, config map[string]ctypes.ConfigValue) error
	GetConfigPolicyNode() (cpolicy.ConfigPolicyNode, error)
}
