/*
http://www.apache.org/licenses/LICENSE-2.0.txt


Copyright 2015 Intel Coporation

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package psutil

import (
	"fmt"
	"regexp"
	"runtime"

	"github.com/intelsdi-x/pulse/control/plugin"
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
		case regexp.MustCompile(`^/psutil/cpu.*/nice`).MatchString(joinNamespace(ns)):
			return &plugin.PluginMetricType{
				Namespace_: ns,
				Data_:      cpu.Nice,
			}, nil
		case regexp.MustCompile(`^/psutil/cpu.*/iowait`).MatchString(joinNamespace(ns)):
			return &plugin.PluginMetricType{
				Namespace_: ns,
				Data_:      cpu.Iowait,
			}, nil
		case regexp.MustCompile(`^/psutil/cpu.*/irq`).MatchString(joinNamespace(ns)):
			return &plugin.PluginMetricType{
				Namespace_: ns,
				Data_:      cpu.Irq,
			}, nil
		case regexp.MustCompile(`^/psutil/cpu.*/softirq`).MatchString(joinNamespace(ns)):
			return &plugin.PluginMetricType{
				Namespace_: ns,
				Data_:      cpu.Softirq,
			}, nil
		case regexp.MustCompile(`^/psutil/cpu.*/steal`).MatchString(joinNamespace(ns)):
			return &plugin.PluginMetricType{
				Namespace_: ns,
				Data_:      cpu.Steal,
			}, nil
		case regexp.MustCompile(`^/psutil/cpu.*/guest`).MatchString(joinNamespace(ns)):
			return &plugin.PluginMetricType{
				Namespace_: ns,
				Data_:      cpu.Guest,
			}, nil
		case regexp.MustCompile(`^/psutil/cpu.*/guest_nice`).MatchString(joinNamespace(ns)):
			return &plugin.PluginMetricType{
				Namespace_: ns,
				Data_:      cpu.GuestNice,
			}, nil
		case regexp.MustCompile(`^/psutil/cpu.*/stolen`).MatchString(joinNamespace(ns)):
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
