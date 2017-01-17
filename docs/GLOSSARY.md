# Glossary
Snap is simple in scope and it becomes more simple when you know the terminology we use throughout the project. Here they are:

* [Glossary](#glossary)
    * [Config: Global Config](#config-global-config)
    * [Config: Global Options](#config-global-options)
    * [Config: Metric Config](#config-metric-config)
    * [Config: Plugin Config](#config-plugin-config)
    * [Metric Catalog](#metric-catalog)
    * [Metric: Dynamic](#metric-dynamic)
    * [Metric: Namespace](#metric-namespace)
    * [Metric Namespace: Dynamic Element](#metric-dynamic-namespace-element)
    * [Metric Namespace: Static Element](#metric-static-namespace-element)
    * [Plugin](#plugin)
    * [Plugin Status](#plugin-status)
    * [Plugin Type: Collector](#plugin-type-collector)
    * [Plugin Type: Processor](#plugin-type-processor)
    * [Plugin Type: Publisher](#plugin-type-publisher)
    * [Snap](#snap)
    * [Snap Telemetry](#snap-telemetry)
    * [snaptel](#snaptel)
    * [snapteld](#snapteld)
    * [Task](#task)
    * [Task Manifest](#task-manifest)
    * [Tribe](#tribe)
    * [Workflow](#workflow)
    * [Workflow: Distributed](#workflow-distributed)
    * [Workflow Manifest](#workflow-manifest)

### Config: Global Config
* Values loaded at runtime of the daemon ([reference](SNAPTELD_CONFIGURATION.md))

### Config: Global Options
* Values passed as command-line parameters or environmental variables ([reference](SNAPTEL.md#global-options))

### Config: Metric Config
* key/value pairs shared by collector namespace in a Task Manifest ([example](https://github.com/intelsdi-x/snap-plugin-collector-meminfo/blob/master/examples/tasks/task-mem.json#L15))

### Config: Plugin Config
* key/value pairs stored in the `config` block within a Task Manifest ([example](https://github.com/intelsdi-x/snap-plugin-collector-meminfo/blob/master/examples/tasks/task-mem.json#L24))

### Metric Catalog
* List of all available metrics exposed by a running instance of the Snap daemon ([reference](PLUGIN_LIFECYCLE.md#what-happens-when-a-plugin-is-loaded))

### Metric: Dynamic
* A metric is described as dynamic when it includes one or more wildcards in its namespace ([reference](METRICS.md#dynamic-metrics))

### Metric: Namespace
* Namespaces are a series of namespace elements that uniquely identify a metric in Snap ([reference](METRICS.md))

#### Metric: Dynamic Namespace Element
* An element of a metric whose value is set at runtime ([reference](METRICS.md))

#### Metric: Static Namespace Element
* An element of a metric whose value is set at load time ([reference](METRICS.md))

### Plugin
* An independent [binary][binary] that is compatible with Snap (see [Plugin Lifecycle](PLUGIN_LIFECYCLE.md))

### Plugin Status
* An indicator of whether a plugin meets the maintainer's recommendations for best practices (see [Plugin Status](PLUGIN_STATUS.md))

### Plugin Type: Collector
* Gathers data and presents as a dynamically-generated namespaced metric catalog ([reference](PLUGIN_AUTHORING.md#plugin-type))

### Plugin Type: Processor
* Extends or filters collected metrics ([reference](PLUGIN_AUTHORING.md#plugin-type))

### Plugin Type: Publisher
* Persists metrics into a target endpoint ([reference](PLUGIN_AUTHORING.md#plugin-type))

### Snap
* The project name, focused on the Snap daemon and the plugins that power its collection, processing and publishing of telemetry

### Snap Telemetry
* The full name of the Snap project, used mostly for easy searching (like snap-telemetry.io) or hashtag (#SnapTelemetry)

### 'snaptel'
* The command-line interface (CLI) for Snap, released as a [binary][binary]

### 'snapteld'
* The [daemon process](http://www.linfo.org/daemon.html) for Snap, released as a [binary][binary]

### Task
* A job running within Snap, including the API version, schedule and workflow (all documented [here](TASKS.md))

### Task Manifest
* A file that includes the API version, schedule and workflow of a Task in a declarative form ([reference](TASKS.md#task-manifest))

### Tribe
* The clustering feature of Snap, documented [here](TRIBE.md)

### Workflow
* The explicit map of how collectors, processors and publishers are used in Snap ([reference](TASKS.md#the-workflow))

### Workflow: Distributed
* A workflow where one or more steps have a remote target specified ([reference](DISTRIBUTED_WORKFLOW_ARCHITECTURE.md))

### Workflow Manifest
* A file that describes only the workflow of a Task ([example at the bottom](SNAPTEL.md#load-and-unload-plugins-create-and-start-a-task))

[binary]: https://www.quora.com/Whats-the-difference-between-an-installer-source-code-and-a-binary-package-when-installing-software
