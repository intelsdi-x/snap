# Notice
This document describes API v1, which is currently supported, but no longer in active development.

In the future, it will be fully replaced by [API v2](REST_API_V2.md) - see details in [issue 1637](https://github.com/intelsdi-x/snap/issues/1637).

**Notice that [Tribe API](#tribe-api) is available only in API v1 at the moment.**

The process of deprecation will start only after API v2 achieves full feature parity with API v1 (adding Tribe API) and will be preceded by an deprecation announcement giving Snap users time to switch to API v2.

# Snap API
Snap exposes a list of RESTful APIs to perform various actions. All of Snap's API requests return `JSON`-formatted responses, including errors. Any non-2xx HTTP status code may contain an error message. All API URLs listed in this documentation have the endpoint:
> http://localhost:8181

## API Response Meta
| Parameter | Description                |
|:----------|:---------------------------|
| code      | HTTP status code           |
| message   | operation response message |
| type      | operation type             |
| version   | API meta version           |

## API Index
1. [Authentication](#authentication)
2. [Plugin API](#plugin-api)  
   * [Plugin Response Parameters](#plugin-response-parameters)
   * [Plugin APIs and Examples](#plugin-apis-and-examples)
3. [Metric API](#metric-api)  
   * [Metric Response Parameters](#metric-response-parameters)  
   * [Metric APIs and Examples](#metric-apis-and-examples)
4. [Task API](#task-api)  
   * [Task API Response Parameters](#task-api-response-parameters)  
   * [Task APIs and Examples](#task-apis-and-examples)
5. [Tribe API](#tribe-api)  
   * [Tribe API Response Parameters](#tribe-api-response-parameters)  
   * [Tribe APIs and Examples](#tribe-apis-and-examples)

### Authentication
Enabled in snapteld
```
curl -L http://localhost:8181/v1/plugins
```
```json
{
  "code": 401,
  "message": "Not authorized. Please specify the same password that used to start snapteld. E.g: [snaptel -p plugin list] or [curl http://localhost:8181/v2/plugins -u snap]"
}

```

```
curl -L http://localhost:8181/v1/plugins -u snap
Enter host password for user 'snap':
```
```json
{
  "meta": {
    "code": 200,
    "message": "Plugin list returned",
    "type": "plugin_list_returned",
    "version": 1
  },
  "body": {}
}
```

## Plugin API
Plugin RESTful APIs provide the functionality to load, unload and retrieve plugin information. You may see plugin APIs along with their request and response attributes as following:

### Plugin Response Parameters
| Parameter        | Description                                           |
|:-----------------|:------------------------------------------------------|
| name             | plugin name                                           |
| version          | plugin version                                        |
| type             | plugin type                                           |
| signed           | bool value to indicate if the plugin is signed or not |
| status           | plugin status                                         |
| loaded_timestamp | time plugin loaded                                    |

### Plugin APIs and Examples
**GET /v1/plugins**:
List all loaded plugins

_**Example Request**_
```
curl -L http://localhost:8181/v1/plugins
```
_**Example Response**_
```json
{
  "meta": {
    "code": 200,
    "message": "Plugin list returned",
    "type": "plugin_list_returned",
    "version": 1
  },
  "body": {
    "loaded_plugins": [
      {
        "name": "mock-file",
        "version": 3,
        "type": "publisher",
        "signed": false,
        "status": "loaded",
        "loaded_timestamp": 1501848167,
        "href": "http://localhost:8181/v1/plugins/publisher/mock-file/3"
      },
      {
        "name": "mock",
        "version": 1,
        "type": "collector",
        "signed": false,
        "status": "loaded",
        "loaded_timestamp": 1501848140,
        "href": "http://localhost:8181/v1/plugins/collector/mock/1"
      },
      {
        "name": "mock",
        "version": 2,
        "type": "collector",
        "signed": false,
        "status": "loaded",
        "loaded_timestamp": 1501848145,
        "href": "http://localhost:8181/v1/plugins/collector/mock/2"
      },
      {
        "name": "passthru",
        "version": 1,
        "type": "processor",
        "signed": false,
        "status": "loaded",
        "loaded_timestamp": 1501848157,
        "href": "http://localhost:8181/v1/plugins/processor/passthru/1"
      }
    ]
}
```
**GET /v1/plugins/:type/:name/:version**:
List plugins for the given type, name, and version

_**Example Request**_
```
curl -L http://localhost:8181/v1/plugins/collector/mock/1
```
_**Example Response**_
```json
{
  "meta": {
    "code": 200,
    "message": "Plugin returned",
    "type": "plugin_returned",
    "version": 1
  },
  "body": {
    "name": "mock",
    "version": 1,
    "type": "collector",
    "signed": false,
    "status": "loaded",
    "loaded_timestamp": 1501848140,
    "href": "http://localhost:8181/v1/plugins/collector/mock/1"
  }
}
```
**POST /v1/plugins**:
Load a plugin

_**Example Request**_
```
curl -X POST -F plugin=@snap-plugin-collector-mock1 http://localhost:8181/v1/plugins
```
_**Example Response**_
```json
{
  "meta": {
    "code": 201,
    "message": "Plugins loaded: mock(collector v1)",
    "type": "plugins_loaded",
    "version": 1
  },
  "body": {
    "loaded_plugins": [
      {
        "name": "mock",
        "version": 1,
        "type": "collector",
        "signed": false,
        "status": "loaded",
        "loaded_timestamp": 1501848303,
        "href": "http://localhost:8181/v1/plugins/collector/mock/1"
      }
    ]
  }
}
```
**DELETE /v1/plugins/:type/:name/:version**:
Unload a plugin for the given type, name, and version

_**Example Request**_
```
curl -X DELETE http://localhost:8181/v1/plugins/collector/mock/1
```
_**Example Response**_
```json
{
  "meta": {
    "code": 200,
    "message": "Plugin successfully unloaded (mockv1)",
    "type": "plugin_unloaded",
    "version": 1
  },
  "body": {
    "name": "mock",
    "version": 1,
    "type": "collector"
  }
}
```
**GET /v1/plugins/:type/:name/:version/config**:
Retrieve the config for the given type, name, and version plugin

_**Example Request**_
```
curl -L http://localhost:8181/v1/plugins/collector/mock/1/config
```
_**Example Response**_
```json
{
  "meta": {
    "code": 200,
    "message": "Plugin config item retrieved",
    "type": "config_plugin_item_returned",
    "version": 1
  },
  "body": {}
}
```
**PUT /v1/plugins/:type/:name/:version/config**:
Set the config for the given type, name, and version plugin

_**Example Request**_
```
curl -L -X PUT http://localhost:8181/v1/plugins/collector/mock/1/config --data '{"password": "xyz"}'
```
_**Example Response**_
```json
{
  "meta": {
    "code": 200,
    "message": "Plugin config item(s) set",
    "type": "config_plugin_item_created",
    "version": 1
  },
  "body": {
    "password": "xyz"
  }
}
```
## Metric API
Snap metric APIs allow you to retrieve all or particular running metric information by invoking different APIs.  

## Metric Response Parameters
| Parameter                 | Description                                       |
|:--------------------------|:--------------------------------------------------|
| last_advertised_timestamp | last time the metric was used                     |
| namespace                 | metric namespace                                  |
| version                   | metric version                                    |
| policy.name               | metric policy name                                |
| policy.type               | policy data type                                  |
| policy.default            | flag to indicate if the policy is default one     |
| policy.required           | bool value to indicate if the policy is mandatory |

### Metric APIs and Examples
**GET /v1/metrics**:
List all collected metrics

_**Example Request**_
```
curl -L http://localhost:8181/v1/metrics
```
_**Example Response**_
```json
{
  "meta": {
    "code": 200,
    "message": "Metrics returned",
    "type": "metrics_returned",
    "version": 1
  },
  "body": [
    {
      "last_advertised_timestamp": 1501848145,
      "namespace": "/intel/mock/*/baz",
      "version": 2,
      "dynamic": true,
      "dynamic_elements": [
        {
          "index": 2,
          "name": "host",
          "description": "name of the host"
        }
      ],
      "description": "mock description",
      "unit": "mock unit",
      "href": "http://localhost:8181/v1/metrics?ns=%2Fintel%2Fmock%2F%2A%2Fbaz&ver=2"
    },
    {
      "last_advertised_timestamp": 1501848145,
      "namespace": "/intel/mock/all/baz",
      "version": 2,
      "dynamic": false,
      "description": "mock description",
      "unit": "mock unit",
      "href": "http://localhost:8181/v1/metrics?ns=%2Fintel%2Fmock%2Fall%2Fbaz&ver=2"
    },
    {
      "last_advertised_timestamp": 1501848145,
      "namespace": "/intel/mock/bar",
      "version": 2,
      "dynamic": false,
      "description": "mock description",
      "unit": "mock unit",
      "href": "http://localhost:8181/v1/metrics?ns=%2Fintel%2Fmock%2Fbar&ver=2"
    },
    {
      "last_advertised_timestamp": 1501848145,
      "namespace": "/intel/mock/foo",
      "version": 2,
      "dynamic": false,
      "description": "mock description",
      "unit": "mock unit",
      "policy": [
        {
          "name": "name",
          "type": "string",
          "default": "bob",
          "required": false
        },
        {
          "name": "password",
          "type": "string",
          "required": true
        }
      ],
      "href": "http://localhost:8181/v1/metrics?ns=%2Fintel%2Fmock%2Ffoo&ver=2"
    }
  ]
}
```
**GET /v1/metrics/:namespace**:
List metrics given metric namespace

_**Example Request**_
```
curl -L http://localhost:8181/v1/metrics/intel/mock/*/baz
```
_**Example Response**_
```json
{
  "meta": {
    "code": 200,
    "message": "Metrics returned",
    "type": "metrics_returned",
    "version": 1
  },
  "body": [
    {
      "last_advertised_timestamp": 1501848145,
      "namespace": "/intel/mock/*/baz",
      "version": 2,
      "dynamic": true,
      "dynamic_elements": [
        {
          "index": 2,
          "name": "host",
          "description": "name of the host"
        }
      ],
      "description": "mock description",
      "unit": "mock unit",
      "href": "http://localhost:8181/v1/metrics?ns=%2Fintel%2Fmock%2F%2A%2Fbaz&ver=2"
    }
  ]
}
```
## Task API
Snap task APIs provide the functionality to create, start, stop, remove, enable, retrieve and watch scheduled tasks.

### Task API Response Parameters
| Parameter                        | Description                             |
|:---------------------------------|:----------------------------------------|
| id                               | task id defined in UUID                 |
| name                             | task name                               |
| deadline                         | task timeout time                       |
| creation_timestamp               | task creation time                      |
| last_run_timestamp               | last running time of a task             |
| hit_count                        | number of times a task succeeded        |
| task_state                       | state of a task                         |
| workflow.collect.metrics         | map of collected metrics                |
| workflow.collect.config          | map of collected metrics configurations |
| workflow.collect.process         | array of processors used in the task    |
| workflow.collect.process.publish | array of publishers used in the task    |

## Task APIs and Examples

**GET /v1/tasks**:
List all scheduled tasks

_**Example Request**_
```
curl -L http://localhost:8181/v1/tasks
```
_**Example Response**_
```json
{
  "meta": {
    "code": 200,
    "message": "Scheduled tasks retrieved",
    "type": "scheduled_task_list_returned",
    "version": 1
  },
  "body": {
    "ScheduledTasks": [
      {
        "id": "83965e64-0b45-4df2-bb8a-bc0cbf1b2538",
        "name": "Task-83965e64-0b45-4df2-bb8a-bc0cbf1b2538",
        "deadline": "5s",
        "creation_timestamp": 1501849812,
        "last_run_timestamp": 1501849877,
        "hit_count": 2,
        "task_state": "Running",
        "href": "http://localhost:8181/v1/tasks/83965e64-0b45-4df2-bb8a-bc0cbf1b2538"
      }
    ]
  }
}
```
**GET /v1/tasks/:id**:
Retrieve a task given the task ID

_**Example Request**_
```
curl -L http://localhost:8181/v1/tasks/83965e64-0b45-4df2-bb8a-bc0cbf1b2538
```
_**Example Response**_
```json
{
  "meta": {
    "code": 200,
    "message": "Scheduled task (83965e64-0b45-4df2-bb8a-bc0cbf1b2538) returned",
    "type": "scheduled_task_returned",
    "version": 1
  },
  "body": {
    "id": "83965e64-0b45-4df2-bb8a-bc0cbf1b2538",
    "name": "Task-83965e64-0b45-4df2-bb8a-bc0cbf1b2538",
    "deadline": "5s",
    "workflow": {
      "collect": {
        "metrics": {
          "/intel/mock/*/baz": {
            "version": 0
          },
          "/intel/mock/bar": {
            "version": 0
          },
          "/intel/mock/foo": {
            "version": 0
          }
        },
        "config": {
          "/intel/mock": {
            "name": "root",
            "password": "secret"
          }
        },
        "process": [
          {
            "plugin_name": "passthru",
            "plugin_version": 0,
            "publish": [
              {
                "plugin_name": "mock-file",
                "plugin_version": 0,
                "config": {
                  "file": "/tmp/published"
                },
                "target": ""
              }
            ],
            "target": ""
          }
        ]
      }
    },
    "schedule": {
      "type": "windowed",
      "interval": "1s"
    },
    "creation_timestamp": 1501849812,
    "last_run_timestamp": -1,
    "task_state": "Stopped",
    "href": "http://localhost:8181/v1/tasks/83965e64-0b45-4df2-bb8a-bc0cbf1b2538"
  }
}
```
**GET /v1/tasks/:id/watch**:
Watch a task activity stream given a task ID. Watch is an event stream sent over a long running HTTP connection.

_**Example Request**_
```
curl -L http://localhost:8181/v1/tasks/83965e64-0b45-4df2-bb8a-bc0cbf1b2538/watch
```
_**Example Response**_
```json
data: {"type":"stream-open","message":"Stream opened"}

data: {"type":"metric-event","message":"","event":[{"namespace":"/intel/mock/foo","data":1089,"timestamp":"2017-08-04T14:31:49.656031232+02:00","tags":{"plugin_running_on":"mkleina-dev"}},{"namespace":"/intel/mock/bar","data":1082,"timestamp":"2017-08-04T14:31:49.656063662+02:00","tags":{"plugin_running_on":"mkleina-dev"}},{"namespace":"/intel/mock/host0/baz","data":1073,"timestamp":"2017-08-04T14:31:49.65606846+02:00","tags":{"plugin_running_on":"mkleina-dev"}},{"namespace":"/intel/mock/host1/baz","data":1078,"timestamp":"2017-08-04T14:31:49.656070654+02:00","tags":{"plugin_running_on":"mkleina-dev"}},{"namespace":"/intel/mock/host2/baz","data":1075,"timestamp":"2017-08-04T14:31:49.656071603+02:00","tags":{"plugin_running_on":"mkleina-dev"}},{"namespace":"/intel/mock/host3/baz","data":1075,"timestamp":"2017-08-04T14:31:49.65607365+02:00","tags":{"plugin_running_on":"mkleina-dev"}},{"namespace":"/intel/mock/host4/baz","data":1085,"timestamp":"2017-08-04T14:31:49.656074448+02:00","tags":{"plugin_running_on":"mkleina-dev"}},{"namespace":"/intel/mock/host5/baz","data":1066,"timestamp":"2017-08-04T14:31:49.656075342+02:00","tags":{"plugin_running_on":"mkleina-dev"}},{"namespace":"/intel/mock/host6/baz","data":1080,"timestamp":"2017-08-04T14:31:49.656076127+02:00","tags":{"plugin_running_on":"mkleina-dev"}},{"namespace":"/intel/mock/host7/baz","data":1081,"timestamp":"2017-08-04T14:31:49.656078495+02:00","tags":{"plugin_running_on":"mkleina-dev"}},{"namespace":"/intel/mock/host8/baz","data":1066,"timestamp":"2017-08-04T14:31:49.656079292+02:00","tags":{"plugin_running_on":"mkleina-dev"}},{"namespace":"/intel/mock/host9/baz","data":1067,"timestamp":"2017-08-04T14:31:49.656080167+02:00","tags":{"plugin_running_on":"mkleina-dev"}},{"namespace":"/intel/mock/all/baz","data":1001,"timestamp":"2017-08-04T14:31:49.656104944+02:00","tags":{"plugin_running_on":"mkleina-dev"}}]}

data: {"type":"metric-event","message":"","event":[{"namespace":"/intel/mock/foo","data":1086,"timestamp":"2017-08-04T14:31:50.656527448+02:00","tags":{"plugin_running_on":"mkleina-dev"}},{"namespace":"/intel/mock/bar","data":1071,"timestamp":"2017-08-04T14:31:50.656552844+02:00","tags":{"plugin_running_on":"mkleina-dev"}},{"namespace":"/intel/mock/host0/baz","data":1083,"timestamp":"2017-08-04T14:31:50.656557614+02:00","tags":{"plugin_running_on":"mkleina-dev"}},{"namespace":"/intel/mock/host1/baz","data":1076,"timestamp":"2017-08-04T14:31:50.656560121+02:00","tags":{"plugin_running_on":"mkleina-dev"}},{"namespace":"/intel/mock/host2/baz","data":1076,"timestamp":"2017-08-04T14:31:50.656561354+02:00","tags":{"plugin_running_on":"mkleina-dev"}},{"namespace":"/intel/mock/host3/baz","data":1085,"timestamp":"2017-08-04T14:31:50.656564678+02:00","tags":{"plugin_running_on":"mkleina-dev"}},{"namespace":"/intel/mock/host4/baz","data":1065,"timestamp":"2017-08-04T14:31:50.656566053+02:00","tags":{"plugin_running_on":"mkleina-dev"}},{"namespace":"/intel/mock/host5/baz","data":1081,"timestamp":"2017-08-04T14:31:50.656567371+02:00","tags":{"plugin_running_on":"mkleina-dev"}},{"namespace":"/intel/mock/host6/baz","data":1068,"timestamp":"2017-08-04T14:31:50.656568729+02:00","tags":{"plugin_running_on":"mkleina-dev"}},{"namespace":"/intel/mock/host7/baz","data":1065,"timestamp":"2017-08-04T14:31:50.656572315+02:00","tags":{"plugin_running_on":"mkleina-dev"}},{"namespace":"/intel/mock/host8/baz","data":1067,"timestamp":"2017-08-04T14:31:50.656573867+02:00","tags":{"plugin_running_on":"mkleina-dev"}},{"namespace":"/intel/mock/host9/baz","data":1076,"timestamp":"2017-08-04T14:31:50.656575325+02:00","tags":{"plugin_running_on":"mkleina-dev"}},{"namespace":"/intel/mock/all/baz","data":1001,"timestamp":"2017-08-04T14:31:50.656601896+02:00","tags":{"plugin_running_on":"mkleina-dev"}}]}

...
```
**POST /v1/tasks**:
Create a task with the JSON input, using for example mock-file.json with following content:
```json
{
  "meta": {
    "code": 201,
    "message": "Scheduled task created (83965e64-0b45-4df2-bb8a-bc0cbf1b2538)",
    "type": "scheduled_task_created",
    "version": 1
  },
  "body": {
    "id": "83965e64-0b45-4df2-bb8a-bc0cbf1b2538",
    "name": "Task-83965e64-0b45-4df2-bb8a-bc0cbf1b2538",
    "deadline": "5s",
    "workflow": {
      "collect": {
        "metrics": {
          "/intel/mock/*/baz": {
            "version": 0
          },
          "/intel/mock/bar": {
            "version": 0
          },
          "/intel/mock/foo": {
            "version": 0
          }
        },
        "config": {
          "/intel/mock": {
            "name": "root",
            "password": "secret"
          }
        },
        "process": [
          {
            "plugin_name": "passthru",
            "plugin_version": 0,
            "publish": [
              {
                "plugin_name": "mock-file",
                "plugin_version": 0,
                "config": {
                  "file": "/tmp/published"
                },
                "target": ""
              }
            ],
            "target": ""
          }
        ]
      }
    },
    "schedule": {
      "type": "windowed",
      "interval": "1s"
    },
    "creation_timestamp": 1501849812,
    "last_run_timestamp": -1,
    "task_state": "Stopped",
    "href": "http://localhost:8181/v1/tasks/83965e64-0b45-4df2-bb8a-bc0cbf1b2538"
  }
}
```

_**Example Request**_
```
curl -L -X POST http://localhost:8181/v1/tasks -d @mock-file.json --header "Content-Type: application/json"
```
_**Example Response**_
```json
{
  "meta": {
    "code": 201,
    "message": "Scheduled task created (4097d262-6ef3-4f69-b749-35252b2f401a)",
    "type": "scheduled_task_created",
    "version": 1
  },
  "body": {
    "id": "4097d262-6ef3-4f69-b749-35252b2f401a",
    "name": "Task-4097d262-6ef3-4f69-b749-35252b2f401a",
    "deadline": "5s",
    "workflow": {
      "collect": {
        "metrics": {
          "/intel/mock/*": {
            "version": 0
          }
        },
        "config": {
          "/intel/mock": {
            "password": "xyz"
          }
        },
        "publish": [
          {
            "plugin_name": "mock-file",
            "plugin_version": 0,
            "config": {
              "file": "/tmp/published_mock.log"
            },
            "target": ""
          }
        ]
      }
    },
    "schedule": {
      "type": "windowed",
      "interval": "1s"
    },
    "creation_timestamp": 1501849488,
    "last_run_timestamp": -1,
    "task_state": "Stopped",
    "href": "http://localhost:8181/v1/tasks/4097d262-6ef3-4f69-b749-35252b2f401a"
  }
}
```
**PUT /v1/tasks/:id/start**:
Start a task given a task ID

_**Example Request**_
```
curl -X PUT http://localhost:8181/v1/tasks/83965e64-0b45-4df2-bb8a-bc0cbf1b2538/start
```
_**Example Response**_
```json
{
  "meta": {
    "code": 200,
    "message": "Scheduled task (83965e64-0b45-4df2-bb8a-bc0cbf1b2538) started",
    "type": "scheduled_task_started",
    "version": 1
  },
  "body": {
    "id": "83965e64-0b45-4df2-bb8a-bc0cbf1b2538"
  }
}
```
**PUT /v1/tasks/:id/stop**:
Stop a running task given a task ID

_**Example Request**_
```
curl -X PUT http://localhost:8181/v1/tasks/83965e64-0b45-4df2-bb8a-bc0cbf1b2538/stop
```
_**Example Response**_
```json
{
  "meta": {
    "code": 200,
    "message": "Scheduled task (83965e64-0b45-4df2-bb8a-bc0cbf1b2538) stopped",
    "type": "scheduled_task_stopped",
    "version": 1
  },
  "body": {
    "id": "83965e64-0b45-4df2-bb8a-bc0cbf1b2538"
  }
}
```
**DELETE /v1/tasks/:id**:
Remove a task from the scheduled task list given a task ID

_**Example Request**_
```
curl -X DELETE http://localhost:8181/v1/tasks/83965e64-0b45-4df2-bb8a-bc0cbf1b2538
```
_**Example Response**_
```json
{
  "meta": {
    "code": 200,
    "message": "Scheduled task (83965e64-0b45-4df2-bb8a-bc0cbf1b2538) removed",
    "type": "scheduled_task_removed",
    "version": 1
  },
  "body": {
    "id": "83965e64-0b45-4df2-bb8a-bc0cbf1b2538"
  }
}
```
**PUT /v1/tasks/:id/enable**:
Enable a disabled task given a task ID

_**Example Request**_
```
curl -X PUT http://localhost:8181/v1/tasks/84fd498b-9232-40b7-81bd-ac7e86b1f252/enable
```
_**Example Response**_
```json
{
  "meta": {
    "code": 500,
    "message": "Task must be disabled",
    "type": "error",
    "version": 1
  },
  "body": {
    "message": "Task must be disabled",
    "fields": {}
  }
}
```
## Tribe API
Snap tribe APIs provide the functionality for managing tribe agreements and for tribe members to join or leave tribe contracts.

### Tribe API Response Parameters
| Parameter                             | Description                      |
|:--------------------------------------|:---------------------------------|
| agreements.[agreement].name           | agreement name                   |
| agreements.[agreement].plug_agreement | plugins loaded for the agreement |
| agreements.[agreement].task_agreement | agreement scheduled tasks        |
| agreements.members                    | map of tribe members             |
| agreements.members.[member].tags      | map of node properties           |
| agreements.members.[member].name      | node name                        |

### Tribe APIs and Examples
**GET /v1/tribe/agreements**:
List all tribe agreements

_**Example Request**_
```
curl -L http://localhost:8181/v1/tribe/agreements
curl -L http://localhost:8182/v1/tribe/agreements
curl -L http://localhost:8183/v1/tribe/agreements
```
_**Example Response**_
```json
{
  "meta": {
    "code": 200,
    "message": "Tribe agreements retrieved",
    "type": "tribe_agreement_list_returned",
    "version": 1
  },
  "body": {
    "agreements": {
      "three-nodes": {
        "name": "three-nodes",
        "plugin_agreement": {},
        "task_agreement": {},
        "members": {
          "snap-1": {
            "tags": {
              "host": "172.19.0.2",
              "rest_api_port": "8181",
              "rest_insecure": "true",
              "rest_proto": "http"
            },
            "name": "snap-1"
          },
          "snap-2": {
            "tags": {
              "host": "172.19.0.3",
              "rest_api_port": "8181",
              "rest_insecure": "true",
              "rest_proto": "http"
            },
            "name": "snap-2"
          },
          "snap-3": {
            "tags": {
              "host": "172.19.0.4",
              "rest_api_port": "8181",
              "rest_insecure": "true",
              "rest_proto": "http"
            },
            "name": "snap-3"
          }
        }
      },
      "single-node": {
        "name": "single-node",
        "plugin_agreement": {},
        "task_agreement": {},
        "members": {
          "snap-4": {
            "tags": {
              "host": "172.19.0.5",
              "rest_api_port": "8181",
              "rest_insecure": "true",
              "rest_proto": "http"
            },
            "name": "snap-4"
          }
        }
      }
    }
  }
}
```
**POST /v1/tribe/agreements**:
Create a new tribe agreement

_**Example Request**_
```
curl -L -X POST http://localhost:8182/v1/tribe/agreements -d '{"name":"cold-agreement"}'
```
_**Example Response**_
```json
{
  "meta": {
    "code": 200,
    "message": "Tribe agreement created",
    "type": "tribe_agreement_created",
    "version": 1
  },
  "body": {
    "agreements": {
      "cold-agreement": {
        "name": "cold-agreement",
        "plugin_agreement": {},
        "task_agreement": {}
      },
      "single-node": {
        "name": "single-node",
        "plugin_agreement": {},
        "task_agreement": {},
        "members": {
          "snap-4": {
            "tags": {
              "host": "172.19.0.5",
              "rest_api_port": "8181",
              "rest_insecure": "true",
              "rest_proto": "http"
            },
            "name": "snap-4"
          }
        }
      },
      "three-nodes": {
        "name": "three-nodes",
        "plugin_agreement": {},
        "task_agreement": {},
        "members": {
          "snap-1": {
            "tags": {
              "host": "172.19.0.2",
              "rest_api_port": "8181",
              "rest_insecure": "true",
              "rest_proto": "http"
            },
            "name": "snap-1"
          },
          "snap-2": {
            "tags": {
              "host": "172.19.0.3",
              "rest_api_port": "8181",
              "rest_insecure": "true",
              "rest_proto": "http"
            },
            "name": "snap-2"
          },
          "snap-3": {
            "tags": {
              "host": "172.19.0.4",
              "rest_api_port": "8181",
              "rest_insecure": "true",
              "rest_proto": "http"
            },
            "name": "snap-3"
          }
        }
      }
    }
  }
}
```
**GET /v1/tribe/agreements/:name**:
Retrieve a tribe agreement given the agreement name

_**Example Request**_
```
curl -L http://localhost:8183/v1/tribe/agreements/three-nodes
```
_**Example Response**_
```json
{
  "meta": {
    "code": 200,
    "message": "Tribe agreement returned",
    "type": "tribe_agreement_returned",
    "version": 1
  },
  "body": {
    "agreement": {
      "name": "three-nodes",
      "plugin_agreement": {},
      "task_agreement": {},
      "members": {
        "snap-1": {
          "tags": {
            "host": "172.19.0.2",
            "rest_api_port": "8181",
            "rest_insecure": "true",
            "rest_proto": "http"
          },
          "name": "snap-1"
        },
        "snap-2": {
          "tags": {
            "host": "172.19.0.3",
            "rest_api_port": "8181",
            "rest_insecure": "true",
            "rest_proto": "http"
          },
          "name": "snap-2"
        },
        "snap-3": {
          "tags": {
            "host": "172.19.0.4",
            "rest_api_port": "8181",
            "rest_insecure": "true",
            "rest_proto": "http"
          },
          "name": "snap-3"
        }
      }
    }
  }
}
```
**DELETE /v1/tribe/agreements/:name**:
Remove an agreement given the agreement name

_**Example Request**_
```
curl -L -X DELETE http://localhost:8183/v1/tribe/agreements/three-nodes
```
_**Example Response**_
```json
{
  "meta": {
    "code": 200,
    "message": "Tribe agreement deleted",
    "type": "tribe_agreement_deleted",
    "version": 1
  },
  "body": {
    "agreements": {
      "cold-agreement": {
        "name": "cold-agreement",
        "plugin_agreement": {},
        "task_agreement": {}
      },
      "single-node": {
        "name": "single-node",
        "plugin_agreement": {},
        "task_agreement": {},
        "members": {
          "snap-4": {
            "tags": {
              "host": "172.19.0.5",
              "rest_api_port": "8181",
              "rest_insecure": "true",
              "rest_proto": "http"
            },
            "name": "snap-4"
          }
        }
      }
    }
  }
}
```
**PUT /v1/tribe/agreements/:name/join**:
Join a member node into an agreement given the agreement name

_**Example Request**_
```
curl -L -X PUT http://localhost:8183/v1/tribe/agreements/cold-agreement/join -d '{"member_name": "snap-1"}'
```
_**Example Response**_
```json
{
  "meta": {
    "code": 200,
    "message": "Tribe agreement joined",
    "type": "tribe_agreement_joined",
    "version": 1
  },
  "body": {
    "agreement": {
      "name": "cold-agreement",
      "plugin_agreement": {},
      "task_agreement": {},
      "members": {
        "snap-1": {
          "tags": {
            "host": "172.19.0.2",
            "rest_api_port": "8181",
            "rest_insecure": "true",
            "rest_proto": "http"
          },
          "name": "snap-1"
        }
      }
    }
  }
}
```
**DELETE /v1/tribe/agreements/:name/leave**:
Remove a member node from an agreement given the agreement name

_**Example Request**_
```
curl -L -X DELETE http://localhost:8183/v1/tribe/agreements/cold-agreement/leave -d '{"member_name": "snap-1"}'
```
_**Example Response**_
```json
{
  "meta": {
    "code": 200,
    "message": "Tribe agreement left",
    "type": "tribe_agreement_left",
    "version": 1
  },
  "body": {
    "agreement": {
      "name": "cold-agreement",
      "plugin_agreement": {},
      "task_agreement": {}
    }
  }
}
```
**GET /v1/tribe/members**:
List all tribe members

_**Example Request**_
```
curl -L http://localhost:8183/v1/tribe/members
```
_**Example Response**_
```json
{
  "meta": {
    "code": 200,
    "message": "Tribe members retrieved",
    "type": "tribe_member_list_returned",
    "version": 1
  },
  "body": {
    "members": [
      "snap-2",
      "snap-3",
      "snap-4",
      "snap-1"
    ]
  }
}
```
**GET /v1/tribe/member/:name**:
List tribe member information given the node name

_**Example Request**_
```
curl -L http://localhost:8183/v1/tribe/member/snap-1
```
_**Example Response**_
```json
{
  "meta": {
    "code": 200,
    "message": "Tribe member details retrieved",
    "type": "tribe_member_details_returned",
    "version": 1
  },
  "body": {
    "name": "snap-1",
    "plugin_agreement": "",
    "tags": {
      "host": "172.19.0.2",
      "rest_api_port": "8181",
      "rest_insecure": "true",
      "rest_proto": "http"
    },
    "task_agreements": null
  }
}
```
