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

// Calls collector plugins for the metric types and returns collection response containing metrics. Blocking method.
// this methods:
// uses metricCatalog to get loadedPlugin by metrictype.namespace
// uses pluginRunner to get pluginPoll by
func (p *pluginRouter) Collect(metricTypes []core.MetricType, config *cdata.ConfigDataNode, deadline time.Time) (response *collectionResponse, err error) {

	// selection is a mapping from key -> metricTypes
	selections, err := groupMetricTypesByLoadedPlugin(metricTypes, p.metricCatalog)
	if err != nil {
		return nil, err
	}

	// For each available plugin call available plugin using RPC client and wait for response (goroutines)
	var selectedAP *availablePlugin
	resp := newCollectionResponse()
	for key, _ := range *selections {

		// ??????? wyciagnij tym razem z runner pool of plugins to do what ???
		pool := p.pluginRunner.AvailablePlugins().Collectors.GetPluginPool(key)

		// is pool can be used
		if pool == nil {
			// return error because this plugin has no pool
			return nil, errors.New(fmt.Sprintf("no available plugins for plugin type (%s)", key))
		}

		// Lock this apPool so we are the only one operating on it.
		if pool.Count() == 0 {
			// return error indicating we have no available plugins to call for Collect
			return nil, errors.New(fmt.Sprintf("no available plugins for plugin type (%s)", key))
		}

		// Use a router strategy to select an available plugin from the pool
		selectedAP, err = pool.SelectUsingStrategy(p.Strategy)

		if err != nil {
			return nil, err
		}

		if selectedAP == nil {
			return nil, errors.New(fmt.Sprintf("no available plugin selected (%s)", key))
		}

		// Attempt collection on selected available plugin
		selectedAP.hitCount++
		selectedAP.lastHitTime = time.Now()
		metrics, err := selectedAP.Client.(client.PluginCollectorClient).CollectMetrics(metricTypes)
		if err != nil {
			resp.Errors = append(resp.Errors, err)
			return resp, nil
		}
		// extern metrics
		resp.Metrics = metrics
	}

	return resp, nil
}

func (p *pluginRouter) SetRunner(r runsPlugins) {
	p.pluginRunner = r
}

func (p *pluginRouter) SetMetricCatalog(m catalogsMetrics) {
	p.metricCatalog = m
}

// pluginCallSelection - used as elementy in mapping loadedPlugin -> MetricTypes
type pluginCallSelection struct {
	Plugin      *loadedPlugin
	MetricTypes []core.MetricType
}

func (p *pluginCallSelection) Count() int {
	return len(p.MetricTypes)
}

// groupMetricTypesByLoadedPlugin - take a bunch of metricTypes and return
// a mapping from loadedPlugin.Key() -> collection of metricTypes (called selection)
// return errors where therie no such metricType available
func groupMetricTypesByLoadedPlugin(
	metricTypes []core.MetricType,
	metricCatalog catalogsMetrics,
) (*map[string]*pluginCallSelection, error) {

	// group metricsTypes by loadedPlugin.key (key = name + version)
	selections := make(map[string]*pluginCallSelection)

	// For each plugin type select a matching available plugin to call
	for _, m := range metricTypes {

		// This is set to choose the newest and not pin version. TODO, be sure version is set to -1 if not provided by user on Task creation.
		lp, err := metricCatalog.GetPlugin(m.Namespace(), -1)

		// Single error fails entire operation - propagate the error received from catalog
		if err != nil {
			return nil, err
		}

		// Single error fails entire operation - there is plugin to handle these metricTypes
		if lp == nil {
			return nil, errors.New(fmt.Sprintf("Metric missing: %s", strings.Join(m.Namespace(), "/")))
		}

		fmt.Printf("Found plugin (%s v%d) for metric (%s)\n", lp.Name(), lp.Version(), strings.Join(m.Namespace(), "/"))

		// group them
		sel, ok := selections[lp.Key()]
		if ok {
			// selection already exists, just update the collections of metricTypes inside
			sel.MetricTypes = append(sel.MetricTypes, m)
		} else {
			// create new selection structure
			selections[lp.Key()] = &pluginCallSelection{
				Plugin:      lp,
				MetricTypes: []core.MetricType{m},
			}

		}
	}
	return &selections, nil
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
