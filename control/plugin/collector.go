package plugin

import (
	"encoding/gob"

	"github.com/intelsdi-x/pulse/control/plugin/cpolicy"
	"github.com/intelsdi-x/pulse/core/ctypes"
)

// Acts as a proxy for RPC calls to a CollectorPlugin. This helps keep the function signature simple
// within plugins vs. having to match required RPC patterns.

// Collector plugin
type CollectorPlugin interface {
	Plugin
	CollectMetrics([]PluginMetricType) ([]PluginMetricType, error)
	GetMetricTypes() ([]PluginMetricType, error)
	GetConfigPolicy() (cpolicy.ConfigPolicy, error)
}

func init() {
	gob.Register(*(&ctypes.ConfigValueInt{}))
	gob.Register(*(&ctypes.ConfigValueStr{}))
	gob.Register(*(&ctypes.ConfigValueFloat{}))

	gob.Register(cpolicy.NewPolicyNode())
	gob.Register(&cpolicy.StringRule{})
	gob.Register(&cpolicy.IntRule{})
	gob.Register(&cpolicy.FloatRule{})
}
