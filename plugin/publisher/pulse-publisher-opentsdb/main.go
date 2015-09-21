package main

import (
	"os"

	"github.com/intelsdi-x/pulse/control/plugin"
	"github.com/intelsdi-x/pulse/plugin/publisher/pulse-publisher-opentsdb/opentsdb"
)

func main() {
	meta := opentsdb.Meta()
	plugin.Start(meta, opentsdb.NewOpentsdbPublisher(), os.Args[1])
}
