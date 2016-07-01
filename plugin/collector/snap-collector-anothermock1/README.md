<!--
http://www.apache.org/licenses/LICENSE-2.0.txt


Copyright 2016 Intel Corporation

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

## snap Collector Plugin Structure
---
Mock plugins are for testing purposes and not meant as examples. 

#### Plugin binary

./main.go

#### Collector Implementation

./collector/collector.go

#### JSON RPC examples (using curl)

If calling a GO based plugin you will want to ensure that the plugin is started in JSON RPC mode.  This is done by setting the plugins meta data field RPCType to plugin.JSONRPC. 

You can start a plugin manually for testing by increasing the ping timeout duration.  The timeout will be reset each time you call into the plugin.

```
./snap-collector-anothermock1 '{"NoDaemon": false, "PingTimeoutDuration": 1000000000000}'
```

###### GetConfigPolicy

```
curl -d '{"method": "Collector.GetConfigPolicy", "params": [], "id": 1}' http://127.0.0.1:<REPLACE WITH PORT> | python -m "json.tool"
```

###### GetMetricTypes

```
curl -d '{"method": "Collector.GetMetricTypes", "params": [], "id": 1}' http://127.0.0.1:<REPLACE WITH PORT>
```

###### CollectMetrics

```
curl -X POST -H "Content-Type: application/json" -d '{"method": "Collector.CollectMetrics", "params": [[{"namespace": ["intel","anothermock", "bar"]},{"namespace": ["intel","anothermock","foo"], "config": {"table": {"password": {"Value": "asdf"}}}}]], "id": 1}' http://127.0.0.1:<REPLACE WITH PORT> | python -m "json.tool"
```
