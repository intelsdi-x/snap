package main

import (
	"os"

	"github.com/intelsdi-x/pulse/control/plugin"
	"github.com/intelsdi-x/pulse/plugin/publisher/pulse-publisher-influxdb/influx"
)

func main() {
	meta := influx.Meta()
	plugin.Start(meta, influx.NewInfluxPublisher(), os.Args[1])
}
