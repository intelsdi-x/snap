package main

import (
	"github.com/intelsdilabs/pulse/control/plugin"
	"github.com/intelsdilabs/pulse/plugin/collector/pulse-collector-facter/facter"
)

func main() {
	// Three things provided:
	//   the implementation satfiying plugin.CollectorPlugin
	//   the collector configuration policy satifying plugin.ConfigRules
	policy := facter.ConfigPolicy()
	plugin.StartCollector(facter.Name, facter.Version, new(facter.Facter), policy)
}
