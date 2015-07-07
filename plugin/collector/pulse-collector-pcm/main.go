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
		plugin.NewPluginMeta(pcm.Name, pcm.Version, pcm.Type, []string{}, []string{plugin.PulseGOBContentType}),
		p,
		os.Args[1],
	)
}
