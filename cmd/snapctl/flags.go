/*
http://www.apache.org/licenses/LICENSE-2.0.txt


Copyright 2015,2016 Intel Corporation

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

import "github.com/codegangsta/cli"

var (

	// Main flags
	flURL = cli.StringFlag{
		Name:   "url, u",
		Usage:  "Sets the URL to use",
		EnvVar: "SNAP_URL",
		Value:  "http://localhost:8181",
	}
	flAPIVer = cli.StringFlag{
		Name:  "api-version, a",
		Usage: "The snap API version",
		Value: "v1",
	}
	flSecure = cli.BoolFlag{
		Name:  "insecure",
		Usage: "Ignore certificate errors when snap's API is running HTTPS",
	}
	flRunning = cli.BoolFlag{
		Name:  "running",
		Usage: "Shows running plugins",
	}
	flPassword = cli.BoolFlag{
		Name:  "password, p",
		Usage: "Password for REST API authentication",
	}
	flConfig = cli.StringFlag{
		Name:   "config, c",
		EnvVar: "SNAPCTL_CONFIG_PATH",
		Usage:  "Path to a config file",
		Value:  "",
	}

	// Plugin flags
	flPluginAsc = cli.StringFlag{
		Name:  "plugin-asc, a",
		Usage: "The plugin asc",
	}
	flPluginType = cli.StringFlag{
		Name:  "plugin-type, t",
		Usage: "The plugin type",
	}
	flPluginName = cli.StringFlag{
		Name:  "plugin-name, n",
		Usage: "The plugin name",
	}
	flPluginVersion = cli.IntFlag{
		Name:  "plugin-version, v",
		Usage: "The plugin version",
	}

	// Task flags
	flTaskName = cli.StringFlag{
		Name:  "name, n",
		Usage: "Optional requirement for giving task names",
		Value: "",
	}
	flTaskManifest = cli.StringFlag{
		Name:  "task-manifest, t",
		Usage: "File path for task manifest to use for task creation.",
	}

	flWorkfowManifest = cli.StringFlag{
		Name:  "workflow-manifest, w",
		Usage: "File path for workflow manifest to use for task creation",
	}

	flTaskSchedInterval = cli.StringFlag{
		Name:  "interval, i",
		Usage: "Interval for the task schedule [ex (simple schedule): 250ms, 1s, 30m (cron schedule): \"0 * * * * *\"]",
	}

	flTaskSchedStartTime = cli.StringFlag{
		Name:  "start-time",
		Usage: "Start time for the task schedule [defaults to now]",
	}
	flTaskSchedStopTime = cli.StringFlag{
		Name:  "stop-time",
		Usage: "Start time for the task schedule [defaults to now]",
	}

	flTaskSchedStartDate = cli.StringFlag{
		Name:  "start-date",
		Usage: "Start date for the task schedule [defaults to today]",
	}
	flTaskSchedStopDate = cli.StringFlag{
		Name:  "stop-date",
		Usage: "Stop date for the task schedule [defaults to today]",
	}
	flTaskSchedDuration = cli.StringFlag{
		Name:  "duration, d",
		Usage: "The amount of time to run the task [appends to start or creates a start time before a stop]",
	}
	flTaskSchedNoStart = cli.BoolFlag{
		Name:  "no-start",
		Usage: "Do not start task on creation [normally started on creation]",
	}
	flTaskDeadline = cli.StringFlag{
		Name:  "deadline",
		Usage: "The deadline for the task to be killed after started if the task runs too long (All tasks default to 5s)",
	}
	flTaskMaxFailures = cli.StringFlag{
		Name:  "max-failures",
		Usage: "The number of consecutive failures before snap disables the task",
	}

	// metric
	flMetricVersion = cli.IntFlag{
		Name:  "metric-version, v",
		Usage: "The metric version. 0 (default) returns all. -1 returns latest.",
	}
	flMetricNamespace = cli.StringFlag{
		Name:  "metric-namespace, m",
		Usage: "A metric namespace",
	}

	// general
	flVerbose = cli.BoolFlag{
		Name:  "verbose",
		Usage: "Verbose output",
	}
)
