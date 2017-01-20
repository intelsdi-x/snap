package v2

import (
	"bytes"
	"encoding/json"
	"fmt"
	"sync"

	"net/http"

	log "github.com/Sirupsen/logrus"
	"github.com/intelsdi-x/snap/core/api"
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

type V2 struct {
	metricManager api.Metrics
	taskManager   api.Tasks
	configManager api.Config

	wg       *sync.WaitGroup
	killChan chan struct{}
}

func New(wg *sync.WaitGroup, killChan chan struct{}, protocol string) *V2 {
	protocolPrefix = protocol
	return &V2{wg: wg, killChan: killChan}
}

func (s *V2) GetRoutes() []api.Route {
	routes := []api.Route{
		// plugin routes
		api.Route{Method: "GET", Path: prefix + "/plugins", Handle: s.getPlugins},
		api.Route{Method: "GET", Path: prefix + "/plugins/:type", Handle: s.getPlugins},
		api.Route{Method: "GET", Path: prefix + "/plugins/:type/:name", Handle: s.getPlugins},
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
		api.Route{Method: "PUT", Path: prefix + "/tasks/:id", Handle: s.updateTask},
		api.Route{Method: "DELETE", Path: prefix + "/tasks/:id", Handle: s.removeTask},
	}
	return routes
}

func (s *V2) BindMetricManager(metricManager api.Metrics) {
	s.metricManager = metricManager
}

func (s *V2) BindTaskManager(taskManager api.Tasks) {
	s.taskManager = taskManager
}

func (s *V2) BindTribeManager(tribeManager api.Tribe) {}

func (s *V2) BindConfigManager(configManager api.Config) {
	s.configManager = configManager
}

func Write(code int, body interface{}, w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json; version=2")
	w.Header().Set("Version", "beta")

	if !w.(negroni.ResponseWriter).Written() {
		w.WriteHeader(code)
	}

	if body != nil {
		j, err := json.MarshalIndent(body, "", "  ")
		if err != nil {
			restLogger.Fatalln(err)
		}
		j = bytes.Replace(j, []byte("\\u0026"), []byte("&"), -1)
		fmt.Fprint(w, string(j))
	}
}
