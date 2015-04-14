Task Schema
===========

This is used in REST, and in a file format for the cli.

```yaml
---
version: 1
task:
  deadline: 5s
  config:
  /foo/bar:
    key: value
  workflow:
  collect:
    metric_types:
    - /foo/bar/kernel
    - /foo/bar/uptime
    publish:
    - plugin:
        name: "influx"
        version: 2
    process:
    - plugin:
        name: "averager"
      publish:
      - plugin:
          name: "rabbitmq"
          version: 1
```

```json
{
  "version": 1,
  "task": {
    "deadline": "5s",
    "config": {
      "/foo/bar": {
        "key": "value"
      }
    },
    "workflow": {
      "collect": {
        "metric_types": [
          "/foo/bar/kernel",
          "/foo/bar/uptime"
        ],
        "publish": [
          {
            "plugin": {
              "name": "influx",
              "version": 2
            }
          }
        ],
        "process": [
          {
            "plugin": {
              "name": "averager"
            },
            "publish": [
              {
                "plugin": {
                  "name": "rabbitmq",
                  "version": 1
                }
              }
            ]
          }
        ]
      }
    }
  }
}
```
