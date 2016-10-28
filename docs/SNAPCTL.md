<!--
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
-->

# snapctl
A powerful telemetry framework

## Usage
```
$ $SNAP_PATH/bin/snapctl [global options] command [command options] [arguments...]
```
### Global Options
```
--url, -u 'http://localhost:8181'    Sets the URL to use [$SNAP_URL]
--insecure                           Ignore certificate errors when snap's API is running HTTPS
--api-version, -a 'v1'               The snap API version
--password, -p			             Password for REST API authentication
--config, -c 			             Path to a config file [$SNAPCTL_CONFIG_PATH]
--help, -h                           show help
--version, -v                        print the version
```
### Commands
```
metric
plugin
task
help, h      Shows a list of commands or help for one command
```
### Command Options
#### task
```
$ $SNAP_PATH/bin/snapctl task command [command options] [arguments...]
```
```
create      There are two ways to create a task.
                1) Use a task manifest with [--task-manifest, t]
                2) Provide a workflow manifest and schedule details [--workflow-manifest, -w]

               --task-manifest, -t          File path for task manifest to use for task creation.
			   --workflow-manifest, -w      File path for workflow manifest to use for task creation
			   --interval, -i               Interval for the task schedule [ex (simple schedule): 250ms, 1s, 30m (cron schedule): "0 * * * * *"]
			   --start-date                 Start date for the task schedule [defaults to today]
			   --start-time                 Start time for the task schedule [defaults to now]
			   --stop-date                  Stop date for the task schedule [defaults to today]
			   --stop-time                  Start time for the task schedule [defaults to now]
			   --name, -n                   Optional requirement for giving task names
			   --duration, -d               The amount of time to run the task [appends to start or creates a start time before a stop]
			   --no-start                   Do not start task on creation [normally started on creation]

        	* Note: Start and stop date/time are optional.
list         list
start        start <task_id>
stop         stop <task_id>
remove       remove <task_id>
export       export <task_id>
watch        watch <task_id>
enable       enable <task_id>
help, h      Shows a list of commands or help for one command
```
#### plugin
```
$ $SNAP_PATH/bin/snapctl plugin command [command options] [arguments...]
```
```
load		load <plugin path>
				--plugin-asc, -a     The armored detached plugin signature file (.asc)
unload		unload -t <plugin-type> -n <plugin_name> -v <plugin_version>
				--plugin-type, -t            The plugin type
			    --plugin-name, -n            The plugin name
			    --plugin-version, -v '0'     The plugin version
list		list
help, h		Shows a list of commands or help for one command
```
#### metric
```
$ $SNAP_PATH/bin/snapctl metric command [command options] [arguments...]
```
```
list         list
get          get details on a single metric
help, h      Shows a list of commands or help for one command
```

Example Usage
-------------

### Load and unload plugins, create and start a task

In one terminal window, run snapd (log level is set to 1 and signing is turned off for this example):
```
$ $SNAP_PATH/bin/snapd -l 1 -t 0
```

prepare a task manifest file, for example, task.json with following content:
```json
{
    "version": 1,
    "schedule": {
        "type": "simple",
        "interval": "1s"
    },
    "max-failures": 10,
    "workflow": {
        "collect": {
            "metrics": {
                "/intel/mock/foo": {},
                "/intel/mock/bar": {},
                "/intel/mock/*/baz": {}
            },
            "config": {
                "/intel/mock": {
                    "name": "root",
                    "password": "secret"
                }
            },
            "process": [
                {
                    "plugin_name": "passthru",
                    "process": null,
                    "publish": [
                        {
                            "plugin_name": "mock-file",
                            "config": {
                                "file": "/tmp/snap_published_mock_file.log"
                            }
                        }
                    ]
                }
            ]
        }
    }
}
```

prepare a workflow manifest file, for example, workflow.json with the following content:
```json
{
    "collect": {
        "metrics": {
            "/intel/mock/foo": {}
        },
        "config": {
	    "/intel/mock/foo": {
		"password": "testval"
            }
	},
        "process": [],
        "publish": [
            {
                "plugin_name": "mock-file",
                "config": {
                    "file": "/tmp/rest.test"
                }
            }
        ]
    }
}
```

and then:

1. load a collector plugin
2. load a processing plugin
3. load a publishing plugin
4. list the plugins
5. start a task with a task manifest
6. start a task with a workflow manifest
7. list the tasks
8. unload the plugins

```
$ $SNAP_PATH/bin/snapctl plugin load $SNAP_PATH/plugin/snap-plugin-collector-mock1
$ $SNAP_PATH/bin/snapctl plugin load $SNAP_PATH/plugin/snap-plugin-processor-passthru
$ $SNAP_PATH/bin/snapctl plugin load $SNAP_PATH/plugin/snap-plugin-publisher-mock-file
$ $SNAP_PATH/bin/snapctl plugin list
$ $SNAP_PATH/bin/snapctl task create -t mock-file.json
$ $SNAP_PATH/bin/snapctl task create -w workflow.json -i 1s -d 10s
$ $SNAP_PATH/bin/snapctl task list
$ $SNAP_PATH/bin/snapctl plugin unload -t collector -n mock -v <version>
$ $SNAP_PATH/bin/snapctl plugin unload -t processor -n passthru -v <version>
$ $SNAP_PATH/bin/snapctl plugin unload -t publisher -n publisher -v <version>
```
