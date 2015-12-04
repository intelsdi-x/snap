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

# Examples
Example package contains a collection of snap examples and [videos](./videos.md) that demonstrate the snap libraries and usages.

* [configs](./configs) is an example of the global configuration file that powers your plugins
* [influxdb-grafana](./influxdb-grafana) has an example of [snap Influxdb publisher plugin](https://github.com/intelsdi-x/snap-plugin-publisher-influxdb). The example demonstrates
Intel [PCM](https://software.intel.com/en-us/articles/intel-performance-counter-monitor) (Performance Counter Monitor) and [PSUTIL](https://github.com/intelsdi-x/snap-plugin-collector-psutil) (Processes and System Utilization) plugins to publish data into Influxdb and using Grafana to view the results. 
* [riemann](./riemann) has an example of [snap Riemann publisher plugin](https://github.com/intelsdi-x/snap-plugin-publisher-riemann) 
* [tasks](./tasks) has JSON and YAML formatted execution requests for snap tasks