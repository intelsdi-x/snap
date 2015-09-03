package main

import "github.com/codegangsta/cli"

var (

	// Main flags
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
		Usage: "Interval for the task schedule [ex: 250ms, 1s, 30m]",
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
		Name:  "--no-start",
		Usage: "Do not start task on creation [normally started on creation]",
	}

	// metric
	flMetricVersion = cli.IntFlag{
		Name:  "metric-version, v",
		Usage: "The metric version",
	}
	flMetricNamespace = cli.StringFlag{
		Name:  "metric-namespace, m",
		Usage: "A metric namespace",
	}
)
