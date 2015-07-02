# pulsed

## Usage
```
$PULSE_PATH/bin/pulsed [global options] [arguments...]
```

### Options
```
--auto-discover, -a 		Autodiscover paths separated by colons.
--log-level, -l			1-5 (Debug, Info, Warning, Error, Fatal)
--log-path, -o			Path for logs. (Default: Logs to stdout)
--max-procs, -c			Set max cores to use for Pulse Agent. (Default: 1 core)
--api-port, -p			Port rest api will listen on. (Default: 8181)
--disable-api, -d		Flag to enable/disable rest api. (Default: true)
--version, -v			Print Pulse version. 
```

## Example Usage
```
$PULSE_PATH/bin/pulsed
$PULSE_PATH/bin/pulsed --log-level 1 
$PULSE_PATH/bin/pulsed --version
```

## Example Output
```
INFO[0000] Starting pulsed (version: alpha)
INFO[0000] started                                       _block=start _module=control
INFO[0000] module started                                _module=pulsed block=main pulse-module=control
INFO[0000] scheduler started                             _block=start-scheduler _module=scheduler
INFO[0000] module started                                _module=pulsed block=main pulse-module=scheduler
INFO[0000] [pulse-rest] listening on :8181
```
