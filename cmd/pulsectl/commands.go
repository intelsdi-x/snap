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
					Name:        "create",
					Description: "Creates a new task in the pulse scheduler",
					Usage:       "There are two ways to create a task.\n\t1) Use a task manifest with [--task-manifest]\n\t2) Provide a workflow manifest and schedule details.\n\n\t* Note: Start and stop date/time are optional.\n",
					Action:      createTask,
					Flags: []cli.Flag{
						flTaskManifest,
						flWorkfowManifest,
						flTaskSchedInterval,
						flTaskSchedStartDate,
						flTaskSchedStartTime,
						flTaskSchedStopDate,
						flTaskSchedStopTime,
						flTaskName,
						flTaskSchedDuration,
						flTaskSchedNoStart,
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
				{
					Name:   "export",
					Usage:  "export <task_id>",
					Action: exportTask,
				},
				{
					Name:   "watch",
					Usage:  "watch <task_id>",
					Action: watchTask,
				},
				{
					Name:   "enable",
					Usage:  "enable <task_id>",
					Action: enableTask,
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
					Flags: []cli.Flag{
						flPluginAsc,
					},
				},
				{
					Name:   "unload",
					Usage:  "unload",
					Action: unloadPlugin,
					Flags: []cli.Flag{
						flPluginType,
						flPluginName,
						flPluginVersion,
					},
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
					Flags: []cli.Flag{
						flMetricVersion,
						flMetricNamespace,
					},
				},
				{
					Name:   "get",
					Usage:  "get details on a single metric",
					Action: getMetric,
					Flags: []cli.Flag{
						flMetricVersion,
						flMetricNamespace,
					},
				},
			},
		},
	}
)

func printFields(tw *tabwriter.Writer, indent bool, width int, fields ...interface{}) {
	var argArray []interface{}
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
