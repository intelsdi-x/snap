package main

import (
	"os"
	"time"

	"github.com/codegangsta/cli"
	"github.com/intelsdi-x/pulse/mgmt/rest/client"
)

var (
	gitversion string
	pClient    *client.Client
	timeFormat = time.RFC1123
)

func main() {

	app := cli.NewApp()
	app.Name = "pulsectl"
	app.Version = gitversion
	app.Usage = "A powerful telemetry agent framework"
	app.Flags = []cli.Flag{flURL, flSecure, flAPIVer}
	app.Commands = commands

	app.Before = func(c *cli.Context) error {
		if pClient == nil {
			pClient = client.New(c.GlobalString("url"), c.GlobalString("api-version"), c.GlobalBool("insecure"))
		}
		return nil
	}

	app.Run(os.Args)
}
