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

type pluginCallSelection struct {
	Plugin      *loadedPlugin
	MetricTypes []core.MetricType
}

func (p *pluginCallSelection) Count() int {
	return len(p.MetricTypes)
}

// Calls collector plugins for the metric types and returns collection response containing metrics. Blocking method.
func (p *pluginRouter) Collect(metricTypes []core.MetricType, config *cdata.ConfigDataNode, deadline time.Time) (response *collectionResponse, err error) {
	pluginCallSelectionMap := make(map[string]pluginCallSelection)
	// For each plugin type select a matching available plugin to call
	for _, m := range metricTypes {

		// This is set to choose the newest and not pin version. TODO, be sure version is set to -1 if not provided by user on Task creation.
		lp, err := p.metricCatalog.GetPlugin(m.Namespace(), -1)

		// Single error fails entire operation.
		if err != nil {
			return nil, err
		}

		// Single error fails entire operation.
		if lp == nil {
			return nil, errors.New(fmt.Sprintf("Metric missing: %s", strings.Join(m.Namespace(), "/")))
		}

		// fmt.Printf("Found plugin (%s v%d) for metric (%s)\n", lp.Name(), lp.Version(), strings.Join(m.Namespace(), "/"))
		x, _ := pluginCallSelectionMap[lp.Key()]
		x.Plugin = lp
		x.MetricTypes = append(x.MetricTypes, m)
		pluginCallSelectionMap[lp.Key()] = x

	}

	// For each available plugin call available plugin using RPC client and wait for response (goroutines)
	var selectedAP *availablePlugin
	for pluginKey, _ := range pluginCallSelectionMap {
		// fmt.Printf("plugin: (%s) has (%d) metrics to gather\n", pluginKey, metrics.Count())

		apPluginPool := p.pluginRunner.AvailablePlugins().Collectors.GetPluginPool(pluginKey)

		if apPluginPool == nil {
			// return error because this plugin has no pool
			return nil, errors.New(fmt.Sprintf("no available plugins for plugin type (%s)", pluginKey))
		}

		// Lock this apPool so we are the only one operating on it.
		if apPluginPool.Count() == 0 {
			// return error indicating we have no available plugins to call for Collect
			return nil, errors.New(fmt.Sprintf("no available plugins for plugin type (%s)", pluginKey))
		}

		// Use a router strategy to select an available plugin from the pool
		// fmt.Printf("%d available plugin in pool for (%s)\n", apPluginPool.Count(), pluginKey)
		ap, err := apPluginPool.SelectUsingStrategy(p.Strategy)

		if err != nil {
			return nil, err
		}

		if ap == nil {
			return nil, errors.New(fmt.Sprintf("no available plugin selected (%s)", pluginKey))
		}
		selectedAP = ap
	}

	resp := newCollectionResponse()
	// Attempt collection on selected available plugin
	selectedAP.hitCount++
	selectedAP.lastHitTime = time.Now()
	metrics, err := selectedAP.Client.(client.PluginCollectorClient).CollectMetrics(metricTypes)
	if err != nil {
		resp.Errors = append(resp.Errors, err)
		return resp, nil
	}
	resp.Metrics = metrics
	return resp, nil
}

func (p *pluginRouter) SetRunner(r runsPlugins) {
	p.pluginRunner = r
}

func (p *pluginRouter) SetMetricCatalog(m catalogsMetrics) {
	p.metricCatalog = m
}

type collectionResponse struct {
	Metrics []core.Metric
	Errors  []error
}

func newCollectionResponse() *collectionResponse {
	return &collectionResponse{
		Metrics: make([]core.Metric, 0),
		Errors:  make([]error, 0),
	}
}
