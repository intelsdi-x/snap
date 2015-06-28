package main

import (
	"os"

	"github.com/codegangsta/cli"
	"github.com/intelsdi-x/pulse/mgmt/rest/client"
)

var gitversion string

var pClient = client.New(flURL.Value, "")

func main() {
	app := cli.NewApp()
	app.Name = "pulse-ctl"
	app.Commands = commands
	app.Version = gitversion
	app.Flags = []cli.Flag{flURL}

	app.Run(os.Args)
}
