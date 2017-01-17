package v1

import (
	"sync"

	log "github.com/Sirupsen/logrus"
	"github.com/intelsdi-x/snap/pkg/api"
)

const (
	version = "v1"
	prefix  = "/" + version
)

var (
	restLogger     = log.WithField("_module", "_mgmt-rest-v1")
	protocolPrefix = "http"
)

type V1 struct {
	metricManager api.Metrics
	taskManager   api.Tasks
	tribeManager  api.Tribe
	configManager api.Config

	wg       *sync.WaitGroup
	killChan chan struct{}
}

func NewV1(wg *sync.WaitGroup, killChan chan struct{}, protocol string) *V1 {
	protocolPrefix = protocol
	return &V1{wg: wg, killChan: killChan}
}

func (s *V1) GetRoutes() []api.Route {
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
		api.Route{Method: "PUT", Path: prefix + "/tasks/:id/start", Handler: s.startTask},
		api.Route{Method: "PUT", Path: prefix + "/tasks/:id/stop", Handler: s.stopTask},
		api.Route{Method: "DELETE", Path: prefix + "/tasks/:id", Handler: s.removeTask},
		api.Route{Method: "PUT", Path: prefix + "/tasks/:id/enable", Handler: s.enableTask},
	}
	// tribe routes
	if s.tribeManager != nil {
		routes = append(routes, []api.Route{
			api.Route{Method: "GET", Path: prefix + "/tribe/agreements", Handler: s.getAgreements},
			api.Route{Method: "POST", Path: prefix + "/tribe/agreements", Handler: s.addAgreement},
			api.Route{Method: "GET", Path: prefix + "/tribe/agreements/:name", Handler: s.getAgreement},
			api.Route{Method: "DELETE", Path: prefix + "/tribe/agreements/:name", Handler: s.deleteAgreement},
			api.Route{Method: "PUT", Path: prefix + "/tribe/agreements/:name/join", Handler: s.joinAgreement},
			api.Route{Method: "DELETE", Path: prefix + "/tribe/agreements/:name/leave", Handler: s.leaveAgreement},
			api.Route{Method: "GET", Path: prefix + "/tribe/members", Handler: s.getMembers},
			api.Route{Method: "GET", Path: prefix + "/tribe/member/:name", Handler: s.getMember},
		}...)
	}
	return routes
}

func (s *V1) BindMetricManager(metricManager api.Metrics) {
	s.metricManager = metricManager
}

func (s *V1) BindTaskManager(taskManager api.Tasks) {
	s.taskManager = taskManager
}

func (s *V1) BindTribeManager(tribeManager api.Tribe) {
	s.tribeManager = tribeManager
}

func (s *V1) BindConfigManager(configManager api.Config) {
	s.configManager = configManager
}
