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
	app.Name = "pulse-ctl"
	app.Version = gitversion
	app.Usage = "A powerful telemetry agent framework"
	app.Flags = []cli.Flag{flURL}
	app.Action = mainFlags

	// Parse main flags first
	app.Run(os.Args)

	// Add subcommands and reparse
	app.Commands = commands
	app.Run(os.Args)
}

func mainFlags(c *cli.Context) {
	// Get url main flag
	url := c.String("url")
	// Fall back to a default value if flag and env var have not been provided
	if url == "" {
		url = "http://localhost:8181"
	}
	if pClient == nil {
		pClient = client.New(url, "")
	}
}
