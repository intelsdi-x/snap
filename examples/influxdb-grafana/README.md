<!--
http://www.apache.org/licenses/LICENSE-2.0.txt


Copyright 2015 Intel Coporation

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

# docker-compose based examples

These examples all include configuring and starting influxdb (a time series database) and grafana (a metrics dashboard).

You can find demo videos [here](../videos.md).

### How to run the example

These examples all include configuring and starting influxdb (a time series database) and grafana (a metrics dashboard).

The following scripts can be used to start the demos

- run-pcm.sh *\<docker-machine name\>* 
  - starts pulsed
  - creates and starts a task that collects from both the pcm and psutil plugins
- run-psutil.sh *\<docker-machine\>*
  - starts pulsed
  - creates and starts a task that collects from psutil plugin
- ./run.sh *\<docker-machine\>*   
  - configured influxdb and grafana only 

### Requirements
- docker-machine 
    + installed and configured

- docker-compose
    + installed

- PCM configured on the host

### Issues/Warning

- Make sure the time on your docker-machine vm is syncd with the time on your host 


   

