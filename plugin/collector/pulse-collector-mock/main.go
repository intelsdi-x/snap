package main

import (
	"errors"
	"log"
	"os"
	"time"

	"github.com/intelsdilabs/pulse/control/plugin"
	"github.com/intelsdilabs/pulse/control/plugin/cpolicy"
)

type mock struct{}

// CollectMetrics returns whatever was asked about
func (_ *mock) CollectMetrics(mts []plugin.PluginMetricType) ([]plugin.PluginMetric, error) {
	log.Println("mockPlugin: asked for metrics")
	for _, mt := range mts {
		if mt.Namespace()[2] == "good" {
			log.Println("mockPlugin: returns good metric")
			return []plugin.PluginMetric{plugin.PluginMetric{Namespace_: mt.Namespace(), Data_: 1}}, nil
		}
		if mt.Namespace()[2] == "error" {
			log.Println("mockPlugin: returns error")
			return nil, errors.New("special error")
		}
		if mt.Namespace()[2] == "panic" {
			log.Println("mockPlugin: panics")
			panic("special panic")
		}
		if mt.Namespace()[2] == "timeout" {
			log.Println("mockPlugin: timeouts")
			time.Sleep(3600 * time.Second)
		}
	}
	return nil, errors.New("unknow metric type")
}

// GetMetricTypes always returns two metrics
func (_ *mock) GetMetricTypes() ([]plugin.PluginMetricType, error) {
	log.Println("mockPlugin: asked for metrics types")
	m1 := plugin.NewPluginMetricType([]string{"intel", "mock", "good"})
	m2 := plugin.NewPluginMetricType([]string{"intel", "mock", "error"})
	m3 := plugin.NewPluginMetricType([]string{"intel", "mock", "timeout"})
	m4 := plugin.NewPluginMetricType([]string{"intel", "mock", "panic"})
	return []plugin.PluginMetricType{*m1, *m2, *m3, *m4}, nil
}

func main() {
	plugin.Start(
		plugin.NewPluginMeta("intel mock", 1, plugin.CollectorPluginType),
		new(mock),
		cpolicy.NewTree(),
		os.Args[1],
	)
}
