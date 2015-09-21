package main

import (
	"os"

	"github.com/intelsdi-x/pulse/control/plugin"
	"github.com/intelsdi-x/pulse/plugin/collector/pulse-collector-perfevents/perfevents"
)

// plugin bootstrap
func main() {
	plugin.Start(
		plugin.NewPluginMeta(perfevents.Name, perfevents.Version, perfevents.Type, []string{}, []string{plugin.PulseGOBContentType}, plugin.ConcurrencyCount(1)),
		perfevents.NewPerfevents(),
		os.Args[1],
	)
}
