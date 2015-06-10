Example usage
-------------

### Load plugins, create and start a task

Start pulse with the REST interface enabled

```$PULSE_PATH/bin/pulse-agent -rest```

1. load a collector plugin
2. load a processing plugin
3. load a publishing plugin
4. create a task
5. start a task
6. list the tasks

```
$PULSE_PATH/bin/pulse-ctl plugin load  $PULSE_PATH/plugin/collector/pulse-collector-psutil
$PULSE_PATH/bin/pulse-ctl plugin load  $PULSE_PATH/plugin/collector/pulse-collector-dummy1
$PULSE_PATH/bin/pulse-ctl plugin load  $PULSE_PATH/plugin/processor/pulse-processor-passthru
$PULSE_PATH/bin/pulse-ctl plugin load  $PULSE_PATH/plugin/publisher/pulse-publisher-influxdb
$PULSE_PATH/bin/pulse-ctl task create $PULSE_PATH/../cmd/pulse-ctl/sample/psutil-influx.json
$PULSE_PATH/bin/pulse-ctl task start 1
$PULSE_PATH/bin/pulse-ctl task list
```

ISSUES:
- I was able to create a task with a wmap that defined a plugin that wasn't loaded.  