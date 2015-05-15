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

func getLoadAvgMetricTypes(mts []plugin.PluginMetricType) []plugin.PluginMetricType {
	mts = append(mts, plugin.PluginMetricType{Namespace_: []string{"psutil", "load", "load1"}})
	mts = append(mts, plugin.PluginMetricType{Namespace_: []string{"psutil", "load", "load5"}})
	mts = append(mts, plugin.PluginMetricType{Namespace_: []string{"psutil", "load", "load15"}})
	return mts
}
