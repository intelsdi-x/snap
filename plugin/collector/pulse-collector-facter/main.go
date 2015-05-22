package main

import (
	"os"
	// Import the pulse plugin library
	"github.com/intelsdi-x/pulse/control/plugin"
	// Import our collector plugin implementation
	"github.com/intelsdi-x/pulse/plugin/collector/pulse-collector-facter/facter"
)

// meta date about plugin
const (
	name       = "Intel Fact Gathering Plugin"
	version    = 1
	pluginType = plugin.CollectorPluginType
)

// plugin bootstrap
func main() {
	plugin.Start(
		plugin.NewPluginMeta(name, version, pluginType, []string{plugin.PulseGOBContentType}, []string{plugin.PulseGOBContentType}),
		facter.NewFacter(), // CollectorPlugin interface
		os.Args[1],
	)
}
