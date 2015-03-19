// Router is the entry point for execution commands and routing to plugins
package control

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/intelsdilabs/pulse/control/plugin/client"
	"github.com/intelsdilabs/pulse/control/routing"
	"github.com/intelsdilabs/pulse/core"
	"github.com/intelsdilabs/pulse/core/cdata"
)

type RouterResponse interface {
}

type RoutingStrategy interface {
	Select(routing.SelectablePluginPool, []routing.SelectablePlugin) (routing.SelectablePlugin, error)
}

type pluginRouter struct {
	Strategy RoutingStrategy

	metricCatalog catalogsMetrics
	pluginRunner  runsPlugins
}

func newPluginRouter() *pluginRouter {
	return &pluginRouter{
		Strategy: &routing.RoundRobinStrategy{},
	}
}

func (p *pluginRouter) SetRunner(r runsPlugins) {
	p.pluginRunner = r
}

func (p *pluginRouter) SetMetricCatalog(m catalogsMetrics) {
	p.metricCatalog = m
}
