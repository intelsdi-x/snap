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
$SNAP_PATH/bin/snapctl [global options] command [command options] [arguments...]
```
### Global Options
```
--url, -u 'http://localhost:8181'	sets the URL to use [$SNAP_URL]
--help, -h							show help
--version, -v						print the version
```
### Commands
```
task
plugin
metric
help, h		Shows a list of commands or help for one command
```
### Command Options
#### task
```
create		create <task file json|yaml>
list		list
start		start <task_id>
stop		stop <task_id>
remove		remove <task_id>
enable		enable <task_id>
help, h		Shows a list of commands or help for one command
```
#### plugin
```
load		load <plugin path>
unload		unload -n <plugin_name> -v <plugin_version>
list		list
help, h		Shows a list of commands or help for one command
```
#### metric
```
list		list
help, h		Shows a list of commands or help for one command
```

Example Usage
-------------

### Load and unload plugins, create and start a task

Start snap with the REST interface enabled

```
$SNAP_PATH/bin/snapd
```

1. load a collector plugin
2. load a processing plugin
3. load a publishing plugin
4. list the plugins
4. create a task
5. start a task
6. list the tasks
7. unload a plugin

```
$SNAP_PATH/bin/snapctl plugin load $SNAP_PATH/plugin/snap-collector-psutil
$SNAP_PATH/bin/snapctl plugin load $SNAP_PATH/plugin/snap-collector-mock1
$SNAP_PATH/bin/snapctl plugin load $SNAP_PATH/plugin/snap-processor-passthru
$SNAP_PATH/bin/snapctl plugin load $SNAP_PATH/plugin/snap-publisher-influxdb
$SNAP_PATH/bin/snapctl plugin load $SNAP_PATH/plugin/snap-publisher-file
$SNAP_PATH/bin/snapctl plugin list
$SNAP_PATH/bin/snapctl task create -t $SNAP_PATH/../cmd/snapctl/sample/psutil-influx.json
$SNAP_PATH/bin/snapctl task list
$SNAP_PATH/bin/snapctl plugin unload -n psutil -v 1
```
