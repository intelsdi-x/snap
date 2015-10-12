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

	"github.com/intelsdi-x/pulse/control/plugin"
	"github.com/shirou/gopsutil/net"
)

var netIOCounterLabels = []string{
	"bytes_sent",
	"bytes_recv",
	"packets_sent",
	"packets_recv",
	"errin",
	"errout",
	"dropin",
	"dropout",
}

func netIOCounters(ns []string) (*plugin.PluginMetricType, error) {
	nets, err := net.NetIOCounters(true)
	if err != nil {
		return nil, err
	}

	for _, net := range nets {
		switch {
		case regexp.MustCompile(`^/psutil/net/.*/bytes_sent$`).MatchString(joinNamespace(ns)):
			return &plugin.PluginMetricType{
				Namespace_: ns,
				Data_:      net.BytesSent,
			}, nil
		case regexp.MustCompile(`^/psutil/net/.*/bytes_recv`).MatchString(joinNamespace(ns)):
			return &plugin.PluginMetricType{
				Namespace_: ns,
				Data_:      net.BytesRecv,
			}, nil
		case regexp.MustCompile(`^/psutil/net/.*/packets_sent`).MatchString(joinNamespace(ns)):
			return &plugin.PluginMetricType{
				Namespace_: ns,
				Data_:      net.BytesSent,
			}, nil
		case regexp.MustCompile(`^/psutil/net/.*/packets_recv`).MatchString(joinNamespace(ns)):
			return &plugin.PluginMetricType{
				Namespace_: ns,
				Data_:      net.BytesRecv,
			}, nil
		case regexp.MustCompile(`^/psutil/net/.*/errin`).MatchString(joinNamespace(ns)):
			return &plugin.PluginMetricType{
				Namespace_: ns,
				Data_:      net.Errin,
			}, nil
		case regexp.MustCompile(`^/psutil/net/.*/errout`).MatchString(joinNamespace(ns)):
			return &plugin.PluginMetricType{
				Namespace_: ns,
				Data_:      net.Errout,
			}, nil
		case regexp.MustCompile(`^/psutil/net/.*/dropin`).MatchString(joinNamespace(ns)):
			return &plugin.PluginMetricType{
				Namespace_: ns,
				Data_:      net.Dropin,
			}, nil
		case regexp.MustCompile(`^/psutil/net/.*/dropout`).MatchString(joinNamespace(ns)):
			return &plugin.PluginMetricType{
				Namespace_: ns,
				Data_:      net.Dropout,
			}, nil
		}
	}

	return nil, fmt.Errorf("Unknown error processing %v", ns)
}

func getNetIOCounterMetricTypes() ([]plugin.PluginMetricType, error) {
	mts := make([]plugin.PluginMetricType, 0)
	nets, err := net.NetIOCounters(false)
	if err != nil {
		return nil, err
	}
	//total for all nics
	for _, name := range netIOCounterLabels {
		mts = append(mts, plugin.PluginMetricType{Namespace_: []string{"psutil", "net", nets[0].Name, name}})
	}
	//per nic
	nets, err = net.NetIOCounters(true)
	if err != nil {
		return nil, err
	}
	for _, net := range nets {
		for _, name := range netIOCounterLabels {
			mts = append(mts, plugin.PluginMetricType{Namespace_: []string{"psutil", "net", net.Name, name}})
		}
	}
	return mts, nil
}
