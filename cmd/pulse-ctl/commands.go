package main

import (
	"fmt"
	"strings"
	"text/tabwriter"

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
					Flags: []cli.Flag{
						taskName,
					},
				},
				{
					Name:   "list",
					Usage:  "list",
					Action: listTask,
				},
				{
					Name:   "start",
					Usage:  "start <task_id>",
					Action: startTask,
				},
				{
					Name:   "stop",
					Usage:  "stop <task_id>",
					Action: stopTask,
				},
				{
					Name:   "remove",
					Usage:  "remove <task_id>",
					Action: removeTask,
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
				{
					Name:   "list",
					Usage:  "list",
					Action: listPlugins,
					Flags: []cli.Flag{
						flRunning,
					},
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

func printFields(tw *tabwriter.Writer, indent bool, width int, fields ...interface{}) {
	argArray := make([]interface{}, 0)
	if indent {
		argArray = append(argArray, strings.Repeat(" ", width))
	}
	for i, field := range fields {
		argArray = append(argArray, field)
		if i < (len(fields) - 1) {
			argArray = append(argArray, "\t")
		}
	}
	fmt.Fprintln(tw, argArray...)
}
