package main

import (
	"os"

	"github.com/intelsdi-x/pulse/control/plugin"
	"github.com/intelsdi-x/pulse/plugin/publisher/pulse-publisher-mysql/mysql"
)

func main() {
	meta := mysql.Meta()
	plugin.Start(meta, mysql.NewMySQLPublisher(), os.Args[1])
}
