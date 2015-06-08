package main

import "github.com/codegangsta/cli"

var (
	flURL = cli.StringFlag{
		Name:   "url",
		Usage:  "Sets the URL to use",
		EnvVar: "PULSE_URL",
		Value:  "http://localhost:8181",
	}
	flRunning = cli.BoolFlag{
		Name:  "running",
		Usage: "Shows running plugins",
	}
)
