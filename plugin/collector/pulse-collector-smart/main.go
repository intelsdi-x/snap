package main

import (
	"os"

	"github.com/intelsdi-x/pulse/control/plugin"
	"github.com/intelsdi-x/pulse/plugin/collector/pulse-collector-smart/smart"
)

func main() {

	smartCollector := &smart.SmartCollector{}

	plugin.Start(plugin.NewPluginMeta(smart.Name, smart.Version, smart.Type, []string{}, []string{plugin.PulseGOBContentType}, plugin.ConcurrencyCount(1)),
		smartCollector,
		os.Args[1],
	)
}
