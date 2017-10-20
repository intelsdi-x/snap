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

	"github.com/intelsdi-x/snap/mgmt/rest/api"
	log "github.com/sirupsen/logrus"
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
		// swagger:route GET /plugins plugins getPlugins
		//
		// Get All
		//
		// An empty list is returned if there are no loaded plugins.
		//
		// Produces:
		// application/json
		//
		// Schemes: http, https
		//
		// Responses:
		// 200: PluginsResponse
		// 401: UnauthResponse
		api.Route{Method: "GET", Path: prefix + "/plugins", Handle: s.getPlugins},
		// swagger:route GET /plugins/{ptype}/{pname}/{pversion} plugins getPlugin
		//
		// Get
		//
		// An error will be returned if the plugin does not exist.
		//
		// Produces:
		// application/json
		//
		// Schemes: http, https
		//
		// Responses:
		// 200: PluginResponse
		// 400: ErrorResponse
		// 404: ErrorResponse
		// 500: ErrorResponse
		// 401: UnauthResponse
		api.Route{Method: "GET", Path: prefix + "/plugins/:type/:name/:version", Handle: s.getPlugin},
		// swagger:route POST /plugins plugins loadPlugin
		//
		// Load
		//
		// A plugin binary is required.
		//
		// Consumes:
		// multipart/form-data
		//
		// Produces:
		// application/json
		//
		// Schemes: http, https
		//
		// Responses:
		// 201: PluginResponse
		// 400: ErrorResponse
		// 409: ErrorResponse
		// 415: ErrorResponse
		// 500: ErrorResponse
		// 401: UnauthResponse
		api.Route{Method: "POST", Path: prefix + "/plugins", Handle: s.loadPlugin},
		// swagger:route DELETE /plugins/{ptype}/{pname}/{pversion} plugins unloadPlugin
		//
		// Unload
		//
		// Required fields are plugin type, name and version.
		//
		// Produces:
		// application/json
		//
		// Schemes: http, https
		//
		// Responses:
		// 204: PluginResponse
		// 400: ErrorResponse
		// 404: ErrorResponse
		// 409: ErrorResponse
		// 500: ErrorResponse
		// 401: UnauthResponse
		api.Route{Method: "DELETE", Path: prefix + "/plugins/:type/:name/:version", Handle: s.unloadPlugin},
		// swagger:route GET /plugins/{ptype}/{pname}/{pversion}/config plugins getPluginConfigItem
		//
		// Get Config
		//
		// An empty config is returned if there are no configs for the plugin.
		//
		// Produces:
		// application/json
		//
		// Schemes: http, https
		//
		// Responses:
		// 200: PluginConfigResponse
		// 400: ErrorResponse
		// 401: UnauthResponse
		api.Route{Method: "GET", Path: prefix + "/plugins/:type/:name/:version/config", Handle: s.getPluginConfigItem},
		// swagger:route PUT /plugins/{ptype}/{pname}/{pversion}/config plugins setPluginConfigItem
		//
		// Set Config
		//
		// A config is JSON. For example: {"user":"snap", "host":"ocean_eleven"}.
		//
		// Consumes:
		// application/json
		//
		// Produces:
		// application/json
		//
		// Schemes: http, https
		//
		// Responses:
		// 200: PluginConfigResponse
		// 400: ErrorResponse
		// 401: UnauthResponse
		api.Route{Method: "PUT", Path: prefix + "/plugins/:type/:name/:version/config", Handle: s.setPluginConfigItem},
		// swagger:route DELETE /plugins/{ptype}/{pname}/{pversion}/config plugins deletePluginConfigItem
		//
		// Delete Config
		//
		// A minimum of one config key is required for this operation.
		//
		// Consumes:
		// application/json
		//
		// Produces:
		// application/json
		//
		// Schemes: http, https
		//
		// Responses:
		// 200: PluginConfigResponse
		// 400: ErrorResponse
		// 401: UnauthResponse
		api.Route{Method: "DELETE", Path: prefix + "/plugins/:type/:name/:version/config", Handle: s.deletePluginConfigItem},
		// swagger:route GET /metrics plugins getMetrics
		//
		// Get Metrics
		//
		// An empty list returns if there is no loaded metrics.
		//
		// Produces:
		// application/json
		//
		// Schemes: http, https
		//
		// Responses:
		// 200: MetricsResponse
		// 404: ErrorResponse
		// 500: ErrorResponse
		// 401: UnauthResponse
		api.Route{Method: "GET", Path: prefix + "/metrics", Handle: s.getMetrics},
		// swagger:route GET /tasks tasks getTasks
		//
		// Get All
		//
		// An empty list returns if no tasks exist.
		//
		// Produces:
		// application/json
		//
		// Schemes: http, https
		//
		// Responses:
		// 200: TasksResponse
		// 401: UnauthResponse
		api.Route{Method: "GET", Path: prefix + "/tasks", Handle: s.getTasks},
		// swagger:route GET /tasks/{id} tasks getTask
		//
		// Get
		//
		// The task ID is required.
		//
		// Produces:
		// application/json
		//
		// Schemes: http, https
		//
		// Responses:
		// 200: TaskResponse
		// 404: ErrorResponse
		// 401: UnauthResponse
		api.Route{Method: "GET", Path: prefix + "/tasks/:id", Handle: s.getTask},
		// swagger:route GET /tasks/{id}/watch tasks watchTask
		//
		// Watch
		//
		// The task ID is required.
		//
		// Produces:
		// text/event-stream
		//
		// Schemes: http, https
		//
		// Responses:
		// 200: TaskWatchResponse
		// 404: ErrorResponse
		// 500: ErrorResponse
		// 401: UnauthResponse
		api.Route{Method: "GET", Path: prefix + "/tasks/:id/watch", Handle: s.watchTask},
		// swagger:route POST /tasks tasks addTask
		//
		// Add
		//
		// A string representation of Snap task manifest is required.
		//
		// Consumes:
		// application/json
		//
		// Produces:
		// application/json
		//
		// Schemes: http, https
		//
		// Responses:
		// 201: TaskResponse
		// 500: ErrorResponse
		// 401: UnauthResponse
		api.Route{Method: "POST", Path: prefix + "/tasks", Handle: s.addTask},
		// swagger:route PUT /tasks/{id} tasks updateTaskState
		//
		// Enable/Start/Stop
		//
		// The task ID is required.
		//
		// Consumes:
		// application/json
		//
		// Produces:
		// application/json
		//
		// Schemes: http, https
		//
		// Responses:
		// 204: TaskResponse
		// 400: ErrorResponse
		// 409: ErrorResponse
		// 500: ErrorResponse
		// 401: UnauthResponse
		api.Route{Method: "PUT", Path: prefix + "/tasks/:id", Handle: s.updateTaskState},
		// swagger:route DELETE /tasks/{id} tasks removeTask
		//
		// Remove
		//
		// The task ID is required.
		//
		// Produces:
		// application/json
		//
		// Schemes: http, https
		//
		// Responses:
		// 204: TaskResponse
		// 404: ErrorResponse
		// 500: TaskErrorResponse
		// 401: UnauthResponse
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
