# Distributed Workflow

A distributed workflow is a workflow where one or more steps have a remote target specified. An example of this is:

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

## Architecture

Distributed workflow is accomplished by allowing remote targets to be specified as part of a task workflow. This is done by having a gRPC server running that can handle actions needed by the scheduler to run a task. These are defined in the [managesMetrics](https://github.com/intelsdi-x/snap/blob/master/scheduler/scheduler.go) interface defined in scheduler/scheduler.go. This interface is implemented by both pluginControl in control/control.go and  ControlProxy in grpc/controlproxy/controlproxy.go. This allows the scheduler to not know/care where a step in the workflow is running. On task creation, the workflow is walked and the appropriate type is selected or created for each step in the workflow.

## Performance considerations

The main performance penalty for using remote targets is that data is now sent over the network instead of locally. This is minimized since Snap will only make remote calls for steps in the workflow that specify a remote target.
