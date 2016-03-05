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
The snap daemon/agent (snapd) is a modular application that consists of a control module, a scheduler module, and a REST API. The control module is responsible for loading and unloading plugins, managing loaded plugins, and maintaining an available pool of running plugins for running tasks. The scheduler module is responsible for running the workflows in created tasks per the schedule stated. The REST API provides an interface for loading and unloading plugins, creating and removing tasks, starting and stopping tasks, and listing metrics available for collection.

## Usage
```
$ $SNAP_PATH/bin/snapd [global options] command [command options] [arguments...]
```

### Options
```
--disable-api, -d                            Disable the agent REST API
--api-port, -p '8181'                        API port (Default: 8181)
--log-level, -l '3'                          1-5 (Debug, Info, Warning, Error, Fatal) [$SNAP_LOG_LEVEL]
--log-path, -o                               Path for logs. Empty path logs to stdout. [$SNAP_LOG_PATH]
--max-procs, -c '1'                          Set max cores to use for snap Agent. Default is 1 core. [$GOMAXPROCS]
--auto-discover, -a                          Auto discover paths separated by colons. [$SNAP_AUTODISCOVER_PATH]
--max-running-plugins, -m '3'                The maximum number of instances of a loaded plugin to run [$SNAP_MAX_PLUGINS]
--cache-expiration '500ms'                   The time limit for which a metric cache entry is valid [$SNAP_CACHE_EXPIRATION]
--plugin-trust, -t '1'                       0-2 (Disabled, Enabled, Warning) [$SNAP_TRUST_LEVEL]
--keyring-paths, -k                          Keyring paths for signing verification separated by colons [$SNAP_KEYRING_PATHS]
--rest-cert                                  A path to a certificate to use for HTTPS deployment of snap's REST API
--config                                     A path to a config file
--rest-https                                 start snap's API as https
--rest-key                                   A path to a key file to use for HTTPS deployment of snap's REST API
--rest-auth					                 Enables snap's REST API authentication
--tribe-node-name 'tjerniga-mac01.local'     Name of this node in tribe cluster (default: hostname) [$SNAP_TRIBE_NODE_NAME]
--tribe                                      Enable tribe mode [$SNAP_TRIBE]
--tribe-seed                                 IP (or hostname) and port of a node to join (e.g. 127.0.0.1:6000) [$SNAP_TRIBE_SEED]
--tribe-addr '192.168.10.101'                Addr tribe gossips over to maintain membership [$SNAP_TRIBE_ADDR]
--tribe-port '6000'                          Port tribe gossips over to maintain membership [$SNAP_TRIBE_PORT]
--help, -h                                   show help
--version, -v                                print the version
```

## Examples
### Commands
```
$SNAP_PATH/bin/snapd
$SNAP_PATH/bin/snapd -log-level 4
$SNAP_PATH/bin/snapd -l 1 -t 2 -k <keyringPath>
$SNAP_PATH/bin/snapd -a $SNAP_PATH/plugins/
$SNAP_PATH/bin/snapd --version
```

### Output
```
$ $SNAP_PATH/bin/snapd -l 1 -t 0 --rest-auth 
```
```
INFO[0000] Starting snapd (version: unknown)
INFO[0000] setting GOMAXPROCS to: 1 core(s)
INFO[0000] control started                               _block=start _module=control
INFO[0000] module started                                _module=snapd block=main snap-module=control
INFO[0000] scheduler started                             _block=start-scheduler _module=scheduler
INFO[0000] module started                                _module=snapd block=main snap-module=scheduler
INFO[0000] setting plugin trust level to: disabled
INFO[0000] auto discover path is disabled
INFO[0000] Configuring REST API with HTTPS set to: false  _module=_mgmt-rest
INFO[0000] REST API authentication is enabled
What password do you want to use for authentication?
Password:
INFO[0111] REST API authentication password is set
INFO[0111] Starting REST API on :8181                    _module=_mgmt-rest
INFO[0111] REST API is enabled
INFO[0111] snapd started                                 _module=snapd block=main
INFO[0111] setting log level to: debug
```
## More information
* [REST_API.md](REST_API.md)
* [PLUGIN_SIGNING.md](PLUGIN_SIGNING.md)
* [TRIBE.md](TRIBE.md)
