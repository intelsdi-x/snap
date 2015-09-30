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

package main

import (
	"os"

	"github.com/intelsdi-x/pulse/control/plugin"
	"github.com/intelsdi-x/pulse/plugin/collector/pulse-collector-pcm/pcm"
)

// Import the pulse plugin library

// Import our collector plugin implementation

// plugin bootstrap
func main() {
	p, err := pcm.New()
	if err != nil {
		panic(err)
	}
	// mts := []plugin.PluginMetricType{
	// 	plugin.PluginMetricType{Namespace_: []string{"intel", "pcm", "L3HIT"}},
	// }
	// time.Sleep(2)
	// mts_, err := p.CollectMetrics(mts)
	// // fmt.Printf("Keys >>>", p.Keys())
	// fmt.Printf("Data >>>", p.Data())
	// fmt.Printf("err >>> %v \n mts >>> %v", err, mts_)
	plugin.Start(
		plugin.NewPluginMeta(pcm.Name, pcm.Version, pcm.Type, []string{}, []string{plugin.PulseGOBContentType}, plugin.ConcurrencyCount(1)),
		p,
		os.Args[1],
	)
}
