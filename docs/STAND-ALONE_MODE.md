# Stand-alone mode

Stand-alone mode enables plugin launching on different machine than Snap daemon (`snapteld`).
This feature works for plugins written using one of our snap-plugin-libs ([snap-plugin-lib-go](https://github.com/intelsdi-x/snap-plugin-lib-go),
[snap-plugin-lib-py](https://github.com/intelsdi-x/snap-plugin-lib-py), [snap-plugin-lib-cpp](https://github.com/intelsdi-x/snap-plugin-lib-cpp)).

## Running a plugin in stand-alone mode
To run a plugin in stand-alone mode, you must start it with the `--stand-alone` flag, e.g.:
```
$ ./snap-plugin-collector-psutil --stand-alone
```

A plugin running in stand-alone mode creates a HTTP server for communication with the Snap framework.
By default the plugin listens on port `8182`.

To specify a different listening port, use the `--stand-alone-port` flag, e.g.:
```
$ ./snap-plugin-collector-psutil --stand-alone --stand-alone-port 8183
```
##  Loading a plugin
When loading a plugin in stand-alone mode into `snapteld` you must provide a URL to indicate which
machine the plugin is running (IP address/hostname with port number), e.g.:

```
$ snaptel plugin load http://127.0.0.1:8182
```

or

```
$ snaptel plugin load http://localhost:8182
```

The rest of operations remains exactly the same as is it for plugins running in regular mode.

## Notice

If there is any disruption in the connection between Snap and a stand-alone plugin then the task is disabled and the plugin is unloaded,
see [github issue](https://github.com/intelsdi-x/snap/issues/1697).