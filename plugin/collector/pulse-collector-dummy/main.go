package main

import (
	"os"

	// Import the pulse plugin library
	"github.com/intelsdilabs/pulse/control/plugin"
	// Import our collector plugin implementation
	"github.com/intelsdilabs/pulse/plugin/collector/pulse-collector-dummy/dummy"
)

func main() {
	// Three things provided:
	//   the definition of the plugin metadata
	//   the implementation satfiying plugin.CollectorPlugin
	//   the collector configuration policy satifying plugin.ConfigRules

	// Define default policy
	policy := dummy.ConfigPolicy()

	// Define metadata about Plugin
	meta := dummy.Meta()

	// Start a collector
	e := plugin.StartCollector(meta, new(dummy.Dummy), policy)
	if e != nil {
		os.Exit(2)
	}
}
