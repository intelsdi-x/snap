package main

import (
	"os"

	// Import the pulse plugin library
	"github.com/intelsdi-x/pulse/control/plugin"
	// Import our collector plugin implementation
	"github.com/intelsdi-x/pulse/plugin/collector/pulse-collector-dummy2/dummy"
)

func main() {
	// Provided:
	//   the definition of the plugin metadata
	//   the implementation satfiying plugin.CollectorPlugin

	// Define metadata about Plugin
	meta := dummy.Meta()
	// meta.RPCType = plugin.JSONRPC

	// Start a collector
	plugin.Start(meta, new(dummy.Dummy), os.Args[1])
}
