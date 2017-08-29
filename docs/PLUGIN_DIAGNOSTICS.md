# Plugin Diagnostics

Plugin diagnostics provides a simple way to verify that a plugin is capable of running without requiring the plugin to be loaded and a task started.
This feature works for collector plugins written using one of our snap-plugin-libs ([snap-plugin-lib-go](https://github.com/intelsdi-x/snap-plugin-lib-go), [snap-plugin-lib-py](https://github.com/intelsdi-x/snap-plugin-lib-py), [snap-plugin-lib-cpp](https://github.com/intelsdi-x/snap-plugin-lib-cpp)).

Plugin diagnostics delivers following information:
- runtime details (plugin name, version, etc.)
- configuration details
- list of metrics exposed by the plugin
- metrics with their values that can be collected right now
- and times required to retrieve this information

## Running plugin diagnostics
It is possible to run plugin diagnostics without any arguments (default values are used), e.g.:
```
$ ./snap-plugin-collector-psutil
```
or with plugin configuration in the form of a JSON:
```
$ ./snap-plugin-collector-psutil --config '{"mount_points":"/dev/sda"}'
```

Example output:

```
$ ./snap-plugin-collector-psutil

Runtime Details:
    PluginName: psutil, Version: 14
    RPC Type: gRPC, RPC Version: 1
    Operating system: linux
    Architecture: amd64
    Go version: go1.7
printRuntimeDetails took 11.878µs

Config Policy:
NAMESPACE 		 KEY 		 TYPE 		 REQUIRED 	 DEFAULT 	 MINIMUM 	 MAXIMUM
intel.psutil.disk 	 mount_points 	 string 	 false 		
printConfigPolicy took 67.654µs

Metric catalog will be updated to include: 
    Namespace: /intel/psutil/load/load1 
    Namespace: /intel/psutil/load/load5 
    Namespace: /intel/psutil/load/load15 
    Namespace: /intel/psutil/cpu/*/softirq 
    Namespace: /intel/psutil/cpu/cpu-total/softirq 
    [...]
printMetricTypes took 504.483µs 

Metrics that can be collected right now are: 
    Namespace: /intel/psutil/load/load1        Type: float64     Value: 1.48 
    Namespace: /intel/psutil/load/load5        Type: float64     Value: 1.68 
    Namespace: /intel/psutil/load/load15       Type: float64     Value: 1.65 
    Namespace: /intel/psutil/cpu/cpu0/softirq  Type: float64     Value: 20.36 
    Namespace: /intel/psutil/cpu/cpu1/softirq  Type: float64     Value: 13.62 
    Namespace: /intel/psutil/cpu/cpu2/softirq  Type: float64     Value: 9.96 
    Namespace: /intel/psutil/cpu/cpu3/softirq  Type: float64     Value: 3.6 
    Namespace: /intel/psutil/cpu/cpu4/softirq  Type: float64     Value: 1.42 
    Namespace: /intel/psutil/cpu/cpu5/softirq  Type: float64     Value: 0.69 
    Namespace: /intel/psutil/cpu/cpu6/softirq  Type: float64     Value: 0.54 
    Namespace: /intel/psutil/cpu/cpu7/softirq  Type: float64     Value: 0.31 
    Namespace: /intel/psutil/cpu/cpu-total/softirq  Type: float64     Value: 50.52
    [...]
printCollectMetrics took 7.470091ms 

showDiagnostics took 8.076025ms
```
