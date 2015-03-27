package main

import (
	"os"

	// Import the pulse plugin library
	"github.com/intelsdilabs/pulse/control/plugin"
	// Import our collector plugin implementation
	"github.com/intelsdilabs/pulse/plugin/collector/pulse-collector-facter/facter"
)

func main() {
	// Provided:
	//   the definition of the plugin metadata
	//   the implementation satfiying plugin.CollectorPlugin

	// Define metadata about Plugin
	meta := facter.Meta()

	// Start a collector
	//plugin.StartCollector(meta, new(facter.Facter), policy, os.Args[0], os.Args[1])
	plugin.Start(meta, new(facter.Facter), os.Args[1])
}
