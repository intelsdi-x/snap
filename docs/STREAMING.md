# Streaming 


Streaming plugins use grpc streams to allow the plugin to send data immediately or after a certain period of time instead of on an interval governed by Snap.
Streaming by snap enables - 
* Improved performance by enabling event based data flows
* Runtime configuration controlling event throughput
* Buffer configurations which dispatch events after a given duration and/or when a given event count has been reached

Currently there are two plugins that support streaming - [snap relay](https://github.com/intelsdi-x/snap-relay) and [snap-plugin-collector-rand-streaming](https://github.com/intelsdi-x/snap-plugin-lib-go/tree/master/examples/snap-plugin-collector-rand-streaming). 

# Configuration options 
MaxCollectDuration and MaxMetricsBuffer are two configuration options that can be set through streaming task manifest or flags. 
* MaxCollectDuration sets the maximum duration between collections before metrics are sent. It is always greater than 0 and defaults to 10s which means that after 10 seconds if no new metrics are received, the plugin should send whatever data it has in the buffer.
* MaxMetricsBuffer is the maximum number of metrics the plugin is buffering before sending metrics. It defaults to 0 which means the metrics are sent immediately.  

# Streaming task schedule
```
---
  version: 1
  schedule:
    type: "streaming"
  workflow:
    collect:
      metrics:
        /random/integer: {}
      config:
        /random/integer:
          MaxCollectDuration: "6s"
          MaxMetricsBuffer: 600
```

# Streaming configuration flags
Below is an example of the how to run the snap-relay using the configurable flags. 
1. Start the Snap daemon:
* Run
```
$ snapteld -l 1 -t 0
```
The option "-l 1" is for setting the debugging log level and "-t 0" is for disabling plugin signing.

2. Start the relay using flags: 
* Run snap-relay plugin
```
$ go run main.go --stand-alone --stand-alone-port 8182 --log-level 5 --max-collect-duration 10s --max-metrics-buffer 50

Output: Preamble URL: [::]:8182
```

3. Obtain meta information:
* Curl preamble
```
$ curl localhost:8182
{"Meta":{"Type":3,"Name":"plugin-relay","Version":1,"RPCType":3,"RPCVersion":1,"ConcurrencyCount":5,"Exclusive":false,"Unsecure":true,"CacheTTL":0,"RoutingStrategy":0,"CertPath":"","KeyPath":"","TLSEnabled":false,"RootCertPaths":""},"ListenAddress":"127.0.0.1:54541","PprofAddress":"0","Type":3,"State":0,"ErrorMessage":""}
```

4. Start client in snap-relay:
* Run client (from a different terminal) using the listen address obtained from the previous step
```
$ go run client/main.go 127.0.0.1:54541
```

5. Load plugin:
```
$ snaptel plugin load http://localhost:8182
Plugin loaded
Name: plugin-relay
Version: 1
Type: streaming-collector
Signed: false
Loaded Time: Tue, 05 Sep 2017 17:41:42 PDT
```

* List the metric catalog by running:
```
$ snaptel metric list
NAMESPACE 		           VERSIONS
/intel/relay/collectd 	           1
/intel/relay/statsd 	           1
```

6. Create a task manifest for snap-relay (see [exemplary files](https://github.com/intelsdi-x/snap-relay/tree/master/examples/tasks)):
```
---
  version: 1
  schedule:
    type: "streaming"
  workflow:
    collect:
      metrics:
       /intel/relay/collectd: {}
```

7. Create a task:
```
$ snaptel task create -t collectd.yaml
Using task manifest to create task
Task created
ID: c168b992-8aaf-4eec-8e3c-883510962789
Name: Task-c168b992-8aaf-4eec-8e3c-883510962789
State: Running
```

# Metrics exposed by streaming collectors
Below are some of the metrics collected by the streaming plugins currently: 
```
/intel/relay/collectd
/intel/relay/statsd
/random/integer
/random/float
/random/string
```




