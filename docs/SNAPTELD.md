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

# snapteld
The Snap daemon/agent (snapteld) is a modular application that consists of a control module, a scheduler module, and a REST API. The control module is responsible for loading and unloading plugins, managing loaded plugins, and maintaining an available pool of running plugins for running tasks. The scheduler module is responsible for running the workflows in created tasks per the schedule stated. The REST API provides an interface for loading and unloading plugins, creating and removing tasks, starting and stopping tasks, and listing metrics available for collection.

## Usage
The `snap-telemetry` package installs `snapteld` in `/usr/local/sbin/snapteld`. Either ensure `/usr/local/sbin` is in your path, or use fully qualified filepath to the `snapteld` binary:

```
$ snapteld [global options] command [command options] [arguments...]
```

### Options
```
--log-level value, -l value                  1-5 (Debug, Info, Warning, Error, Fatal; default: 3) [$SNAP_LOG_LEVEL]
--log-path value, -o value                   Path for logs. Empty path logs to stdout. [$SNAP_LOG_PATH]
--log-truncate                               Log file truncating mode. Default is false => append (true => truncate).
--log-colors                                 Log file coloring mode. Default is true => colored (--log-colors=false => no colors).
--max-procs value, -c value                  Set max cores to use for Snap Agent (default: 1) [$GOMAXPROCS]
--config value                               A path to a config file [$SNAP_CONFIG_PATH]
--max-running-plugins value, -m value        The maximum number of instances of a loaded plugin to run (default: 3) [$SNAP_MAX_PLUGINS]
--plugin-load-timeout value                  The maximum number seconds a plugin can take to load (default: 3) [$SNAP_PLUGIN_LOAD_TIMEOUT]
--auto-discover value, -a value              Auto discover paths separated by colons. [$SNAP_AUTODISCOVER_PATH]
--plugin-trust value, -t value               0-2 (Disabled, Enabled, Warning; default: 1) [$SNAP_TRUST_LEVEL]
--keyring-paths value, -k value              Keyring paths for signing verification separated by colons [$SNAP_KEYRING_PATHS]
--cache-expiration value                     The time limit for which a metric cache entry is valid (default: 500ms) [$SNAP_CACHE_EXPIRATION]
--control-listen-port value                  Listen port for control RPC server (default: 8082) [$SNAP_CONTROL_LISTEN_PORT]
--control-listen-addr value                  Listen address for control RPC server [$SNAP_CONTROL_LISTEN_ADDR]
--temp_dir_path value                        Temporary path for loading plugins [$SNAP_TEMP_DIR_PATH]
--tls-cert value                             A path to PEM-encoded certificate for framework to use for securing communication channels to plugins over TLS
--tls-key value                              A path to PEM-encoded private key file for framework to use for securing communication channels to plugins over TLS
--ca-cert-paths                              List of paths (directories/files) to CA certificates for validating plugin certificates in secure TLS communication
--work-manager-queue-size value              Size of the work manager queue (default: 25) [$WORK_MANAGER_QUEUE_SIZE]
--work-manager-pool-size value               Size of the work manager pool (default: 4) [$WORK_MANAGER_POOL_SIZE]
--disable-api, -d                            Disable the agent REST API
--api-addr value, -b value                   API Address[:port] to bind to/listen on. Default: empty string => listen on all interfaces [$SNAP_ADDR]
--api-port value, -p value                   API port (default: 8181) [$SNAP_PORT]
--rest-https                                 start Snap's API as https
--rest-cert value                            A path to a certificate to use for HTTPS deployment of Snap's REST API
--rest-key value                             A path to a key file to use for HTTPS deployment of Snap's REST API
--rest-auth                                  Enables Snap's REST API authentication
--pprof                                      Enables profiling tools
--tribe-node-name value                      Name of this node in tribe cluster (default: hostname) [$SNAP_TRIBE_NODE_NAME]
--tribe                                      Enable tribe mode [$SNAP_TRIBE]
--tribe-seed value                           IP (or hostname) and port of a node to join (e.g. 127.0.0.1:6000) [$SNAP_TRIBE_SEED]
--tribe-addr value                           Addr tribe gossips over to maintain membership [$SNAP_TRIBE_ADDR]
--tribe-port value                           Port tribe gossips over to maintain membership (default: 6000) [$SNAP_TRIBE_PORT]
--help, -h                                   show help
--version, -v                                print the version
```

## Examples

### Commands
```
$ snapteld
$ snapteld --version
$ snapteld --log-level 4
$ snapteld --auto-discover /opt/snap/plugins/
$ snapteld --log-level 1 --plugin-trust 2 --keyring-paths /etc/snap/keyrings
$ snapteld --log-level 1 --tls-cert /etc/snap/cert/snapteld.crt --tls-key /etc/snap/key/snapteld.key
--ca-cert-paths /etc/ssl/certs/sample_organization_CA.crt:/etc/snap/ca/
```

### Debug output
By default, Snap daemon loads the configuration in `/etc/snap/snapteld.conf` and writes logs to `/var/log/snap/snapteld.log`. When debugging Snap issues, instead of a daemon, you can run it as a foreground process to review the logs directly:

* shutdown `snap-telemetry` service and ensure there's no Snap processes running:
    RedHat 6/Ubuntu 14.04:
```
$ sevice snap-telemetry stop
$ pgrep snap
```
    RedHat 7/Ubuntu 16.04:
```
$ systemctl stop snap-telemetry
$ pgrep snap
```

* run Snap with debug log `--log-level 1`, no log file `--log-path ''`, along with any other custom startup options (it will use the snapteld.conf file settings if an option is omitted):
```
$ snapteld --log-level 1 --log-path '' --plugin-trust 0
```

This should result in the following log output:
```
INFO[2017-01-09T12:55:05-08:00] setting log level to: debug
INFO[2017-01-09T12:55:05-08:00] Starting snapteld (version: 1.0.0)
INFO[2017-01-09T12:55:05-08:00] setting GOMAXPROCS to: 1 core(s)
...
DEBU[2017-01-09T12:55:05-08:00] metric manager linked                         _block=set-metric-manager _module=scheduler
INFO[2017-01-09T12:55:05-08:00] Configuring REST API with HTTPS set to: false  _module=_mgmt-rest
INFO[2017-01-09T12:55:05-08:00] REST API is enabled
INFO[2017-01-09T12:55:05-08:00] control started                               _block=start _module=control
INFO[2017-01-09T12:55:05-08:00] auto discover path is disabled                _block=start _module=control
INFO[2017-01-09T12:55:05-08:00] module started                                _module=snapteld block=main snap-module=control
INFO[2017-01-09T12:55:05-08:00] scheduler started                             _block=start-scheduler _module=scheduler
INFO[2017-01-09T12:55:05-08:00] auto discover path is disabled                _block=start-scheduler _module=scheduler
INFO[2017-01-09T12:55:05-08:00] module started                                _module=snapteld block=main snap-module=scheduler
INFO[2017-01-09T12:55:05-08:00] Starting REST API on :8181                    _module=_mgmt-rest
INFO[2017-01-09T12:55:05-08:00] REST started                                  _block=start _module=_mgmt-rest
INFO[2017-01-09T12:55:05-08:00] module started                                _module=snapteld block=main snap-module=REST
INFO[2017-01-09T12:55:05-08:00] setting plugin trust level to: disabled
INFO[2017-01-09T12:55:05-08:00] snapteld started
                                        ss                  ss
                                    odyssyhhyo         oyhysshhoyo
                                 ddddyssyyysssssyyyyyyyssssyyysssyhy+-
                           ssssssooosyhhysssyyyyyyyyyyyyyyyyyyyyyyyyyyyyssyhh+.
                          ssss lhyssssssyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyssydo
                sssssssssshhhhs lsyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyshh.
           ssyyyysssssssssssyhdo syyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyd.
       syyyyyyyyhhyyyyyyyyyyyyhdo syyyyyyyyyyyyydddhyyyyyyyyyyyyyhhhyyyyyyyyyyyyhh
     ssyyyyyyyh  hhyyyyyyyyyyyyhdo syyyyyyyyyyyddyddddhhhhhhhhdddhhddyyyyyyyyyyyydo
     ddyyyyyyh |  hyyyyyyyyyyyyydds syyyyyyyhhdhyyyhhhhhhhhhhyyyyyyhhdhyyyyyyyyyyym+
     dhyyyyyyyhhhhyyyyyyyyyyyyyyhmds shyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyydyyl
     dhddhyyyyyyhdhyyyyyyyyyyyyyydhmo yhyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyhdhsh
     dhyyhysshhdmdhyyyyyyyyyyyyyyhhdh  hhyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyhhdyoohmh ylo
      yy       dmyyyyyyyyyyyyyyyyhhhm  odhyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyhdhsoshdyyhdddy ylo
            odhhyyyyyyyyyyyyyyyyyyhhdy  oyhhhyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyhhmhyhdhhyyyyyhddoyddy ylo
           dddhhyyyyyyyyyyyyyyyhhhyhhdhs ooosydhhhhyyyyyyyyyyyyyyyyyyyyhhhhhyso+oymyyyyyyyyyyyhhddydmmyys
             ohdyyyyyyyyyyyyyyyhdhyyhhhddyoooohhhhhhhhhhhhhhhhhhhhhdhhyysooosyhhhhhyhhhhhhyyyhyyyhhhhddyy
                dyhyyyyyyyyyyyyydhyyyyyhhdddoooooooooooooooooooooooyysyyhddddhhyyyyydmdddddddmmddddhyyy
               dmmmmoNddddddmddhhhhyyyyyhhhdddddddhhhhhhhhhdddddddddhhhyyyyyyyyyyhNmmmooooooooyyy
                     Nhhhhhhhdmmddmhyyyyyyyyyhhhhhhhhhhhhhhhhhhhhyyyyyyyyyyyyyyyhhm
                     NhhhhhhhhhdmyhdyyyyyyyyyyyyhyyyyyyyyyyyyyyyhhhdhyyyyyyyyyyyhhN
                     NhyyyyyyyyyN dmyyyyyyyyyyhdmdhhhhhhhhhhdhhmmmmN NyyyyyyyyyyhhN
                     NhyyyyyyyyyN  Nyyyyyyyyyhhm               NmddmH Nyyyyyyyyyhdm
       .d8888b.      dmomomommmmh  dhhhhhhyyhhmh               NddddmH Nyyyyyyyyhdh
      d88P  Y88b                   dmomomommmmmh                dmomoH dmomomommmmh
      Y88b.
      "Y888b.   88888b.   8888b.  88888b.
         "Y88b. 888 "88b     "88b 888 "88b
           "888 888  888 .d888888 888  888
     Y88b  d88P 888  888 888  888 888 d88P
      "Y8888P"  888  888 "Y888888 88888P"
                                  888
                                  888
                                  888        _module=snapteld block=main
```

## More information
* [SNAPTELD_CONFIGURATION.md](SNAPTELD_CONFIGURATION.md)
* [REST_API_V1.md](REST_API_V1.md)
* [PLUGIN_SIGNING.md](PLUGIN_SIGNING.md)
* [TRIBE.md](TRIBE.md)
* [SECURE_PLUGIN_COMMUNICATION](SECURE_PLUGIN_COMMUNICATION.md)
