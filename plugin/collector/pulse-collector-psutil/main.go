package main

import (
	"os"
	// Import the pulse plugin library
	"github.com/intelsdi-x/pulse/control/plugin"
	// Import our collector plugin implementation
	"github.com/intelsdi-x/pulse/plugin/collector/pulse-collector-psutil/psutil"
)

// plugin bootstrap
func main() {
	plugin.Start(
		plugin.NewPluginMeta(psutil.Name, psutil.Version, psutil.Type, []string{}, []string{plugin.PulseGOBContentType}),
		&psutil.Psutil{}, // CollectorPlugin interface
		os.Args[1],
	)
}
