package psutil

import (
	"fmt"

	"github.com/intelsdi-x/pulse/control/plugin"
	"github.com/shirou/gopsutil/mem"
)

func virtualMemory(ns []string) (*plugin.PluginMetricType, error) {
	mem, err := mem.VirtualMemory()
	if err != nil {
		return nil, err
	}

	switch joinNamespace(ns) {
	case "/psutil/vm/total":
		return &plugin.PluginMetricType{
			Namespace_: ns,
			Data_:      mem.Total,
		}, nil
	case "/psutil/vm/available":
		return &plugin.PluginMetricType{
			Namespace_: ns,
			Data_:      mem.Available,
		}, nil
	case "/psutil/vm/used":
		return &plugin.PluginMetricType{
			Namespace_: ns,
			Data_:      mem.Used,
		}, nil
	case "/psutil/vm/used_percent":
		return &plugin.PluginMetricType{
			Namespace_: ns,
			Data_:      mem.UsedPercent,
		}, nil
	case "/psutil/vm/free":
		return &plugin.PluginMetricType{
			Namespace_: ns,
			Data_:      mem.Free,
		}, nil
	case "/psutil/vm/active":
		return &plugin.PluginMetricType{
			Namespace_: ns,
			Data_:      mem.Active,
		}, nil
	case "/psutil/vm/inactive":
		return &plugin.PluginMetricType{
			Namespace_: ns,
			Data_:      mem.Inactive,
		}, nil
	case "/psutil/vm/buffers":
		return &plugin.PluginMetricType{
			Namespace_: ns,
			Data_:      mem.Buffers,
		}, nil
	case "/psutil/vm/cached":
		return &plugin.PluginMetricType{
			Namespace_: ns,
			Data_:      mem.Cached,
		}, nil
	case "/psutil/vm/wired":
		return &plugin.PluginMetricType{
			Namespace_: ns,
			Data_:      mem.Wired,
		}, nil
	case "/psutil/vm/shared":
		return &plugin.PluginMetricType{
			Namespace_: ns,
			Data_:      mem.Shared,
		}, nil
	}

	return nil, fmt.Errorf("Unknown error processing %v", ns)
}

func getVirtualMemoryMetricTypes() []plugin.PluginMetricType {
	t := []string{"total", "available", "used", "used_percent", "free", "active", "inactive", "buffers", "cached", "wired", "shared"}
	mts := make([]plugin.PluginMetricType, len(t))
	for i, te := range t {
		mts[i] = plugin.PluginMetricType{Namespace_: []string{"psutil", "vm", te}}
	}
	return mts
}
