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

func NewV2(wg *sync.WaitGroup, killChan chan struct{}, protocol string) *V2 {
	protocolPrefix = protocol
	return &V2{wg: wg, killChan: killChan}
}

func (s *V2) GetRoutes() []api.Route {
	routes := []api.Route{
		api.Route{Method: "GET", Path: prefix + "/plugins", Handler: s.getPlugins},
		api.Route{Method: "GET", Path: prefix + "/plugins/:type", Handler: s.getPlugins},
		api.Route{Method: "GET", Path: prefix + "/plugins/:type/:name", Handler: s.getPlugins},
		api.Route{Method: "GET", Path: prefix + "/plugins/:type/:name/:version", Handler: s.getPlugin},
		api.Route{Method: "POST", Path: prefix + "/plugins", Handler: s.loadPlugin},
		api.Route{Method: "DELETE", Path: prefix + "/plugins/:type/:name/:version", Handler: s.unloadPlugin},
		api.Route{Method: "GET", Path: prefix + "/plugins/:type/:name/:version/config", Handler: s.getPluginConfigItem},
		api.Route{Method: "PUT", Path: prefix + "/plugins/:type/:name/:version/config", Handler: s.setPluginConfigItem},
		api.Route{Method: "DELETE", Path: prefix + "/plugins/:type/:name/:version/config", Handler: s.deletePluginConfigItem},

		// metric routes
		api.Route{Method: "GET", Path: prefix + "/metrics", Handler: s.getMetrics},
		api.Route{Method: "GET", Path: prefix + "/metrics/*namespace", Handler: s.getMetricsFromTree},

		// task routes
		api.Route{Method: "GET", Path: prefix + "/tasks", Handler: s.getTasks},
		api.Route{Method: "GET", Path: prefix + "/tasks/:id", Handler: s.getTask},
		api.Route{Method: "GET", Path: prefix + "/tasks/:id/watch", Handler: s.watchTask},
		api.Route{Method: "POST", Path: prefix + "/tasks", Handler: s.addTask},
		api.Route{Method: "PUT", Path: prefix + "/tasks/:id", Handler: s.updateTask},
		api.Route{Method: "DELETE", Path: prefix + "/tasks/:id", Handler: s.removeTask},
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
