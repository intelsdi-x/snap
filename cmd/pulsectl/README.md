# pulsectl
A powerful telemetry agent framework

## Usage
```
$PULSE_PATH/bin/pulsectl [global options] command [command options] [arguments...]
```
### Global Options
```
--url, -u 'http://localhost:8181'	sets the URL to use [$PULSE_URL]
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

Start pulse with the REST interface enabled

```
$PULSE_PATH/bin/pulsed
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
$PULSE_PATH/bin/pulsectl plugin load $PULSE_PATH/plugin/pulse-collector-psutil
$PULSE_PATH/bin/pulsectl plugin load $PULSE_PATH/plugin/pulse-collector-dummy1
$PULSE_PATH/bin/pulsectl plugin load $PULSE_PATH/plugin/pulse-processor-passthru
$PULSE_PATH/bin/pulsectl plugin load $PULSE_PATH/plugin/pulse-publisher-influxdb
$PULSE_PATH/bin/pulsectl plugin load $PULSE_PATH/plugin/pulse-publisher-file
$PULSE_PATH/bin/pulsectl plugin list
$PULSE_PATH/bin/pulsectl task create -t $PULSE_PATH/../cmd/pulsectl/sample/psutil-influx.json
$PULSE_PATH/bin/pulsectl task list
$PULSE_PATH/bin/pulsectl plugin unload -n psutil -v 1
```
