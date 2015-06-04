package main

import (
	"fmt"
	"os"

	"github.com/codegangsta/cli"
)

var (
	commands = []cli.Command{
		{
			Name: "task",
			Subcommands: []cli.Command{
				{
					Name:   "create",
					Usage:  "create <task file json|yaml>",
					Action: createTask,
				},
			},
		},
		{
			Name: "plugin",
			Subcommands: []cli.Command{
				{
					Name:   "load",
					Usage:  "load <plugin path>",
					Action: loadPlugin,
				},
			},
		},
		{
			Name: "metric",
			Subcommands: []cli.Command{
				{
					Name:   "list",
					Usage:  "list",
					Action: listMetrics,
				},
			},
		},
	}
)

func version(ctx *cli.Context) {
	fmt.Println(gitversion)
}

func doSomething(ctx *cli.Context) {
	fmt.Println("doing something")
}

func loadPlugin(ctx *cli.Context) {
	if len(ctx.Args()) != 1 {
		fmt.Print("Incorrect usage\n")
		os.Exit(1)
	}

	err := client.LoadPlugin(ctx.Args().First())
	if err != nil {
		fmt.Printf("Error: %v\n", err.Error())
		os.Exit(1)
	}
}
