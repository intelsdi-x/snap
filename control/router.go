// Router is the entry point for execution commands and routing to plugins
package control

import "github.com/intelsdi-x/pulse/control/routing"

type RouterResponse interface {
}

type RoutingStrategy interface {
	Select(routing.SelectablePluginPool, []routing.SelectablePlugin) (routing.SelectablePlugin, error)
}
