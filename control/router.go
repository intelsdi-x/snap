// Router is the entry point for execution commands and routing to plugins
package control

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/intelsdilabs/pulse/core"
	"github.com/intelsdilabs/pulse/core/cdata"
)

type RouterResponse interface {
}

type pluginRouter struct {
	metricCatalog catalogsMetrics
	pluginRunner  runsPlugins
}

func newPluginRouter() *pluginRouter {
	return &pluginRouter{}
}

type pluginCallSelection struct {
	Plugin      *loadedPlugin
	MetricTypes []core.MetricType
}

// Calls collector plugins for the metric types and returns collection response containing metrics. Blocking method.
func (p *pluginRouter) Collect(metricTypes []core.MetricType, config *cdata.ConfigDataNode, deadline time.Time) (response *collectionResponse, err error) {
	// For each MT sort into plugin types we need to call
	fmt.Println(metricTypes)

	fmt.Println("\nMetric Catalog\n*****")
	fmt.Println(p.metricCatalog)
	for k, m := range p.metricCatalog.Table() {
		fmt.Println(k, m)
	}
	fmt.Println("\n")

	pluginCallSelectionMap := make(map[string]pluginCallSelection)
	// For each plugin type select a matching available plugin to call
	for _, m := range metricTypes {

		// This is set to choose the newest and not pin version. TODO, be sure version is set to -1 if not provided by user on Task creation.
		lp, err := p.metricCatalog.resolvePlugin(m.Namespace(), -1)

		// fmt.Println("\nMetric Catalog\n*****")
		// for k, m := range p.metricCatalog.Table() {
		// 	fmt.Println(k, m)
		// }

		// TODO handle error here. Single error fails entire operation.
		if err != nil {
			// can't find a matching plugin, fail - TODO
		}

		if lp == nil {
			return nil, errors.New(fmt.Sprintf("Metric missing: %s", strings.Join(m.Namespace(), "/")))
		}

		fmt.Printf("Found plugin (%s v%d) for metric (%s)\n", lp.Name(), lp.Version(), strings.Join(m.Namespace(), "/"))
		x, _ := pluginCallSelectionMap[lp.Key()]
		x.Plugin = lp
		x.MetricTypes = append(x.MetricTypes, m)
		pluginCallSelectionMap[lp.Key()] = x

	}
	// For each available plugin call available plugin using RPC client and wait for response (goroutines)
	fmt.Println(pluginCallSelectionMap)
	fmt.Println(p.pluginRunner.AvailablePlugins().Collectors.Table())

	// Wait for all responses(happy) or timeout(unhappy)

	// (happy)reduce responses into single collection response and return

	// (unhappy)return response with timeout state

	return &collectionResponse{}, nil
}

type collectionResponse struct {
	Errors []error
}

func (p *pluginRouter) SetRunner(r runsPlugins) {
	p.pluginRunner = r
}

func (p *pluginRouter) SetMetricCatalog(m catalogsMetrics) {
	p.metricCatalog = m
}
