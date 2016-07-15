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

### How to run the example

This example includes configuring and starting influxdb (a time series database) and grafana (a metrics dashboard).

The example:
  - Starts containers for Influxdb and Grafana (using Docker)
  - Starts snapd
  - Gets, loads and starts snap plugins
  - Creates and starts a task that collects from psutil and publishes to InfluxDB
  
Start the example by running `run-psutil.sh`

![run-psutil.sh](http://i.giphy.com/d2Zhwlh8lMZM9nkQ.gif)

### Requirements
- docker
- docker-compose 
- netcat 
- SNAP_PATH env var set to the snap build directory

Note: The script also supports docker-machine but doesn't require it.

### Issues/Warning

- Make sure the time on your docker-machine vm is syncd with the time on your host 

- There is an unresolved issue with the 1.12.0-rc4-beta19 (build: 10258) Docker for Mac Beta that will throw an error ("unexpected EOF") while attempting to publish to the InfluxDB container. When using Mac OS X, it is suggested to use an InfluxDB daemon (`influxd`) or to utilize a virtual machine with a docker-supported Linux distribution 

   

