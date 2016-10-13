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
Enabled in snapd
```
curl -L http://localhost:8181/v1/plugins
```
```
Not Authorized
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
        "name": "mock",
        "version": 1,
        "type": "collector",
        "signed": false,
        "status": "loaded",
        "loaded_timestamp": 1447977606
      },
      {
        "name": "mock",
        "version": 2,
        "type": "collector",
        "signed": false,
        "status": "loaded",
        "loaded_timestamp": 1447977606
      },
      {
        "name": "passthru",
        "version": 1,
        "type": "processor",
        "signed": false,
        "status": "loaded",
        "loaded_timestamp": 1447977606
      },
      {
        "name": "file",
        "version": 3,
        "type": "publisher",
        "signed": false,
        "status": "loaded",
        "loaded_timestamp": 1447977607
      }
    ]
  }
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
    "loaded_timestamp": 1447977606
  }
}
```
**POST /v1/plugins**:
Load a plugin

_**Example Request**_
```
curl -X POST -F plugin=@build/plugin/snap-collector-mock http://localhost:8181/v1/plugins
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
        "loaded_timestamp": 1448058077
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
    "message": "Metric",
    "type": "metrics_returned",
    "version": 1
  },
  "body": [
    {
      "last_advertised_timestamp": 1447977606,
      "namespace": "/intel/mock/*/baz",
      "version": 1
    },
    {
      "last_advertised_timestamp": 1447977606,
      "namespace": "/intel/mock/*/baz",
      "version": 2
    },
    {
      "last_advertised_timestamp": 1447977606,
      "namespace": "/intel/mock/bar",
      "version": 1
    },
    {
      "last_advertised_timestamp": 1447977606,
      "namespace": "/intel/mock/bar",
      "version": 2
    },
    {
      "last_advertised_timestamp": 1447977606,
      "namespace": "/intel/mock/foo",
      "version": 1,
      "policy": [
        {
          "name": "password",
          "type": "string",
          "required": true
        },
        {
          "name": "name",
          "type": "string",
          "default": "bob",
          "required": false
        }
      ]
    },
    {
      "last_advertised_timestamp": 1447977606,
      "namespace": "/intel/mock/foo",
      "version": 2,
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
      ]
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
  "meta": {
    "code": 200,
    "message": "Metric returned",
    "type": "metric_returned",
    "version": 1
  },
  "body": {
    "Metric": {
      "last_advertised_timestamp": 1447977606,
      "namespace": "/intel/mock/*/baz",
      "version": 2
    }
  }
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
| hit_count                        | number of times a task ran              |
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
        "id": "f573affa-9326-44a8-a64c-7a0d803d5121",
        "name": "Task-f573affa-9326-44a8-a64c-7a0d803d5121",
        "deadline": "5s",
        "creation_timestamp": 1448004968,
        "last_run_timestamp": 1448004989,
        "hit_count": 20,
        "task_state": "Running"
      }
    ]
  }
}
```
**GET /v1/tasks/:id**:
Retrieve a task given the task ID

_**Example Request**_
```
curl -L http://localhost:8181/v1/tasks/36cd2bbf-b9ab-495a-b8ab-d9f87fa9b88e
```
_**Example Response**_
```json
{
  "meta": {
    "code": 200,
    "message": "Scheduled task (36cd2bbf-b9ab-495a-b8ab-d9f87fa9b88e) returned",
    "type": "scheduled_task_returned",
    "version": 1
  },
  "body": {
    "id": "36cd2bbf-b9ab-495a-b8ab-d9f87fa9b88e",
    "name": "Task-36cd2bbf-b9ab-495a-b8ab-d9f87fa9b88e",
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
            "password": "secret",
            "user": "root"
          }
        },
        "process": [
          {
            "plugin_name": "passthru",
            "plugin_version": 0,
            "publish": [
              {
                "plugin_name": "file",
                "plugin_version": 0,
                "config": {
                  "file": "/tmp/published"
                }
              }
            ]
          }
        ]
      }
    },
    "schedule": {
      "type": "simple",
      "interval": "1s"
    },
    "creation_timestamp": 1448315384,
    "last_run_timestamp": 1448318130,
    "hit_count": 2743,
    "task_state": "Running"
  }
}
```
**GET /v1/tasks/:id/watch**:
Watch a task activity stream given a task ID. Watch is an event stream sent over a long running HTTP connection.

_**Example Request**_
```
curl -L http://localhost:8181/v1/tasks/f573affa-9326-44a8-a64c-7a0d803d5121/watch
```
_**Example Response**_
```json
{"type":"stream-open","message":"Stream opened"}
{"type":"metric-event","message":"","event":[{"namespace":"/intel/mock/host0/baz","data":77,"source":"egu-mac01.lan","timestamp":"2015-11-19T23:45:41.075611868-08:00"},{"namespace":"/intel/mock/host1/baz","data":68,"source":"egu-mac01.lan","timestamp":"2015-11-19T23:45:41.075613646-08:00"},{"namespace":"/intel/mock/host2/baz","data":65,"source":"egu-mac01.lan","timestamp":"2015-11-19T23:45:41.075615188-08:00"},{"namespace":"/intel/mock/host3/baz","data":75,"source":"egu-mac01.lan","timestamp":"2015-11-19T23:45:41.075616491-08:00"},{"namespace":"/intel/mock/host4/baz","data":76,"source":"egu-mac01.lan","timestamp":"2015-11-19T23:45:41.075618022-08:00"},{"namespace":"/intel/mock/host5/baz","data":86,"source":"egu-mac01.lan","timestamp":"2015-11-19T23:45:41.075619501-08:00"},{"namespace":"/intel/mock/host6/baz","data":82,"source":"egu-mac01.lan","timestamp":"2015-11-19T23:45:41.075620247-08:00"},{"namespace":"/intel/mock/host7/baz","data":81,"source":"egu-mac01.lan","timestamp":"2015-11-19T23:45:41.075620942-08:00"},{"namespace":"/intel/mock/host8/baz","data":88,"source":"egu-mac01.lan","timestamp":"2015-11-19T23:45:41.075621674-08:00"},{"namespace":"/intel/mock/host9/baz","data":85,"source":"egu-mac01.lan","timestamp":"2015-11-19T23:45:41.075623754-08:00"},{"namespace":"/intel/mock/bar","data":69,"source":"egu-mac01.lan","timestamp":"2015-11-19T23:45:41.075630288-08:00"},{"namespace":"/intel/mock/foo","data":87,"source":"egu-mac01.lan","timestamp":"2015-11-19T23:45:41.075635543-08:00"}]}
{"type":"metric-event","message":"","event":[{"namespace":"/intel/mock/host0/baz","data":87,"source":"egu-mac01.lan","timestamp":"2015-11-19T23:45:42.075605924-08:00"},{"namespace":"/intel/mock/host1/baz","data":89,"source":"egu-mac01.lan","timestamp":"2015-11-19T23:45:42.075609242-08:00"},{"namespace":"/intel/mock/host2/baz","data":84,"source":"egu-mac01.lan","timestamp":"2015-11-19T23:45:42.075611747-08:00"},{"namespace":"/intel/mock/host3/baz","data":82,"source":"egu-mac01.lan","timestamp":"2015-11-19T23:45:42.075613786-08:00"}...
```
**POST /v1/tasks**:
Create a task with the JSON input, using for example mock-file.json with following content:
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
                "/intel/mock/foo": {},
                "/intel/mock/bar": {},
                "/intel/mock/*/baz": {}
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
                    "process": null,
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
curl -vXPOST http://localhost:8181/v1/tasks -d @mock-file.json --header "Content-Type: application/json"
```
_**Example Response**_
```json
{
  "meta": {
    "code": 201,
    "message": "Scheduled task created (15d4e09c-b4ab-4e92-b85e-9d9697212632)",
    "type": "scheduled_task_created",
    "version": 1
  },
  "body": {
    "id": "15d4e09c-b4ab-4e92-b85e-9d9697212632",
    "name": "Task-15d4e09c-b4ab-4e92-b85e-9d9697212632",
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
            "user": "root",
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
                }
              }
            ]
          }
        ]
      }
    },
    "schedule": {
      "type": "simple",
      "interval": "1s"
    },
    "creation_timestamp": 1448007501,
    "last_run_timestamp": -1,
    "task_state": "Stopped"
  }
```
**PUT /v1/tasks/:id/start**:
Start a task given a task ID

_**Example Request**_
```
curl -XPUT http://localhost:8181/v1/tasks/7cd4b229-e12c-4b09-985a-b60e76daac90/start
```
_**Example Response**_
```json
{
  "meta": {
    "code": 200,
    "message": "Scheduled task (7cd4b229-e12c-4b09-985a-b60e76daac90) started",
    "type": "scheduled_task_started",
    "version": 1
  },
  "body": {
    "id": "7cd4b229-e12c-4b09-985a-b60e76daac90"
  }
}             
```
**PUT /v1/tasks/:id/stop**:
Stop a running task given a task ID

_**Example Request**_
```
curl -XPUT http://localhost:8181/v1/tasks/7cd4b229-e12c-4b09-985a-b60e76daac90/stop
```
_**Example Response**_
```json
{
  "meta": {
    "code": 200,
    "message": "Scheduled task (7cd4b229-e12c-4b09-985a-b60e76daac90) stopped",
    "type": "scheduled_task_stopped",
    "version": 1
  },
  "body": {
    "id": "7cd4b229-e12c-4b09-985a-b60e76daac90"
  }
}      
```
**DELETE /v1/tasks/:id**:
Remove a task from the scheduled task list given a task ID

_**Example Request**_
```
curl -X DELETE http://localhost:8181/v1/tasks/7cd4b229-e12c-4b09-985a-b60e76daac90  
```
_**Example Response**_
```json
{
  "meta": {
    "code": 200,
    "message": "Scheduled task (7cd4b229-e12c-4b09-985a-b60e76daac90) removed",
    "type": "scheduled_task_removed",
    "version": 1
  },
  "body": {
    "id": "7cd4b229-e12c-4b09-985a-b60e76daac90"
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
      "hot-agrement": {
        "name": "hot-agrement",
        "plugin_agreement": {
          "plugins": [
            {
              "name": "file",
              "version": 3,
              "type": 2
            }
          ]
        },
        "task_agreement": {},
        "members": {
          "hawaii": {
            "tags": {
              "rest_api_port": "8182",
              "rest_insecure": "",
              "rest_proto": "http"
            },
            "name": "hawaii"
          },
          "maui": {
            "tags": {
              "rest_api_port": "8183",
              "rest_insecure": "",
              "rest_proto": "http"
            },
            "name": "maui"
          },
          "seed": {
            "tags": {
              "rest_api_port": "8181",
              "rest_insecure": "",
              "rest_proto": "http"
            },
            "name": "seed"
          }
        }
      },
      "warm-agreement": {
        "name": "warm-agreement",
        "plugin_agreement": {},
        "task_agreement": {},
        "members": {
          "hawaii": {
            "tags": {
              "rest_api_port": "8182",
              "rest_insecure": "",
              "rest_proto": "http"
            },
            "name": "hawaii"
          },
          "maui": {
            "tags": {
              "rest_api_port": "8183",
              "rest_insecure": "",
              "rest_proto": "http"
            },
            "name": "maui"
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
curl -X POST http://localhost:8182/v1/tribe/agreements -d '{"name":"cold-agreement"}'
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
      "hot-agrement": {
        "name": "hot-agrement",
        "plugin_agreement": {
          "plugins": [
            {
              "name": "file",
              "version": 3,
              "type": 2
            }
          ]
        },
        "task_agreement": {},
        "members": {
          "hawaii": {
            "tags": {
              "rest_api_port": "8182",
              "rest_insecure": "",
              "rest_proto": "http"
            },
            "name": "hawaii"
          },
          "maui": {
            "tags": {
              "rest_api_port": "8183",
              "rest_insecure": "",
              "rest_proto": "http"
            },
            "name": "maui"
          },
          "seed": {
            "tags": {
              "rest_api_port": "8181",
              "rest_insecure": "",
              "rest_proto": "http"
            },
            "name": "seed"
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
curl -L http://localhost:8183/v1/tribe/agreements/warm-agreement
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
      "name": "warm-agreement",
      "plugin_agreement": {},
      "task_agreement": {},
      "members": {
        "hawaii": {
          "tags": {
            "rest_api_port": "8182",
            "rest_insecure": "",
            "rest_proto": "http"
          },
          "name": "hawaii"
        },
        "maui": {
          "tags": {
            "rest_api_port": "8183",
            "rest_insecure": "",
            "rest_proto": "http"
          },
          "name": "maui"
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
curl -X DELETE http://localhost:8183/v1/tribe/agreements/warm-agreement
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
      "hot-agrement": {
        "name": "hot-agrement",
        "plugin_agreement": {
          "plugins": [
            {
              "name": "file",
              "version": 3,
              "type": 2
            }
          ]
        },
        "task_agreement": {},
        "members": {
          "hawaii": {
            "tags": {
              "rest_api_port": "8182",
              "rest_insecure": "",
              "rest_proto": "http"
            },
            "name": "hawaii"
          },
          "maui": {
            "tags": {
              "rest_api_port": "8183",
              "rest_insecure": "",
              "rest_proto": "http"
            },
            "name": "maui"
          },
          "seed": {
            "tags": {
              "rest_api_port": "8181",
              "rest_insecure": "",
              "rest_proto": "http"
            },
            "name": "seed"
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
curl -X PUT http://localhost:8183/v1/tribe/agreements/warm-agreement/join -d '{"member_name": "maui"}'
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
      "name": "warm-agreement",
      "plugin_agreement": {},
      "task_agreement": {},
      "members": {
        "hawaii": {
          "tags": {
            "rest_api_port": "8182",
            "rest_insecure": "",
            "rest_proto": "http"
          },
          "name": "hawaii"
        },
        "maui": {
          "tags": {
            "rest_api_port": "8183",
            "rest_insecure": "",
            "rest_proto": "http"
          },
          "name": "maui"
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
curl -X DELETE http://localhost:8183/v1/tribe/agreements/warm-agreement/leave -d '{"member_name": "maui"}'
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
      "name": "warm-agreement",
      "plugin_agreement": {},
      "task_agreement": {},
      "members": {
        "hawaii": {
          "tags": {
            "rest_api_port": "8182",
            "rest_insecure": "",
            "rest_proto": "http"
          },
          "name": "hawaii"
        }
      }
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
      "hawaii",
      "seed",
      "maui"
    ]
  }
}
```
**GET /v1/tribe/member/:name**:
List tribe member information given the node name

_**Example Request**_
```
curl -L http://localhost:8183/v1/tribe/member/maui
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
    "name": "maui",
    "plugin_agreement": "warm-agreement",
    "tags": {
      "rest_api_port": "8183",
      "rest_insecure": "",
      "rest_proto": "http"
    },
    "task_agreements": null
  }
}
```
