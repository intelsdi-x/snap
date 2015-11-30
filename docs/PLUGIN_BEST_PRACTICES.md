## Best practices in plugins development:
### Leverage plugin configurability options
1.  **Compile time configuration** - use `Plugin.PluginMeta` to define plugins's: name, version, type, accepted and returned content types, concurrency level, exclusiveness, secure communication settings and cache TTL. This type of configuration is usually specified in `main()` in which `plugin.Start()` method is called.
2. **Run time configuration**
    - **Global** - This config is useful if configuration data are needed to obtain list of metrics (for example: user names, paths to tools, etc.). Values from Global cofig (as defined in config json) are available in `GetMetricTypes()` method.
    - **Task level** - This config is useful when you need to pass configuration per metric or plugin in order to collect the metrics. Use `GetConfigPolicy()` to set configurable items for plugin. Values from Task config are available in `CollectMetrics()` method.

### Use `snap-plugin-utilities` library
The library and guide are available [here](https://github.com/intelsdi-x/snap-plugin-utilities). The library consists of following helper packages:
* **`config`** - The config package provides helpful methods to retrieve global config items.
* **`logger`** - The logger package wraps logrus package (https://github.com/Sirupsen/logrus). It sets logging from plugin to separate file. It adds caller function name to each message. It's best to use log level defined during `snapd` start.
* **`ns`** - The ns package provides functions to extract namespace from maps, JSON and struct compositions. It is useful for situations when full knowledge of available metrics is not known at time when `GetMetricTypes()` is called.
* **`pipline`** - Creates array of Pipes connected by channels. Each Pipe can do single processing on data transmitted by channels
* **`source`** - The source package provides handy way of dealing with external command output. It can be used for continuous command execution (PCM like), or for single command calls.
* **`stack`** - The stack package provides simple implementation of stack.

### Namespace definition
* The `GetMetricTypes()` method returns namespaces for your metrics. The good idea is to start with your organization (for example: "intel"), then enumerate information starting from most general to most detailed. Examples: `/intel/server/mem/free` or `/intel/linux/iostat/avg-cpu/%idle`.
* Do not use any special character in namespaces other than: "%", "-" and "_".
* When enumerating sources it's better to have them separated. For example:use `cpu/0` rather than `cpu0`.

### Testability
It's highly recommended to deliver unit and integration tests with your code - use best practices for your language of choice to create testable code (for example in golang consider using dependency injection and interfaces).
