# Snap API v2
Snap exposes RESTful APIs that allow performing various actions. All of Snap's API requests return `JSON`-formatted responses, including errors. Any non-2xx HTTP status code may contain an error message. All API URLs listed in this documentation have the endpoint:
> http://localhost:8181

## API Index
1. [Authentication](#authentication)
2. [Plugin API](#plugin-api)
   * [Plugin Response Parameters](#plugin-response-parameters)
   * [Plugin API endpoints and examples](#plugin-api-endpoints-and-examples)
3. [Metric API](#metric-api)
   * [Metric Response Parameters](#metric-response-parameters)
   * [Metric API endpoints and examples](#metric-api-endpoints-and-examples)
4. [Task API](#task-api)
   * [Task API Response Parameters](#task-api-response-parameters)
   * [Task API endpoints and examples](#task-api-endpoints-and-examples)

### Authentication
If Snap framework is started with `--rest-auth` flag, then all requests without authentication info provided will be unauthorized:
```
curl http://localhost:8181/v2/plugins
```
```json
{
  "code": 401,
  "message": "Not authorized. Please specify the same password that used to start snapteld. E.g: [snaptel -p plugin list] or [curl http://localhost:8181/v2/plugins -u snap]"
}
```
It is required to provide auth info with user `snap`:
```
curl -u snap http://localhost:8181/v2/plugins
Enter host password for user 'snap':
```

## Plugin API
Plugin RESTful API provide the functionality to load, unload and retrieve plugin information.

### Plugin Response Parameters
| Parameter        | Description                                           |
|:-----------------|:------------------------------------------------------|
| name             | plugin name                                           |
| version          | plugin version                                        |
| type             | plugin type                                           |
| signed           | bool value to indicate if the plugin is signed or not |
| status           | plugin status                                         |
| loaded_timestamp | time plugin loaded                                    |

### Plugin API endpoints and examples
**GET /v2/plugins**:
List all loaded plugins

_**Example Request**_
```
curl http://localhost:8181/v2/plugins
```
_**Example Response**_
```json
{
  "plugins": [
    {
      "name": "mock",
      "version": 2,
      "type": "collector",
      "signed": false,
      "status": "loaded",
      "loaded_timestamp": 1504080814,
      "href": "http://localhost:8181/v2/plugins/collector/mock/2"
    },
    {
      "name": "mock-file",
      "version": 3,
      "type": "publisher",
      "signed": false,
      "status": "loaded",
      "loaded_timestamp": 1504080829,
      "href": "http://localhost:8181/v2/plugins/publisher/mock-file/3"
    },
    {
      "name": "passthru",
      "version": 1,
      "type": "processor",
      "signed": false,
      "status": "loaded",
      "loaded_timestamp": 1504080843,
      "href": "http://localhost:8181/v2/plugins/processor/passthru/1"
    }
  ]
}
```
**GET /v2/plugins/:type/:name/:version**:
List plugins for the given type, name, and version

_**Example Request**_
```
curl http://localhost:8181/v2/plugins/collector/mock/1
```
_**Example Response**_
```json
{
  "name": "mock",
  "version": 1,
  "type": "collector",
  "signed": false,
  "status": "loaded",
  "loaded_timestamp": 1504019208,
  "href": "http://localhost:8181/v2/plugins/collector/mock/1"
}
```
**POST /v2/plugins**:
Load a plugin

_**Example Request**_
```
curl -X POST -F snap-plugins=@snap-plugin-collector-mock1 http://localhost:8181/v2/plugins
```
_**Example Response**_
```json
{
  "name": "mock",
  "version": 1,
  "type": "collector",
  "signed": false,
  "status": "loaded",
  "loaded_timestamp": 1504078199,
  "href": "http://localhost:8181/v2/plugins/collector/mock/1"
}
```
**DELETE /v2/plugins/:type/:name/:version**:
Unload a plugin for the given type, name, and version

_**Example Request**_
```
curl -X DELETE http://localhost:8181/v2/plugins/collector/mock/1
```
_**Example Response**_

In case of success, response is empty.

**GET /v2/plugins/:type/:name/:version/config**:
Retrieve the global config for the given type, name, and version plugin

_**Example Request**_
```
curl http://localhost:8181/v2/plugins/collector/mock/1/config
```
_**Example Response**_
```json
{
  "foo": 123
}
```
**PUT /v2/plugins/:type/:name/:version/config**:
Set the config for the given type, name, and version plugin

_**Example Request**_
```
curl -X PUT -d '{"bar": "test"}' http://localhost:8181/v2/plugins/collector/mock/1/config
```
_**Example Response**_
```json
{
  "foo": 123,
  "bar": "test"
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

### Metric API endpoints and examples
**GET /v2/metrics**:
List all collected metrics

_**Example Request**_
```
curl http://localhost:8181/v2/metrics
```
_**Example Response**_
```json
{
  "metrics": [
    {
      "last_advertised_timestamp": 1504080814,
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
      "href": "http://localhost:8181/v2/metrics?ns=%2Fintel%2Fmock%2F%2A%2Fbaz&ver=2"
    },
    {
      "last_advertised_timestamp": 1504080814,
      "namespace": "/intel/mock/all/baz",
      "version": 2,
      "dynamic": false,
      "description": "mock description",
      "unit": "mock unit",
      "href": "http://localhost:8181/v2/metrics?ns=%2Fintel%2Fmock%2Fall%2Fbaz&ver=2"
    },
    {
      "last_advertised_timestamp": 1504080814,
      "namespace": "/intel/mock/bar",
      "version": 2,
      "dynamic": false,
      "description": "mock description",
      "unit": "mock unit",
      "href": "http://localhost:8181/v2/metrics?ns=%2Fintel%2Fmock%2Fbar&ver=2"
    },
    {
      "last_advertised_timestamp": 1504080814,
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
      "href": "http://localhost:8181/v2/metrics?ns=%2Fintel%2Fmock%2Ffoo&ver=2"
    }
  ]
}
```
**GET /v2/metrics?ns=:namespace**:
List metrics given metric namespace

_**Example Request**_
```
curl -G -d "ns=/intel/mock/*/baz" http://localhost:8181/v2/metrics
```
_**Example Response**_
```json
{
  "metrics": [
    {
      "last_advertised_timestamp": 1504080814,
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
      "href": "http://localhost:8181/v2/metrics?ns=%2Fintel%2Fmock%2F%2A%2Fbaz&ver=2"
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

## Task API endpoints and examples

**GET /v2/tasks**:
List all scheduled tasks

_**Example Request**_
```
curl http://localhost:8181/v2/tasks
```
_**Example Response**_
```json
{
  "tasks": [
    {
      "id": "bddc84df-03ec-4f62-a6f8-5f91dcd7d044",
      "name": "Task-bddc84df-03ec-4f62-a6f8-5f91dcd7d044",
      "deadline": "5s",
      "creation_timestamp": 1504089709,
      "last_run_timestamp": 1504089728,
      "hit_count": 19,
      "task_state": "Running",
      "href": "http://localhost:8181/v2/tasks/bddc84df-03ec-4f62-a6f8-5f91dcd7d044"
    }
  ]
}
```
**GET /v2/tasks/:id**:
Retrieve a task given the task ID

_**Example Request**_
```
curl http://localhost:8181/v2/tasks/bddc84df-03ec-4f62-a6f8-5f91dcd7d044
```
_**Example Response**_
```json
{
  "id": "bddc84df-03ec-4f62-a6f8-5f91dcd7d044",
  "name": "Task-bddc84df-03ec-4f62-a6f8-5f91dcd7d044",
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
  "creation_timestamp": 1504089709,
  "last_run_timestamp": 1504089778,
  "hit_count": 69,
  "task_state": "Running",
  "href": "http://localhost:8181/v2/tasks/bddc84df-03ec-4f62-a6f8-5f91dcd7d044"
}
```
**GET /v2/tasks/:id/watch**:
Watch a task activity stream given a task ID. Watch is an event stream sent over a long running HTTP connection.

_**Example Request**_
```
curl http://localhost:8181/v2/tasks/83965e64-0b45-4df2-bb8a-bc0cbf1b2538/watch
```
_**Example Response**_
```json
data: {"type":"stream-open","message":"Stream opened"}

data: {"type":"metric-event","message":"","event":[{"namespace":"/intel/mock/bar","data":1084,"timestamp":"2017-08-30T12:44:41.435210915+02:00","tags":{"plugin_running_on":"kdembler-dev"}},{"namespace":"/intel/mock/foo","data":1071,"timestamp":"2017-08-30T12:44:41.43523699+02:00","tags":{"plugin_running_on":"kdembler-dev"}},{"namespace":"/intel/mock/host0/baz","data":1079,"timestamp":"2017-08-30T12:44:41.435242443+02:00","tags":{"plugin_running_on":"kdembler-dev"}},{"namespace":"/intel/mock/host1/baz","data":1088,"timestamp":"2017-08-30T12:44:41.43524445+02:00","tags":{"plugin_running_on":"kdembler-dev"}},{"namespace":"/intel/mock/host2/baz","data":1065,"timestamp":"2017-08-30T12:44:41.435245417+02:00","tags":{"plugin_running_on":"kdembler-dev"}},{"namespace":"/intel/mock/host3/baz","data":1067,"timestamp":"2017-08-30T12:44:41.435247382+02:00","tags":{"plugin_running_on":"kdembler-dev"}},{"namespace":"/intel/mock/host4/baz","data":1087,"timestamp":"2017-08-30T12:44:41.435248198+02:00","tags":{"plugin_running_on":"kdembler-dev"}},{"namespace":"/intel/mock/host5/baz","data":1075,"timestamp":"2017-08-30T12:44:41.435249073+02:00","tags":{"plugin_running_on":"kdembler-dev"}},{"namespace":"/intel/mock/host6/baz","data":1068,"timestamp":"2017-08-30T12:44:41.435249884+02:00","tags":{"plugin_running_on":"kdembler-dev"}},{"namespace":"/intel/mock/host7/baz","data":1087,"timestamp":"2017-08-30T12:44:41.435252507+02:00","tags":{"plugin_running_on":"kdembler-dev"}},{"namespace":"/intel/mock/host8/baz","data":1076,"timestamp":"2017-08-30T12:44:41.435253323+02:00","tags":{"plugin_running_on":"kdembler-dev"}},{"namespace":"/intel/mock/host9/baz","data":1069,"timestamp":"2017-08-30T12:44:41.435254198+02:00","tags":{"plugin_running_on":"kdembler-dev"}},{"namespace":"/intel/mock/all/baz","data":1001,"timestamp":"2017-08-30T12:44:41.435284983+02:00","tags":{"plugin_running_on":"kdembler-dev"}}]}

data: {"type":"metric-event","message":"","event":[{"namespace":"/intel/mock/bar","data":1070,"timestamp":"2017-08-30T12:44:42.435340464+02:00","tags":{"plugin_running_on":"kdembler-dev"}},{"namespace":"/intel/mock/foo","data":1067,"timestamp":"2017-08-30T12:44:42.435382846+02:00","tags":{"plugin_running_on":"kdembler-dev"}},{"namespace":"/intel/mock/host0/baz","data":1081,"timestamp":"2017-08-30T12:44:42.435388658+02:00","tags":{"plugin_running_on":"kdembler-dev"}},{"namespace":"/intel/mock/host1/baz","data":1071,"timestamp":"2017-08-30T12:44:42.435390776+02:00","tags":{"plugin_running_on":"kdembler-dev"}},{"namespace":"/intel/mock/host2/baz","data":1086,"timestamp":"2017-08-30T12:44:42.435391701+02:00","tags":{"plugin_running_on":"kdembler-dev"}},{"namespace":"/intel/mock/host3/baz","data":1076,"timestamp":"2017-08-30T12:44:42.435393799+02:00","tags":{"plugin_running_on":"kdembler-dev"}},{"namespace":"/intel/mock/host4/baz","data":1065,"timestamp":"2017-08-30T12:44:42.435394611+02:00","tags":{"plugin_running_on":"kdembler-dev"}},{"namespace":"/intel/mock/host5/baz","data":1073,"timestamp":"2017-08-30T12:44:42.435395461+02:00","tags":{"plugin_running_on":"kdembler-dev"}},{"namespace":"/intel/mock/host6/baz","data":1068,"timestamp":"2017-08-30T12:44:42.435396279+02:00","tags":{"plugin_running_on":"kdembler-dev"}},{"namespace":"/intel/mock/host7/baz","data":1078,"timestamp":"2017-08-30T12:44:42.435398486+02:00","tags":{"plugin_running_on":"kdembler-dev"}},{"namespace":"/intel/mock/host8/baz","data":1081,"timestamp":"2017-08-30T12:44:42.435399336+02:00","tags":{"plugin_running_on":"kdembler-dev"}},{"namespace":"/intel/mock/host9/baz","data":1075,"timestamp":"2017-08-30T12:44:42.435400141+02:00","tags":{"plugin_running_on":"kdembler-dev"}},{"namespace":"/intel/mock/all/baz","data":1001,"timestamp":"2017-08-30T12:44:42.435425771+02:00","tags":{"plugin_running_on":"kdembler-dev"}}]}
...
```
**POST /v2/tasks**:
Create a task with JSON input, using for example mock-file.json with following content:
```json
{
  "version": 1,
  "schedule": {
    "type": "simple",
    "interval": "1s"
  },
  "max-failures": 10,
  "workflow": {
    "collect": {
      "metrics": {
        "/intel/mock/*/baz": {
        },
        "/intel/mock/bar": {
        },
        "/intel/mock/foo": {
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
          "publish": [
            {
              "plugin_name": "mock-file",
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

_**Example Request**_
```
curl -X POST -d @mock-file.json http://localhost:8181/v2/tasks
```
_**Example Response**_
```json
{
  "id": "5b931ade-d0f9-42dc-bcbd-3d47a5bc1709",
  "name": "Task-5b931ade-d0f9-42dc-bcbd-3d47a5bc1709",
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
  "creation_timestamp": 1504102661,
  "last_run_timestamp": -1,
  "task_state": "Stopped",
  "href": "http://localhost:8181/v2/tasks/5b931ade-d0f9-42dc-bcbd-3d47a5bc1709"
}
```
**PUT /v2/tasks/:id?action=:action**:
Change state of task with given `id`.
Allowed actions are:
- `enable`
- `start`
- `stop`

_**Example Request**_
```
curl -X PUT -G -d action=stop http://localhost:8181/v2/tasks/5b931ade-d0f9-42dc-bcbd-3d47a5bc1709
```
_**Example Response**_

In case of success, response is empty.

**DELETE /v2/tasks/:id**:
Remove stopped task from the scheduled task list given a task ID

_**Example Request**_
```
curl -X DELETE http://localhost:8181/v2/tasks/5b931ade-d0f9-42dc-bcbd-3d47a5bc1709
```
_**Example Response**_

In case of success, response is empty.
