package main

import (
	"os"

	"github.com/intelsdi-x/pulse/control/plugin"
	"github.com/intelsdi-x/pulse/plugin/publisher/pulse-publisher-sql/sql"
)

func main() {
	meta := sql.Meta()
	plugin.Start(meta, sql.NewSQLPublisher(), os.Args[1])
}
