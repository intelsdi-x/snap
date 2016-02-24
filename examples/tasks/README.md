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

A task describes the how, what, and when to do for a __snap__ job.  A task is described in a task _manifest_, which can be either JSON or YAML. For more information, see the [documentation for tasks](/docs/TASKS.md).

# Examples in this folder

- **mock-file.json/yaml**: a simple example of task structure in both JSON and YAML format.
  - schedule
    - interval (1s)
  - collector
    - mock
  - processor
    - passthru
  - publisher
    - file

- **psutil-file.yaml**: another simple example of collecting statistics and publishing them to a file. This file includes in-line comments to help get oriented with task structure.
  - schedule
    - interval (1s)
  - collector
    - psutil
  - publisher
    - file

- **ceph-file.json**: collect numerous statistics around Ceph, a storage system for OpenStack.
  - schedule
    - interval (1s)
  - collector
    - Ceph
  - publisher
    - file

- **psutil-influx.json**: a more complex example that collects information from psutil and publishes to an instance of InfluxDB running locally. See [influxdb-grafana](../influxdb-grafana/) for other files to get this running.
  - schedule
    - interval (1s)
  - collector
    - psutil
  - publisher
    - influxdb
