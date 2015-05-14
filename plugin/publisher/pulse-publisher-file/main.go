package main

import (
	"os"

	"github.com/intelsdi-x/pulse/control/plugin"
	"github.com/intelsdi-x/pulse/plugin/publisher/pulse-publisher-file/file"
)

func main() {
	meta := file.Meta()
	plugin.Start(meta, file.NewFilePublisher(), os.Args[1])
}
