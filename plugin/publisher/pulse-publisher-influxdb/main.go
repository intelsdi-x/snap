package main

import (
	"os"

	"github.com/intelsdilabs/pulse/control/plugin"
	"github.com/intelsdilabs/pulse/plugin/publisher/pulse-publisher-influxdb/influx"
)

func main() {
	meta := influx.Meta()
	plugin.Start(meta, influx.NewInfluxPublisher(), os.Args[1])
}
