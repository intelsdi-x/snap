## Pulse Collector Plugin Structure
---

#### Plugin binary

./main.go

#### Collector Implementation

./collector/collector.go

#### JSON RPC examples (using curl)

If calling a GO based plugin you will want to ensure that the plugin is started in JSON RPC mode.  This is done by setting the plugins meta data field RPCType to plugin.JSONRPC. 

You can start a plugin manually for testing by increasing the ping timeout duration.  The timeout will be reset each time you call into the plugin.

```
./pulse-collector-dummy2 '{"NoDaemon": false, "PingTimeoutDuration": 1000000000000}'
```

###### GetConfigPolicyTree

```
curl -d '{"method": "Collector.GetConfigPolicyTree", "params": [], "id": 1}' http://127.0.0.1:<REPLACE WITH PORT> | python -m "json.tool"
```

###### GetMetricTypes

```
curl -d '{"method": "Collector.GetMetricTypes", "params": [], "id": 1}' http://127.0.0.1:<REPLACE WITH PORT>
```

###### CollectMetrics

```
curl -X POST -H "Content-Type: application/json" -d '{"method": "Collector.CollectMetrics", "params": [[{"namespace": ["intel","dummy", "bar"]},{"namespace": ["intel","dummy","foo"], "config": {"table": {"password": {"Value": "asdf"}}}}]], "id": 1}' http://127.0.0.1:<REPLACE WITH PORT> | python -m "json.tool"
```
