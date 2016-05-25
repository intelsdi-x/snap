/*
http://www.apache.org/licenses/LICENSE-2.0.txt


Copyright 2015 Intel Corporation

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package fixtures

var TaskJSON = `
{
    "collect": {
        "metrics": {
            "/foo/bar": {
                "version": 1
            },
            "/foo/baz": {}
        },
        "config": {
            "/foo/bar": {
                "password": "drowssap",
                "user": "root"
            }
        },
        "tags": {
            "/foo/bar": {
                "tag1": "val1",
                "tag2": "val2"
            },
            "/foo/baz": {
                "tag3": "val3"
            }
        },
        "process": [
            {
                "plugin_name": "floor",
                "plugin_version": 1,
                "process": [
                    {
                        "plugin_name": "oslo",
                        "plugin_version": 1,
                        "process": null,
                        "publish": null,
                        "config": {
                            "version": "kilo"
                        }
                    }
                ],
                "publish": [
                    {
                        "plugin_name": "rabbitmq",
                        "plugin_version": 5,
                        "config": {
                            "port": 5672,
                            "server": "localhost"
                        }
                    }
                ],
                "config": {
                    "something": true,
                    "somethingelse": false
                }
            }
        ],
        "publish": [
            {
                "plugin_name": "riemann",
                "plugin_version": 3,
                "config": {
                    "port": 8080,
                    "user": "root"
                }
            }
        ]
    }
}
`

var TaskYAML = `
---
  collect:
    metrics:
      /foo/bar:
        version: 1
      /foo/baz:
    config:
      /foo/bar:
        user: "root"
        password: "drowssap"
    tags:
      /foo/bar:
        tag1: "val1"
        tag2: "val2"
      /foo/baz:
        tag3: "val3"
    process:
      -
        plugin_name: "floor"
        plugin_version: 1
        config:
          something: true
          somethingelse: false
        process:
          -
            plugin_name: oslo
            plugin_version: 1
            config:
              version: kilo
        publish:
          -
            plugin_name: "rabbitmq"
            plugin_version: 5
            config:
              server: "localhost"
              port: 5672
    publish:
      -
        plugin_name: "riemann"
        plugin_version: 3
        config:
          user: "root"
          port: 8080
`
