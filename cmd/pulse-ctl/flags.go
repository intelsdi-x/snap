package main

import "github.com/codegangsta/cli"

var (
	flURL = cli.StringFlag{
		Name:   "url, u",
		Usage:  "Sets the URL to use",
		EnvVar: "PULSE_URL",
		Value:  "http://localhost:8181",
	}
	flRunning = cli.BoolFlag{
		Name:  "running",
		Usage: "Shows running plugins",
	}

	flTaskName = cli.StringFlag{
		Name:  "name, n",
		Usage: "Optional requirement for giving task names",
		Value: "",
	}
	// plugin
	flPluginName = cli.StringFlag{
		Name:  "plugin-name, n",
		Usage: "The plugin name",
	}
	flPluginVersion = cli.IntFlag{
		Name:  "plugin-version, v",
		Usage: "The plugin version",
	}
)
