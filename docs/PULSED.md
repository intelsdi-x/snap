# pulsed

## Usage
```
$PULSE_PATH/bin/pulsed [global options] [arguments...]
```

### Options
```
-autodiscover 		Autodiscover paths separated by colons.
-log-level			1-5 (Debug, Info, Warning, Error, Fatal)
-log-path			Path for logs. (Default: Logs to stdout)
-max-procs			Set max cores to use for Pulse Agent. (Default: 1 core)
-port				Port rest api will listen on. (Default: 8181)
-rest				Flag to enable/disable rest api. (Default: true)
-version			Print Pulse version. 
```

## Example Usage
```
go run pulse.go
go run pulse.go -log-level 1 
go run pulse.go -version
```

## Example Output
```
INFO[0000] pulse agent starting                          _module=pulse-agent block=main
INFO[0000] started                                       _block=start _module=control
INFO[0000] module started                                _module=pulse-agent block=main pulse-module=control
INFO[0000] scheduler started                             _block=start-scheduler _module=scheduler
INFO[0000] module started                                _module=pulse-agent block=main pulse-module=scheduler
INFO[0000] [pulse-rest] listening on :8181
```