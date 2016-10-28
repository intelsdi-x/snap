## Best practices in plugin development:
### Leverage plugin configurability options
1.  **Compile time configuration** - use `Plugin.PluginMeta` to define plugin's: name, version, type, accepted and returned content types, concurrency level, exclusiveness, secure communication settings and cache TTL. This type of configuration is usually specified in `main()` in which `plugin.Start()` method is called.
2. **Run time configuration**
    - **Global** - This config is useful if configuration data is needed to obtain the list of metrics (for example: user names, paths to tools, etc.). Values from Global config (as defined in config json) are available in `GetMetricTypes()` method.
    - **Task level** - This config is useful when you need to pass configuration per metric or plugin in order to collect the metrics. Use `GetConfigPolicy()` to set configurable items for plugin. Values from Task config are available in `CollectMetrics()` method.

### Use `snap-plugin-utilities` library
The library and guide are available [here](https://github.com/intelsdi-x/snap-plugin-utilities). The library consists of the following helper packages:
* **`config`** - The config package provides helpful methods to retrieve global config items.
* **`logger`** - The logger package wraps the logrus package. (https://github.com/Sirupsen/logrus). It sets logging from a plugin to separate files and adds a caller function name to each message. It's best to use log level defined during `snapd` start.
* **`ns`** - The ns package provides functions to extract namespace from maps, JSON and struct compositions. It is useful for situations when full knowledge of available metrics is not known at the time when `GetMetricTypes()` is called.
* **`pipeline`** - Creates an array of Pipes connected by channels. Each Pipe can perform a single process on data transmitted by channels.
* **`source`** - The source package provides handy ways of dealing with external command output. It can be used for continuous command execution (PCM like), or for single command calls.
* **`stack`** - The stack package provides a simple implementation of a stack.

### Namespace definition
* The `GetMetricTypes()` method returns namespaces for your metrics. For new metrics, it is a good idea to start with your organization (for example: "intel"), then enumerate information starting from most general to most detailed. Examples: `/intel/server/mem/free` or `/intel/linux/iostat/avg-cpu/%idle`.
* Do not use any special character in namespaces other than: "%", "-" and "_".
* When enumerating metric sources, it is always better to have them separated. For example: use `cpu/0` rather than `cpu0`.

### Testability
It is highly recommended to deliver unit and integration tests with your code - use best practices for your language of choice to create testable code (for example in golang consider using dependency injection and interfaces).
