package main

import (
	"os"

	"github.com/intelsdi-x/pulse/control/plugin"
	"github.com/intelsdi-x/pulse/plugin/processor/pulse-processor-computation/computation"
)

func main() {
	meta := computation.Meta()
	plugin.Start(meta, computation.NewComputationPublisher(), os.Args[1])
}
