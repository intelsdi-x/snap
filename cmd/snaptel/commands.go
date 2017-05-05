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

	"github.com/urfave/cli"
)

var (
	commands = []cli.Command{
		{
			Name: "task",
			Subcommands: []cli.Command{
				{
					Name:        "create",
					Description: "Creates a new task in the snap scheduler",
					Usage:       "There are two ways to create a task.\n\t1) Use a task manifest with [--task-manifest]\n\t2) Provide a workflow manifest and schedule details.\n\n\t* Note: Start and stop date/time are optional.\n",
					Action:      createTask,
					Flags: []cli.Flag{
						flTaskManifest,
						flWorkfowManifest,
						flTaskSchedInterval,
						flTaskSchedCount,
						flTaskSchedStartDate,
						flTaskSchedStartTime,
						flTaskSchedStopDate,
						flTaskSchedStopTime,
						flTaskName,
						flTaskSchedDuration,
						flTaskSchedNoStart,
						flTaskDeadline,
						flTaskMaxFailures,
					},
				},
				{
					Name:   "list",
					Usage:  "list",
					Action: listTask,
					Flags: []cli.Flag{
						flVerbose,
					},
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
					Flags: []cli.Flag{
						flVerbose,
					},
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
					Usage:  "load <plugin_path> [--plugin-cert=<plugin_cert_path> --plugin-key=<plugin_key_path> --plugin-ca-certs=<ca_cert_paths>]",
					Action: loadPlugin,
					Flags: []cli.Flag{
						flPluginAsc,
						flPluginCert,
						flPluginKey,
						flPluginCACerts,
					},
				},
				{
					Name:   "unload",
					Usage:  "unload <plugin_type> <plugin_name> <plugin_version>",
					Action: unloadPlugin,
				},
				{
					Name:   "swap",
					Usage:  "swap <load_plugin_path> <unload_plugin_type>:<unload_plugin_name>:<unload_plugin_version> or swap <load_plugin_path> -t <unload_plugin_type> -n <unload_plugin_name> -v <unload_plugin_version> [--plugin-cert=<plugin_cert_path> --plugin-key=<plugin_key_path> --plugin-ca-certs=<ca_cert_paths>]",
					Action: swapPlugins,
					Flags: []cli.Flag{
						flPluginAsc,
						flPluginType,
						flPluginName,
						flPluginVersion,
						flPluginCert,
						flPluginKey,
						flPluginCACerts,
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
				{
					Name: "config",
					Subcommands: []cli.Command{
						{
							Name:   "get",
							Usage:  "get <plugin_type>:<plugin_name>:<plugin_version> or get -t <plugin_type> -n <plugin_name> -v <plugin_version>",
							Action: getConfig,
							Flags: []cli.Flag{
								flPluginName,
								flPluginType,
								flPluginVersion,
							},
						},
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
						flVerbose,
					},
				},
				{
					Name:   "get",
					Usage:  "get details on metric(s)",
					Action: getMetric,
					Flags: []cli.Flag{
						flMetricVersion,
						flMetricNamespace,
					},
				},
			},
		},
	}
	tribeWarning  = "Can only be used when tribe mode is enabled."
	tribeCommands = []cli.Command{
		{
			Name:  "member",
			Usage: tribeWarning,
			Subcommands: []cli.Command{
				{
					Name:   "list",
					Usage:  "list" + tribeWarning,
					Action: listMembers,
				},
				{
					Name:   "show",
					Usage:  "show <member_name>",
					Action: showMember,
					Flags:  []cli.Flag{flVerbose},
				},
			},
		},
		{
			Name:  "agreement",
			Usage: tribeWarning,
			Subcommands: []cli.Command{
				{
					Name:   "list",
					Usage:  "list",
					Action: listAgreements,
				},
				{
					Name:   "create",
					Usage:  "create <agreement_name>",
					Action: createAgreement,
				},
				{
					Name:   "delete",
					Usage:  "delete <agreement_name>",
					Action: deleteAgreement,
				},
				{
					Name:   "join",
					Usage:  "join <agreement_name> <member_name>",
					Action: joinAgreement,
				},
				{
					Name:   "leave",
					Usage:  "leave <agreement_name> <member_name>",
					Action: leaveAgreement,
				},
				{
					Name:   "members",
					Usage:  "members <agreement_name>",
					Action: agreementMembers,
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
		if field != nil {
			argArray = append(argArray, field)
		} else {
			argArray = append(argArray, "")
		}
		if i < (len(fields) - 1) {
			argArray = append(argArray, "\t")
		}
	}
	fmt.Fprintln(tw, argArray...)
}

// ByCommand contains array of CLI commands.
type ByCommand []cli.Command

func (s ByCommand) Len() int {
	return len(s)
}
func (s ByCommand) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}
func (s ByCommand) Less(i, j int) bool {
	if s[i].Name == "help" {
		return false
	}
	if s[j].Name == "help" {
		return true
	}
	return s[i].Name < s[j].Name
}
