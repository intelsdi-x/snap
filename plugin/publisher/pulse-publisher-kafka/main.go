package main

import (
	"os"

	// Import the pulse plugin library
	"github.com/intelsdi-x/pulse/control/plugin"
	"github.com/intelsdi-x/pulse/plugin/publisher/pulse-publisher-kafka/kafka"
)

func main() {
	// Three things provided:
	//   the definition of the plugin metadata
	//   the implementation satfiying plugin.CollectorPlugin
	//   the collector configuration policy satifying plugin.ConfigRules

	// Define metadata about Plugin
	meta := kafka.Meta()

	// Start a collector
	plugin.Start(meta, kafka.NewKafkaPublisher(), os.Args[1])
}
