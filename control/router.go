// Router is the entry point for execution commands and routing to plugins
package control

import "github.com/intelsdilabs/pulse/control/routing"

type RouterResponse interface {
}

type RoutingStrategy interface {
	Select(routing.SelectablePluginPool, []routing.SelectablePlugin) (routing.SelectablePlugin, error)
}

type pluginRouter struct {
	strategy RoutingStrategy
}

func (p *pluginRouter) Strategy() RoutingStrategy {
	return p.strategy
}

func newPluginRouter() *pluginRouter {
	return &pluginRouter{
		strategy: &routing.RoundRobinStrategy{},
	}
}
