package main

import (
	"github.com/intelsdilabs/pulse/control/plugin"
	"github.com/intelsdilabs/pulse/plugin/collector/pulse-collector-facter/facter"
)

func main() {
	// init plugin rpc server
	server, err := plugin.NewServer()
	if err != nil {
		panic(err)
	}

	// Three things provided:
	//   the implementation satfiying plugin.CollectorPlugin
	//   the collector configuration policy satifying plugin.ConfigRules
	policy := facter.ConfigPolicy()
	server.StartCollector(facter.Name, facter.Version, new(facter.Facter), policy)
}
