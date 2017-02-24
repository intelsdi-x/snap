package api

import (
	"github.com/intelsdi-x/snap/core"
	"github.com/intelsdi-x/snap/core/serror"
)

type Metrics interface {
	MetricCatalog() ([]core.CatalogedMetric, error)
	FetchMetrics(core.Namespace, int) ([]core.CatalogedMetric, error)
	GetMetricVersions(core.Namespace) ([]core.CatalogedMetric, error)
	GetMetric(core.Namespace, int) (core.CatalogedMetric, error)
	Load(*core.RequestedPlugin) (core.CatalogedPlugin, serror.SnapError)
	Unload(core.Plugin) (core.CatalogedPlugin, serror.SnapError)
	PluginCatalog() core.PluginCatalog
	AvailablePlugins() []core.AvailablePlugin
	GetAutodiscoverPaths() []string
	GetTempDir() string
}
