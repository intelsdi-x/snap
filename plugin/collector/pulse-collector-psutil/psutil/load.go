package psutil

import (
	"fmt"

	"github.com/intelsdi-x/pulse/control/plugin"
	"github.com/shirou/gopsutil/load"
)

func loadAvg(ns []string) (*plugin.PluginMetricType, error) {
	load, err := load.LoadAvg()
	if err != nil {
		return nil, err
	}

	switch joinNamespace(ns) {
	case "/psutil/load/load1":
		return &plugin.PluginMetricType{
			Namespace_: ns,
			Data_:      load.Load1,
		}, nil
	case "/psutil/load/load5":
		return &plugin.PluginMetricType{
			Namespace_: ns,
			Data_:      load.Load5,
		}, nil
	case "/psutil/load/load15":
		return &plugin.PluginMetricType{
			Namespace_: ns,
			Data_:      load.Load15,
		}, nil
	}

	return nil, fmt.Errorf("Unknown error processing %v", ns)
}

func getLoadAvgMetricTypes() []plugin.PluginMetricType {
	t := []string{"load1", "load5", "load15"}
	mts := make([]plugin.PluginMetricType, len(t))
	for i, te := range t {
		mts[i] = plugin.PluginMetricType{Namespace_: []string{"psutil", "load", te}}
	}
	return mts
}
