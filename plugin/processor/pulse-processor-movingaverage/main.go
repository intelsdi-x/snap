package main

import (
	"os"

	"github.com/intelsdi-x/pulse/control/plugin"
	"github.com/intelsdi-x/pulse/plugin/processor/pulse-processor-movingaverage/movingaverage"
)

func main() {
	meta := movingaverage.Meta()
	plugin.Start(meta, movingaverage.NewMovingaverageProcessor(), os.Args[1])
}
