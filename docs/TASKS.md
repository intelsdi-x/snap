Tasks
=====

A task describes the how, what, and when to do for a Snap job.  A task is described in a task _manifest_, which can be either JSON or YAML<sup>1</sup>.

_Skip to the TL;DR example [here](#tldr)_.

The manifest can be divided into two parts: Header and Workflow.

### The Header

```yaml
---
  version: 1
  schedule:
    type: "simple"
    interval: "1s"
  max-failures: 10
```

#### Version
The header contains a version, used to differentiate between versions of the task manifest parser.  Right now, there is only one version: `1`.

#### Schedule

The schedule describes the schedule type and interval for running the task.  The type of a schedule could be a simple "run forever" schedule, which is what we see above as `"simple"` or something more complex.  Snap is designed in a way where custom schedulers can easily be dropped in.  If a custom schedule is used, it may require more key/value pairs in the schedule section of the manifest.  At the time of this writing, Snap has three schedules:
- **simple schedule** which is described above,
- **window schedule** which adds a start and stop time,
- **cron schedule** which supports cron-like entries in ```interval``` field, like in this example (workflow will fire every hour on the half hour):
```
    "version": 1,
    "schedule": {
        "type": "cron",
        "interval" : "0 30 * * * *"
    },
    "max-failures": 10,
```
More on cron expressions can be found here: https://godoc.org/github.com/robfig/cron

#### Max-Failures
By default, snap will disable a task if there is 10 consecutive errors from any plugins within the workflow.  The configuration
can be changed by specifying the number of failures value in the task header.  If the max-failures value is -1, snap will
not disable a task with consecutive failure.  Instead, snap will sleep for 1 second for every 10 consective failures
and retry again.

For more on tasks, visit [`SNAPCTL.md`](SNAPCTL.md).

### The Workflow

```yaml
---
  collect:
    metrics:
      /intel/mock/foo: {}
      /intel/mock/bar: {}
      /intel/mock/*/baz: {}
    config:
      /intel/mock:
        user: "root"
        password: "secret"
    process:
      -
        plugin_name: "passthru"
        publish:
          -
            plugin_name: "file"
            config:
              file: "/tmp/published"
```

The workflow is a [DAG](https://en.wikipedia.org/wiki/Directed_acyclic_graph) which describes the how and what of a task.  It is always rooted by a `collect`, and then contains any number of `process`es and `publish`es.

#### Remote Targets

Process and Publish nodes in the workflow can also target remote snap nodes via the 'target' key. The purpose of this is to allow offloading of resource intensive workflow steps from the node where data collection is occuring. Modifying the example above we have:

```yaml
---
  collect:
    metrics:
      /intel/mock/foo: {}
      /intel/mock/bar: {}
      /intel/mock/*/baz: {}
    config:
      /intel/mock:
        user: "root"
        password: "secret"
    process:
      -
        plugin_name: "passthru"
        target: "127.0.0.1:8082"
        publish:
          -
            plugin_name: "file"
            target: "127.0.0.1:8082"
            config:
              file: "/tmp/published"

```

If a target is specified for a step in the workflow, that step will be executed on the remote instance specified by the ip:port target. Each node in the workflow is evaluated independently so a workflow can have any, all, or none of it's steps being done remotely (if `target` key is omitted, that step defaults to local). The ip and port target are the ip and port that has a running control-grpc server. These can be specified to snapd via the `control-listen-addr` and `control-listen-port` flags. The default is the same ip as the snap rest-api and port 8082.

An example json task that uses remote targets can be found under [examples](https://github.com/intelsdi-x/snap/blob/distributed-workflow/examples/tasks/distributed-mock-file.json). More information about the architecture behind this can be found [here](DISTRIBUTED_WORKFLOW_ARCHITECTURE.md).

#### collect

The collect section describes which metrics to collect. Metrics can be enumerated explicitly via:
 - a concrete _namespace_
 - a wildcard, `*`
 - a tuple, `(m1|m2|m3)`

The tuple begins and ends with brackets and items inside are separeted by vertical bar. It works like logical `or`, so it gives an error only if none of these metrics can be collected.

| Metrics declared in task manifest | Collected metrics                                              |                                             |
|:----------------------------------|:---------------------------------------------------------------|:--------------------------------------------|
| /intel/mock/\*                    | /intel/mock/foo <br/> /intel/mock/bar <br/> /intel/mock/\*/baz |                                             |
| /intel/mock/(foo\                 | bar)                                                           | /intel/mock/foo <br/> /intel/mock/bar <br/> |
| /intel/mock/\*/baz                | /intel/mock/\*/baz                                             |                                             |

The namespaces are keys to another nested object which may contain a specific version of a plugin, e.g.:

```yaml
---
/foo/bar/baz:
  version: 4
```

If a version is not given, Snap will __select__ the latest for you.

The config section describes configuration data for metrics.  Since metric namespaces form a tree, config can be described at a branch, and all leaves of that branch will receive the given config.  For example, say a task is going to collect `/intel/perf/foo`, `/intel/perf/bar`, and `/intel/perf/baz`, all of which require a username and password to collect.  That config could be described like so:

```yaml
---
metrics:
  /intel/perf/foo: {}
  /intel/perf/bar: {}
  /intel/perf/baz: {}
config:
  /intel/perf:
    username: jerr
    password: j3rr
```

Applying the config at `/intel/perf` means that all leaves of `/intel/perf` (`/intel/perf/foo`, `/intel/perf/bar`, and `/intel/perf/baz` in this case) will receive the config.

A collect node can also contain any number of process or publish nodes.  These nodes describe what to do next.

#### process

A process node describes which plugin to use to process data coming from either a collection or another process node.  The config section describes config data which may be needed for the chosen plugin.

A process node may have any number of process or publish nodes.

#### publish

A publish node describes which plugin to use to process data coming from either a collection or a process node.  The config section describes config data which may be needed for the chosen plugin.

A publish node is a [pendant vertex (a leaf)](http://mathworld.wolfram.com/PendantVertex.html).  It may contain no collect, process, or publish nodes.

## TL;DR

Below is a complete example task.

### YAML

```yaml
---
  version: 1
  schedule:
    type: "simple"
    interval: "1s"
  workflow:
    collect:
      metrics:
        /intel/mock/foo: {}
        /intel/mock/bar: {}
        /intel/mock/*/baz: {}
      config:
        /intel/mock:
          user: "root"
          password: "secret"
      process:
        -
          plugin_name: "passthru"
          process: null
          publish:
            -
              plugin_name: "file"
              config:
                file: "/tmp/published"
```

### JSON

```json
{
    "version": 1,
    "schedule": {
        "type": "simple",
        "interval": "1s"
    },
    "workflow": {
        "collect": {
            "metrics": {
                "/intel/mock/foo": {},
                "/intel/mock/bar": {},
                "/intel/mock/*/baz": {}
            },
            "config": {
                "/intel/mock": {
                    "user": "root",
                    "password": "secret"
                }
            },
            "process": [
                {
                    "plugin_name": "passthru",
                    "process": null,
                    "publish": [
                        {
                            "plugin_name": "file",
                            "config": {
                                "file": "/tmp/published"
                            }
                        }
                    ]
                }
            ]
        }
    }
}
```

#### footnotes

1. YAML is only supported via the snapctl CLI.  Only JSON is accepted via the REST API.
2. The wildcard must be supported by the target plugin.
