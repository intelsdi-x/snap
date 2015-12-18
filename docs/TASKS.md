Tasks
=====

A task describes the how, what, and when to do for a __snap__ job.  A task is described in a task _manifest_, which can be either JSON or YAML<sup>1</sup>.

_Skip to the TL;DR example [here](#tldr)_.

The manifest can be divided into two parts: Header and Workflow.

### The Header

```yaml
---
  version: 1
  schedule: 
    type: "simple"
    interval: "1s"
```

#### Version
The header contains a version, used to differentiate between versions of the task manifest parser.  Right now, there is only one version: `1`.

#### Schedule

The schedule describes the schedule type and interval for running the task.  The type of a schedule could be a simple "run forever" schedule, which is what we see above as `"simple"` or something more complex.  __snap__ is designed in a way where custom schedulers can easily be dropped in.  If a custom schedule is used, it may require more key/value pairs in the schedule section of the manifest.  At the time of this writing, __snap__ has a simple schedule which is described above, and a window schedule.  The window schedule adds a start and stop time.  For more on tasks, visit [`SNAPCTL.md`](docs/SNAPCTL.md).

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

#### collect

The collect section describes which metrics to collect.  Metrics can be enumerated explicitly via a concrete _namespace_, or a wildcard (`*`) can be used<sup>2</sup>.  The namespaces are keys to another nested object which may contain a specific version of a plugin, e.g.:

```yaml
---
/foo/bar/baz:
  version: 4
```

If a version is not given, __snap__ will __select__ the latest for you.

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
