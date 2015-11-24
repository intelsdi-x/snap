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

# snapd

## Usage
```
$SNAP_PATH/bin/snapd [global options] [arguments...]
```

### Options
```
--auto-discover, -a 		Autodiscover paths separated by colons.
--log-level, -l			1-5 (Debug, Info, Warning, Error, Fatal)
--log-path, -o			Path for logs. (Default: Logs to stdout)
--max-procs, -c			Set max cores to use for snap Agent. (Default: 1 core)
--api-port, -p			Port rest api will listen on. (Default: 8181)
--disable-api, -d		Flag to enable/disable rest api. (Default: true)
--version, -v			Print snap version. 
```

## Example Usage
```
$SNAP_PATH/bin/snapd
$SNAP_PATH/bin/snapd --log-level 1 
$SNAP_PATH/bin/snapd --version
```

## Example Output
```
INFO[0000] Starting snapd (version: alpha)
INFO[0000] started                                       _block=start _module=control
INFO[0000] module started                                _module=snapd block=main snap-module=control
INFO[0000] scheduler started                             _block=start-scheduler _module=scheduler
INFO[0000] module started                                _module=snapd block=main snap-module=scheduler
INFO[0000] [snap-rest] listening on :8181
```
