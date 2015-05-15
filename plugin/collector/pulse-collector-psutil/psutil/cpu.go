package psutil

import (
	"fmt"
	"regexp"
	"runtime"

	"github.com/intelsdi-x/pulse/control/plugin"
	"github.com/shirou/gopsutil/cpu"
)

func cpuTimes(ns []string) (*plugin.PluginMetricType, error) {
	cpus, err := cpu.CPUTimes(true)
	if err != nil {
		return nil, err
	}

	for _, cpu := range cpus {
		switch {
		case regexp.MustCompile(`^/psutil/cpu.*/user`).MatchString(joinNamespace(ns)):
			return &plugin.PluginMetricType{
				Namespace_: ns,
				Data_:      cpu.User,
			}, nil
		case regexp.MustCompile(`^/psutil/cpu.*/system`).MatchString(joinNamespace(ns)):
			return &plugin.PluginMetricType{
				Namespace_: ns,
				Data_:      cpu.System,
			}, nil
		case regexp.MustCompile(`^/psutil/cpu.*/idle`).MatchString(joinNamespace(ns)):
			return &plugin.PluginMetricType{
				Namespace_: ns,
				Data_:      cpu.Idle,
			}, nil
		}
	}

	return nil, fmt.Errorf("Unknown error processing %v", ns)
}

func getCPUTimesMetricTypes(mts []plugin.PluginMetricType) ([]plugin.PluginMetricType, error) {
	//passing true to CPUTimes indicates per CPU
	//CPUTimes does not currently work on OSX https://github.com/shirou/gopsutil/issues/31
	//TODO test on Linux
	if runtime.GOOS != "darwin" {
		c, err := cpu.CPUTimes(true)
		if err != nil {
			return nil, err
		}
		for _, i := range c {
			cpu := fmt.Sprintf("cpu%s", i.CPU)
			mts = append(mts, plugin.PluginMetricType{Namespace_: []string{"psutil", cpu, "cpu"}})
			mts = append(mts, plugin.PluginMetricType{Namespace_: []string{"psutil", "cpu", "user"}})
			mts = append(mts, plugin.PluginMetricType{Namespace_: []string{"psutil", cpu, "system"}})
			mts = append(mts, plugin.PluginMetricType{Namespace_: []string{"psutil", cpu, "idle"}})
			mts = append(mts, plugin.PluginMetricType{Namespace_: []string{"psutil", cpu, "nice"}})
			mts = append(mts, plugin.PluginMetricType{Namespace_: []string{"psutil", cpu, "iowait"}})
			mts = append(mts, plugin.PluginMetricType{Namespace_: []string{"psutil", cpu, "irq"}})
			mts = append(mts, plugin.PluginMetricType{Namespace_: []string{"psutil", cpu, "softirq"}})
			mts = append(mts, plugin.PluginMetricType{Namespace_: []string{"psutil", cpu, "steal"}})
			mts = append(mts, plugin.PluginMetricType{Namespace_: []string{"psutil", cpu, "guest"}})
			mts = append(mts, plugin.PluginMetricType{Namespace_: []string{"psutil", cpu, "guest_nice"}})
			mts = append(mts, plugin.PluginMetricType{Namespace_: []string{"psutil", cpu, "stolen"}})
		}
	}
	return mts, nil
}
