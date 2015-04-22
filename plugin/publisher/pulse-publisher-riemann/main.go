package main

import (
	"os"

	"github.com/intelsdilabs/pulse/control/plugin"
	"github.com/intelsdilabs/pulse/plugin/publisher/pulse-publisher-riemann/riemann"
)

func main() {
	// Three things are provided:
	// - The definition of the plugin metadata
	// - The implementation satisfying plugin.PublisherPlugin
	// - The publisher config policy satisfying plugin.ConfigRules

	// Define metadata about the plugin
	meta := riemann.Meta()

	// Start a publisher
	plugin.Start(meta, riemann.NewRiemannPublisher(), os.Args[1])
}
