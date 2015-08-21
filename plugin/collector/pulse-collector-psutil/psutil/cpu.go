package psutil

import (
	"fmt"
	"regexp"
	"runtime"

	"github.com/intelsdi-x/pulse/control/plugin"
	"github.com/intelsdi-x/pulse/core"
	"github.com/shirou/gopsutil/cpu"
)

var cpuLabels = []string{"user", "system", "idle", "nice", "iowait",
	"irq", "softirq", "steal", "guest", "guest_nice", "stolen"}

func cpuTimes(ns []string) (*plugin.PluginMetricType, error) {
	cpus, err := cpu.CPUTimes(true)
	if err != nil {
		return nil, err
	}

	for _, cpu := range cpus {
		switch {
		case regexp.MustCompile(`^/psutil/cpu.*/user`).MatchString(core.JoinNamespace(ns)):
			return &plugin.PluginMetricType{
				Namespace_: ns,
				Data_:      cpu.User,
			}, nil
		case regexp.MustCompile(`^/psutil/cpu.*/system`).MatchString(core.JoinNamespace(ns)):
			return &plugin.PluginMetricType{
				Namespace_: ns,
				Data_:      cpu.System,
			}, nil
		case regexp.MustCompile(`^/psutil/cpu.*/idle`).MatchString(core.JoinNamespace(ns)):
			return &plugin.PluginMetricType{
				Namespace_: ns,
				Data_:      cpu.Idle,
			}, nil
		case regexp.MustCompile(`^/psutil/cpu.*/nice`).MatchString(core.JoinNamespace(ns)):
			return &plugin.PluginMetricType{
				Namespace_: ns,
				Data_:      cpu.Nice,
			}, nil
		case regexp.MustCompile(`^/psutil/cpu.*/iowait`).MatchString(core.JoinNamespace(ns)):
			return &plugin.PluginMetricType{
				Namespace_: ns,
				Data_:      cpu.Iowait,
			}, nil
		case regexp.MustCompile(`^/psutil/cpu.*/irq`).MatchString(core.JoinNamespace(ns)):
			return &plugin.PluginMetricType{
				Namespace_: ns,
				Data_:      cpu.Irq,
			}, nil
		case regexp.MustCompile(`^/psutil/cpu.*/softirq`).MatchString(core.JoinNamespace(ns)):
			return &plugin.PluginMetricType{
				Namespace_: ns,
				Data_:      cpu.Softirq,
			}, nil
		case regexp.MustCompile(`^/psutil/cpu.*/steal`).MatchString(core.JoinNamespace(ns)):
			return &plugin.PluginMetricType{
				Namespace_: ns,
				Data_:      cpu.Steal,
			}, nil
		case regexp.MustCompile(`^/psutil/cpu.*/guest`).MatchString(core.JoinNamespace(ns)):
			return &plugin.PluginMetricType{
				Namespace_: ns,
				Data_:      cpu.Guest,
			}, nil
		case regexp.MustCompile(`^/psutil/cpu.*/guest_nice`).MatchString(core.JoinNamespace(ns)):
			return &plugin.PluginMetricType{
				Namespace_: ns,
				Data_:      cpu.GuestNice,
			}, nil
		case regexp.MustCompile(`^/psutil/cpu.*/stolen`).MatchString(core.JoinNamespace(ns)):
			return &plugin.PluginMetricType{
				Namespace_: ns,
				Data_:      cpu.Stolen,
			}, nil
		}

	}

	return nil, fmt.Errorf("Unknown error processing %v", ns)
}

func getCPUTimesMetricTypes() ([]plugin.PluginMetricType, error) {
	//passing true to CPUTimes indicates per CPU
	//CPUTimes does not currently work on OSX https://github.com/shirou/gopsutil/issues/31
	mts := make([]plugin.PluginMetricType, 0)
	if runtime.GOOS != "darwin" {
		c, err := cpu.CPUTimes(true)
		if err != nil {
			return nil, err
		}
		for _, i := range c {
			for _, label := range cpuLabels {
				mts = append(mts, plugin.PluginMetricType{Namespace_: []string{"psutil", i.CPU, label}})
			}
		}
	}
	return mts, nil
}
