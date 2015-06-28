package main

import (
	"os"
	"time"

	"github.com/codegangsta/cli"
	"github.com/intelsdi-x/pulse/mgmt/rest/client"
)

var (
	gitversion string
	pClient    = client.New(flURL.Value, "")
	timeFormat = time.RFC1123
)

func main() {
	app := cli.NewApp()
	app.Name = "pulse-ctl"
	app.Commands = commands
	app.Version = gitversion
	app.Flags = []cli.Flag{flURL}

	app.Run(os.Args)
}
