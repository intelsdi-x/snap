package v2

import (
	"sync"

	log "github.com/Sirupsen/logrus"
	"github.com/intelsdi-x/snap/pkg/api"
)

const (
	version = "v2"
	prefix  = "/" + version
)

var (
	restLogger     = log.WithField("_module", "_mgmt-rest-v1")
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
		api.Route{Method: "GET", Path: prefix + "/metrics/*namespace", Handle: s.getMetricsFromTree},

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
