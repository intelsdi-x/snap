package main

import (
	"os"

	"github.com/intelsdilabs/pulse/control/plugin"
	"github.com/intelsdilabs/pulse/plugin/collector/pulse-collector-libcontainer/lcplugin"
)

func main() {
	// Three things provided:
	//   the definition of the plugin metadata
	//   the implementation satfiying plugin.CollectorPlugin
	//   the collector configuration policy satifying plugin.ConfigRules

	// Define default policy
	policy := lcplugin.ConfigPolicy()

	// Define metadata about Plugin
	meta := lcplugin.Meta()

	// Start a collector
	//plugin.StartCollector(meta, new(facter.Facter), policy, os.Args[0], os.Args[1])
	plugin.Start(meta, new(lcplugin.Libcontainer), policy, os.Args[1])
}
