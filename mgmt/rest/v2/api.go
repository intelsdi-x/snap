/*
http://www.apache.org/licenses/LICENSE-2.0.txt


Copyright 2017 Intel Corporation

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

package v2

import (
	"encoding/json"
	"sync"

	"net/http"

	log "github.com/Sirupsen/logrus"
	"github.com/intelsdi-x/snap/mgmt/rest/api"
	"github.com/urfave/negroni"
)

const (
	version = "v2"
	prefix  = "/" + version
)

var (
	restLogger     = log.WithField("_module", "_mgmt-rest-v2")
	protocolPrefix = "http"
)

type apiV2 struct {
	metricManager api.Metrics
	taskManager   api.Tasks
	configManager api.Config

	wg       *sync.WaitGroup
	killChan chan struct{}
}

func New(wg *sync.WaitGroup, killChan chan struct{}, protocol string) *apiV2 {
	protocolPrefix = protocol
	return &apiV2{wg: wg, killChan: killChan}
}

func (s *apiV2) GetRoutes() []api.Route {
	routes := []api.Route{
		// swagger:route GET /plugins getPlugins
		//
		// lists a list of loaded plugins. An empty list is returned if there is no loaded plugins.
		//
		// Consumes:
		// application/json
		// application/x-protobuf
		//
		// Produces:
		// application/json
		// application/x-protobuf
		//
		// Schemes: http, https
		//
		// Responses:
		// 200: PluginsResponse
		api.Route{Method: "GET", Path: prefix + "/plugins", Handle: s.getPlugins},
		// swagger:route GET /plugins/{ptype}/{pname}/{pversion} getPlugin
		//
		// lists a given plugin by its type, name and version. No plugin found error returns if it's not existing.
		//
		// Consumes:
		// application/json
		// application/x-protobuf
		//
		// Produces:
		// application/json
		// application/x-protobuf
		//
		// Schemes: http, https
		//
		// Responses:
		// 200: PluginResponse
		// 400: ErrorResponse
		// 404: ErrorResponse
		// 500: ErrorResponse
		api.Route{Method: "GET", Path: prefix + "/plugins/:type/:name/:version", Handle: s.getPlugin},
		// swagger:route POST /plugins loadPlugin
		//
		// loads a plugin based on input.
		//
		// Consumes:
		// application/json
		// application/x-protobuf
		// multipart/form-data
		//
		// Produces:
		// application/json
		// application/x-protobuf
		// multipart/form-data
		//
		// Schemes: http, https
		//
		// Responses:
		// 201: PluginResponse
		// 400: ErrorResponse
		// 409: ErrorResponse
		// 415: ErrorResponse
		// 500: ErrorResponse
		api.Route{Method: "POST", Path: prefix + "/plugins", Handle: s.loadPlugin},
		// swagger:route POST /plugins/{ptype}/{pname}/{pversion}/swap swapPlugins
		//
		// unloads an existing plugin then loads a new plugin.
		//
		// Consumes:
		// application/json
		// application/x-protobuf
		// multipart/form-data
		//
		// Produces:
		// application/json
		// application/x-protobuf
		// multipart/form-data
		//
		// Schemes: http, https
		//
		// Responses:
		// 201: PluginResponse
		// 400: ErrorResponse
		// 409: ErrorResponse
		// 415: ErrorResponse
		// 500: ErrorResponse
		api.Route{Method: "POST", Path: prefix + "/plugins/:type/:name/:version/swap", Handle: s.swapPlugins},
		// swagger:route DELETE /plugins/{ptype}/{pname}/{pversion} unloadPlugin
		//
		// unloads a plugin by its type, name and version.Otherwise, an error is returned.
		//
		// Consumes:
		// application/json
		// application/x-protobuf
		// text/plain
		//
		// Produces:
		// application/json
		// application/x-protobuf
		// text/plain
		//
		// Schemes: http, https
		//
		// Responses:
		// 204: PluginResponse
		// 400: ErrorResponse
		// 404: ErrorResponse
		// 409: ErrorResponse
		// 500: ErrorResponse
		api.Route{Method: "DELETE", Path: prefix + "/plugins/:type/:name/:version", Handle: s.unloadPlugin},
		// swagger:route GET /plugins/{ptype}/{pname}/{pversion}/config getPluginConfigItem
		//
		// lists the config of a giving plugin. The allowed plugin types are collector, processor, and publisher.
		// Any other type results in error.
		//
		// Consumes:
		// application/json
		// application/x-protobuf
		//
		// Produces:
		// application/json
		// application/x-protobuf
		//
		// Schemes: http, https
		//
		// Responses:
		// 200: PluginConfigResponse
		// 400: ErrorResponse
		api.Route{Method: "GET", Path: prefix + "/plugins/:type/:name/:version/config", Handle: s.getPluginConfigItem},
		// swagger:route PUT /plugins/{ptype}/{pname}/{pversion}/config setPluginConfigItem
		//
		// updates the config of a giving plugin. A wrong plugin type or non-numeric plugin version results in error.
		//
		// Consumes:
		// application/json
		// application/x-protobuf
		//
		// Produces:
		// application/json
		// application/x-protobuf
		//
		// Schemes: http, https
		//
		// Responses:
		// 200: PluginConfigResponse
		// 400: ErrorResponse
		api.Route{Method: "PUT", Path: prefix + "/plugins/:type/:name/:version/config", Handle: s.setPluginConfigItem},
		// swagger:route DELETE /plugins/{ptype}/{pname}/{pversion}/config deletePluginConfigItem
		//
		// deletes the config of a giving plugin. Note that that to be removed config items are a slice of config keys.
		// At lease one config key is required for this operation. An error occurs for any bad request.
		//
		// Consumes:
		// application/json
		// application/x-protobuf
		//
		// Produces:
		// application/json
		// application/x-protobuf
		//
		// Schemes: http, https
		//
		// Responses:
		// 200: PluginConfigResponse
		// 400: ErrorResponse
		api.Route{Method: "DELETE", Path: prefix + "/plugins/:type/:name/:version/config", Handle: s.deletePluginConfigItem},
		// swagger:route GET /metrics getMetrics
		//
		// lists a list of loaded metric types. An empty list returns if there is no loaded metrics. Any bad request results in error.
		//
		// Consumes:
		// application/json
		// application/x-protobuf
		//
		// Produces:
		// application/json
		// application/x-protobuf
		//
		// Schemes: http, https
		//
		// Responses:
		// 200: MetricsResponse
		// 404: ErrorResponse
		// 500: ErrorResponse
		api.Route{Method: "GET", Path: prefix + "/metrics", Handle: s.getMetrics},
		// swagger:route GET /tasks getTasks
		//
		// lists a list of created tasks. An empty list returns if no created tasks.
		//
		// Consumes:
		// application/json
		// application/x-protobuf
		//
		// Produces:
		// application/json
		// application/x-protobuf
		//
		// Schemes: http, https
		//
		// Responses:
		// 200: TasksResponse
		api.Route{Method: "GET", Path: prefix + "/tasks", Handle: s.getTasks},

		// swagger:route GET /tasks/{id} getTask
		//
		// lists a task by the giving task id. Otherwise a not found error returns.
		//
		// Consumes:
		// application/json
		// application/x-protobuf
		//
		// Produces:
		// application/json
		// application/x-protobuf
		//
		// Schemes: http, https
		//
		// Responses:
		// 200: TaskResponse
		// 404: ErrorResponse
		api.Route{Method: "GET", Path: prefix + "/tasks/:id", Handle: s.getTask},
		// swagger:route GET /tasks/{id}/watch watchTask
		//
		// watches a task data stream for the giving task id. Otherwise, an error returns.
		//
		// Consumes:
		// application/json
		// application/x-protobuf
		// text/event-stream
		//
		// Produces:
		// application/json
		// application/x-protobuf
		// text/event-stream
		//
		// Schemes: http, https
		//
		// Responses:
		// 200: TaskWatchResponse
		// 404: ErrorResponse
		// 500: ErrorResponse
		api.Route{Method: "GET", Path: prefix + "/tasks/:id/watch", Handle: s.watchTask},
		// swagger:route POST /tasks addTask
		//
		// creates a task based on the input. Othereise, an error returns if the input misses the required fields or is in a malformed format.
		//
		// Consumes:
		// application/json
		// application/x-protobuf
		//
		// Produces:
		// application/json
		// application/x-protobuf
		//
		// Schemes: http, https
		//
		// Responses:
		// 201: TaskResponse
		// 500: ErrorResponse
		api.Route{Method: "POST", Path: prefix + "/tasks", Handle: s.addTask},
		// swagger:route PUT /tasks/{id} updateTaskState
		//
		// updates a task's state for the giving task id and the input state. An error occurs for any bad request.
		//
		// Consumes:
		// application/json
		// application/x-protobuf
		//
		// Produces:
		// application/json
		// application/x-protobuf
		//
		// Schemes: http, https
		//
		// Responses:
		// 204: TaskResponse
		// 400: ErrorResponse
		// 409: ErrorResponse
		// 500: ErrorResponse
		api.Route{Method: "PUT", Path: prefix + "/tasks/:id", Handle: s.updateTaskState},
		// swagger:route DELETE /tasks/{id} removeTask
		//
		// deletes a task for the giving task id. Note that only a stopped task may be removed. Otherwise, an error occurs.
		//
		// Consumes:
		// application/json
		// application/x-protobuf
		//
		// Produces:
		// application/json
		// application/x-protobuf
		//
		// Schemes: http, https
		//
		// Responses:
		// 204: TaskResponse
		// 404: ErrorResponse
		// 500: TaskErrorResponse
		api.Route{Method: "DELETE", Path: prefix + "/tasks/:id", Handle: s.removeTask},
	}
	return routes
}

func (s *apiV2) BindMetricManager(metricManager api.Metrics) {
	s.metricManager = metricManager
}

func (s *apiV2) BindTaskManager(taskManager api.Tasks) {
	s.taskManager = taskManager
}

func (s *apiV2) BindTribeManager(tribeManager api.Tribe) {}

func (s *apiV2) BindConfigManager(configManager api.Config) {
	s.configManager = configManager
}

func Write(code int, body interface{}, w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json; version=2; charset=utf-8")
	w.Header().Set("Version", "beta")

	if !w.(negroni.ResponseWriter).Written() {
		w.WriteHeader(code)
	}

	if body != nil {
		e := json.NewEncoder(w)
		e.SetIndent("", "  ")
		e.SetEscapeHTML(false)
		err := e.Encode(body)
		if err != nil {
			restLogger.Fatalln(err)
		}
	}
}
