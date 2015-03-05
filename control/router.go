// Router is the entry point for execution commands and routing to plugins
package control

import (
	"log"
	"time"

	"github.com/intelsdilabs/pulse/core"
	"github.com/intelsdilabs/pulse/core/cdata"
)

type Router interface {
	Collect([]core.MetricType, *cdata.ConfigDataNode, time.Time) *collectionResponse
}

type RouterResponse interface {
}

type pluginRouter struct {
	// Pointer to control metric catalog
	metricCatalog *metricCatalog
}

func newRouter(mc *metricCatalog) Router {
	return &pluginRouter{metricCatalog: mc}
}

// Calls collector plugins for the metric types and returns collection response containing metrics. Blocking method.
func (p *pluginRouter) Collect(metricTypes []core.MetricType, config *cdata.ConfigDataNode, deadline time.Time) (response *collectionResponse) {
	// For each MT sort into plugin types we need to call
	log.Println(metricTypes)

	// For each plugin type select a matching available plugin to call
	for _, m := range metricTypes {
		log.Println(m.Namespace())
		lp, err := p.metricCatalog.resolvePlugin(m.Namespace(), m.Version())

		// TODO handle error here. Single error fails entire operation.
		log.Println(lp, err)
	}

	// For each available plugin call available plugin using RPC client and wait for response (goroutines)

	// Wait for all responses(happy) or timeout(unhappy)

	// (happy)reduce responses into single collection response and return

	// (unhappy)return response with timeout state

	return &collectionResponse{}
}

type collectionResponse struct {
	Errors []error
}
