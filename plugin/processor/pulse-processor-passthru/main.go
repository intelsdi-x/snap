package main

import (
	"os"

	"github.com/intelsdi-x/pulse/control/plugin"
	"github.com/intelsdi-x/pulse/plugin/processor/pulse-processor-passthru/passthru"
)

func main() {
	meta := passthru.Meta()
	plugin.Start(meta, passthru.NewPassthruPublisher(), os.Args[1])
}
