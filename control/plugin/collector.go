package plugin

// Acts as a proxy for RPC calls to a CollectorPlugin. This helps keep the function signature simple
// within plugins vs. having to match required RPC patterns.

// Collector plugin
type CollectorPlugin interface {
	Plugin
	CollectMetrics([]PluginMetricType) ([]PluginMetricType, error)
	GetMetricTypes() ([]PluginMetricType, error)
}
