Pulse REST API


Task Schema
===========

This is used in REST, and in a file format for the cli.

```yaml
---
version: 1
task:
  schedule:
    type: simple
    interval: 5s
  deadline: 5s
  config:
    /intel/dummy:
    - key: password
      value: j3rr
  workflow:
    collect:
      metric_types:
      - namespace: /intel/dummy/foo
      - namespace: /intel/dummy/bar
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
    "schedule": {
      "type": "simple",
      "interval": "5s"
    },
    "deadline": "5s",
    "config": {
      "/intel/dummy": [
        {
          "key": "password",
          "value": "j3rr"
        }
      ]
    },
    "workflow": {
      "collect": {
        "metric_types": [
          {
            "namespace": "/intel/dummy/foo"
          },
          {
            "namespace": "/intel/dummy/bar"
          }
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
