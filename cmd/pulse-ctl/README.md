Example usage
-------------


```
$PULSE_PATH/bin/pulse-ctl plugin load  $PULSE_PATH/plugin/collector/pulse-collector-dummy1
$PULSE_PATH/bin/pulse-ctl plugin load  $PULSE_PATH/plugin/processor/pulse-processor-passthru
$PULSE_PATH/bin/pulse-ctl plugin load  $PULSE_PATH/plugin/publisher/pulse-publisher-file
$PULSE_PATH/bin/pulse-ctl task create $PULSE_PATH/../cmd/pulse-ctl/sample/task.json
```