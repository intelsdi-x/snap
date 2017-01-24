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
		// plugin routes
		api.Route{Method: "GET", Path: prefix + "/plugins", Handle: s.getPlugins},
		api.Route{Method: "GET", Path: prefix + "/plugins/:type/:name/:version", Handle: s.getPlugin},
		api.Route{Method: "POST", Path: prefix + "/plugins", Handle: s.loadPlugin},
		api.Route{Method: "DELETE", Path: prefix + "/plugins/:type/:name/:version", Handle: s.unloadPlugin},

		api.Route{Method: "GET", Path: prefix + "/plugins/:type/:name/:version/config", Handle: s.getPluginConfigItem},
		api.Route{Method: "PUT", Path: prefix + "/plugins/:type/:name/:version/config", Handle: s.setPluginConfigItem},
		api.Route{Method: "DELETE", Path: prefix + "/plugins/:type/:name/:version/config", Handle: s.deletePluginConfigItem},

		// metric routes
		api.Route{Method: "GET", Path: prefix + "/metrics", Handle: s.getMetrics},

		// task routes
		api.Route{Method: "GET", Path: prefix + "/tasks", Handle: s.getTasks},
		api.Route{Method: "GET", Path: prefix + "/tasks/:id", Handle: s.getTask},
		api.Route{Method: "GET", Path: prefix + "/tasks/:id/watch", Handle: s.watchTask},
		api.Route{Method: "POST", Path: prefix + "/tasks", Handle: s.addTask},
		api.Route{Method: "PUT", Path: prefix + "/tasks/:id", Handle: s.updateTaskState},
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
