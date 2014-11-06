package main

import (
	// Import the pulse plugin library
	"github.com/intelsdilabs/pulse/control/plugin"
	// Import our collector plugin implementation
	"github.com/intelsdilabs/pulse/plugin/collector/pulse-collector-facter/facter"
)

func main() {
	// Three things provided:
	//   the definition of the plugin metadata
	//   the implementation satfiying plugin.CollectorPlugin
	//   the collector configuration policy satifying plugin.ConfigRules

	// Define default policy
	policy := facter.ConfigPolicy()

	// Define metadata about Plugin
	meta := facter.Meta()

	// Start a collector
	plugin.StartCollector(meta, new(facter.Facter), policy)
}
