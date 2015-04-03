package main

import (
	"os"

	// Import the pulse plugin library
	"github.com/intelsdilabs/pulse/control/plugin"
	// Import our collector plugin implementation
	"github.com/intelsdilabs/pulse/plugin/collector/pulse-collector-dummy/dummy"
)

func main() {
	// Provided:
	//   the definition of the plugin metadata
	//   the implementation satfiying plugin.CollectorPlugin

	// Define metadata about Plugin
	meta := dummy.Meta()

	// Start a collector
	plugin.Start(meta, new(dummy.Dummy), os.Args[1])
}
