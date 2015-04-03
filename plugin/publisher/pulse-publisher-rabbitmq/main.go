package main

import (
	"os"

	// Import the pulse plugin library
	"github.com/intelsdilabs/pulse/control/plugin"
	"github.com/intelsdilabs/pulse/plugin/publisher/pulse-publisher-rabbitmq/rmq"

	// Import our publisher plugin implementation
)

// docker run -d -p 5672:5672 -p 15672:15672 dockerfile/rabbitmq

func main() {
	// Three things provided:
	//   the definition of the plugin metadata
	//   the implementation satfiying plugin.CollectorPlugin
	//   the collector configuration policy satifying plugin.ConfigRules

	// Define metadata about Plugin
	meta := rmq.Meta()

	// Start a collector
	plugin.Start(meta, rmq.NewRmqPublisher(), os.Args[1])
}
