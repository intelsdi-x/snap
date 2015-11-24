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

## Plugin Authoring
Pulse itself runs as a master daemon with the core functionality that may load and unload plugin processes via either CLI or HTTP APIs.

A Pulse plugin is a program, or a set of functions or services, written in Go or any language; that may seamlessly integrate with Pulse as executables. 

Communication between Pulse and plugins uses RPC either through HTTP or TCP protocols. HTTP JSON-RPC is good for any language to use due to its nature of JSON representation of data while the native client is only suitable for plugins written in Golang. The data that plugins report to Pulse is in the form of JSON or GOB CODEC.

Before starting writing Pulse plugins, check out the [Plugin Catalog](https://github.com/intelsdi-x/pulse/blob/master/docs/PLUGIN_CATALOG.md) to see if any suit your needs. If not, you need to reference the plugin packages that defines the type of structures and interfaces inside Pulse and then write plugin endpoints to implement the defined interfaces.

## Getting started with the plugin development
Before writing a plugin, you may sketch the type of metrics you want to collect or understand what kind of metrics you like to publish. Basic steps are:

1. Plan your plugin
2. Decide the CODEC for the plugin
3. Download or clone Pulse
4. Setup your development enviroment properly
5. Write the plugin that implements the type of interfaces defined in Pulse
6. Test the plugin
7. Expose the plugin 

### Naming, Files, and Directory    
Pulse supports three type of plugins. They are collectors, processors, and publishers.  The plugin project name should use the following format:  
>pulse-plugin-[type]-[name]

For example:  
>pulse-plugin-collector-hana      
>pulse-plugin-processor-movingaverage    
>pulse-plugin-publisher-influxdb  

Example files and directory structure:  
```
pulse-plugin-[type]-[name]
 |--[name]
  |--[name].go  
  |--[name]_test.go  
  |--[name]_integration_test.go
 |--main.go
 |--main_test.go
```
### Mandatory packages
There are three mandatory packages that every plugin must use. Other than those three packages, you can use other packages as necessary. There is no danger of colliding dependencies as plugins are separated processes. The mandatory packages are:
```
github.com/intelsdi-x/pulse/control/plugin  
github.com/intelsdi-x/pulse/control/plugin/cpolicy  
github.com/intelsdi-x/pulse/core/ctypes  
```
### Writing a collector plugin
A Pulse collector plugin collects telemetry data by communicating with the Pulse daemon. To confine to collector plugin interfaces and metric types defined in Pulse,  a collector plugin must implement the following methods:
```
GetConfigPolicy() (*cpolicy.ConfigPolicy, error)
CollectMetrics([]PluginMetricType) ([]PluginMetricType, error)
GetMetricTypes(PluginConfigType) ([]PluginMetricType, error)
```
### Writing a processor plugin
A Pulse processor plugin allows filtering, aggregation, transformation, etc of collected telemetry data. To complaint with processor plugin interfaces defined in Pulse,  a processor plugin must implement the following methods:
```
GetConfigPolicy() (*cpolicy.ConfigPolicy, error)
Process(contentType string, content []byte, config map[string]ctypes.ConfigValue) (string, []byte, error)
```
### Writing a publisher plugin
A Pulse publisher plugin allows publishing processed telemetry data into a variety of systems, databases, and monitors through Pulse metrics. To compliant with metric types and plugin interfaces defined in Pulse,  a publisher plugin must implement the following methods:
```
GetConfigPolicy() (*cpolicy.ConfigPolicy, error)
Publish(contentType string, content []byte, config map[string]ctypes.ConfigValue) error
```
### Exposing a plugin
Creating the main program to serve the newly written plugin as an external process in main.go. By defining "Plugin.PluginMeta" with plugin specific settings, the newly created plugin may have its setting to override Pulse global settings. Please refer to [a sample](https://github.com/intelsdi-x/pulse/blob/master/plugin/collector/pulse-collector-mock1/main.go) to see how main.go is written. You may browse [Pulse global settings](https://github.com/intelsdi-x/pulse/blob/master/pulse.go#L45-L119).

Building main.go generates a binary executable. You may choose to sign the executable with our [plugin signing](https://github.com/intelsdi-x/pulse/blob/master/pkg/psigning/README.md).

### Localization
All comments and READMEs within the plugin code should be in English.  For different languages, include appropriate translation files within the plugin package for internationalization.

### README
All plugins should have a README with some standard fields:
```
 1. Pulse version requires at least
 2. Pulse version tested up to
 3. Supported platforms
 4. Contributor
 5. License
```
### Encryption
Pulse provides the encryption capability for both HTTP and TCP clients. The communication between the Pulse Daemon and the plugins is encrypted by default. Should you want to disable the encrypted communication, when authoring a plugin, use the `Unsecure` option for your plugin's meta:
```
//Meta returns the metadata for MyPlugin
func Meta() *plugin.PluginMeta {
    return plugin.NewPluginMeta(name, ver, type, ct, ct2, plugin.Unsecure(true))
}
```

## Logging and debugging
Pulse uses [logrus](http://github.com/Sirupsen/logrus) to log. Your plugins can use it, or any standard Go log package. Each plugin has its log file. If no logging directory is specified, logs are in the /tmp directory of the running machine. INFO is the logging level for the release version of plugins. Loggers are excellent resources for debugging. You can also use Go GDB to debug.

## Building and running the tests
While developing a plugin, unit and integration tests need to be performed. Pulse uses [goconvery](http://github.com/smartystreets/goconvey/convey) for unit tests. You are welcome to use it or any other unit test framework. For the integration tests, you have to set up $PULSE_PATH and some necessary direct, or indirect dependencies. Using Docker container for integration tests is an effective testing strategy. Integration tests may define an input workflow. Refer to a sample [integration test input](https://github.com/intelsdi-x/pulse/blob/master/examples/configs/pulse-config-sample.json).

For example, to run a plugin integration test
```
go test -v tag=integration ./… 
```

For more build and test tips, please refer to our [contributing doc](https://github.com/intelsdi-x/pulse/blob/master/CONTRIBUTING.md).

## Distributing plugins
If you think others would find your plugin useful, we encourage you to submit it to our [Plugin Catalog](https://github.com/intelsdi-x/pulse/blob/master/docs/PLUGIN_CATALOG.md) for possible inclusion.

## License
Project Pulse is released under the Apache 2.0 license.

## For more help
Please browse more at our [repo](https://github.com/intelsdi-x/pulse) or contact Pulse [maintainers](https://github.com/intelsdi-x/pulse#maintainers).
