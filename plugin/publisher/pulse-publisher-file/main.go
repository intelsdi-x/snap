package main

import (
	"os"

	"github.com/intelsdilabs/pulse/control/plugin"
	"github.com/intelsdilabs/pulse/plugin/publisher/pulse-publisher-file/file"
)

func main() {
	meta := file.Meta()
	plugin.Start(meta, file.NewFilePublisher(), os.Args[1])
}
