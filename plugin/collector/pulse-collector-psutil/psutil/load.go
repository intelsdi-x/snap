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
