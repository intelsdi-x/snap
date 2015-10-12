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

# Pulse Examples

This folder contains examples of using Pulse. Each folder contains a README.md which explains the examples.

### Example list

* [influxdb-grafana](influxdb-grafana)
* [tasks](influxdb-grafana)

### Linux pre-setup

Use "linux_prep.sh" to create docker-machine container required by examples.
If you are behind proxy server uncomment and set "ENV http_proxy ..." and "ENV https_proxy ..."
in Dockerfile.

Docker file comes from: https://docs.docker.com/examples/running_ssh_service/
Tested on: Ubuntu 14.04.
Depends on: docker, docker-machine.
