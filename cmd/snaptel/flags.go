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

import (
	"time"

	"github.com/urfave/cli"
)

var (

	// Main flags
	flURL = cli.StringFlag{
		Name:   "url, u",
		Usage:  "Sets the URL to use",
		EnvVar: "SNAP_URL",
		Value:  "http://localhost:8181",
	}
	flAPIVer = cli.StringFlag{
		Name:   "api-version, a",
		Usage:  "The Snap API version",
		EnvVar: "SNAP_API_VERSION",
		Value:  "v1",
	}
	flSecure = cli.BoolFlag{
		Name:   "insecure",
		Usage:  "Ignore certificate errors when Snap's API is running HTTPS",
		EnvVar: "SNAP_INSECURE",
	}
	flRunning = cli.BoolFlag{
		Name:  "running",
		Usage: "Shows running plugins",
	}
	flPassword = cli.BoolFlag{
		Name:   "password, p",
		Usage:  "Require password for REST API authentication",
		EnvVar: "SNAP_REST_PASSWORD",
	}
	flConfig = cli.StringFlag{
		Name:   "config, c",
		EnvVar: "SNAPTEL_CONFIG_PATH,SNAPCTL_CONFIG_PATH",
		Usage:  "Path to a config file",
		Value:  "",
	}
	flTimeout = cli.DurationFlag{
		Name:  "timeout, t",
		Usage: "Timeout to be set on HTTP request to the server",
		Value: 10 * time.Second,
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
	flTaskSchedCount = cli.StringFlag{
		Name:  "count",
		Usage: "The count of runs for the task schedule [defaults to 0 what means no limit, e.g. set to 1 determines a single run task]",
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
		Usage: "The number of consecutive failures before Snap disables the task",
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
