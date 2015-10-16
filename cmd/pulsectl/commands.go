/*
http://www.apache.org/licenses/LICENSE-2.0.txt


Copyright 2015 Intel Corporation

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

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
