package v1

import (
	"sync"

	"github.com/intelsdi-x/snap/mgmt/rest/api"
	log "github.com/sirupsen/logrus"
)

const (
	// TLSCertPrefix defines a prefix for file fragment carrying path to TLS certificate
	TLSCertPrefix = "crt."
	// TLSKeyPrefix defines a prefix for file fragment carrying path to TLS private key
	TLSKeyPrefix = "key."
	// TLSCACertsPrefix defines a prefix for file fragment carrying paths to TLS CA certificates
	TLSCACertsPrefix = "cacerts."

	version = "v1"
	prefix  = "/" + version
)

var (
	restLogger     = log.WithField("_module", "_mgmt-rest-v1")
	protocolPrefix = "http"
)

type apiV1 struct {
	metricManager api.Metrics
	taskManager   api.Tasks
	tribeManager  api.Tribe
	configManager api.Config

	wg       *sync.WaitGroup
	killChan chan struct{}
}

func New(wg *sync.WaitGroup, killChan chan struct{}, protocol string) *apiV1 {
	protocolPrefix = protocol
	return &apiV1{wg: wg, killChan: killChan}
}

func (s *apiV1) GetRoutes() []api.Route {
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
		api.Route{Method: "PUT", Path: prefix + "/tasks/:id/start", Handle: s.startTask},
		api.Route{Method: "PUT", Path: prefix + "/tasks/:id/stop", Handle: s.stopTask},
		api.Route{Method: "DELETE", Path: prefix + "/tasks/:id", Handle: s.removeTask},
		api.Route{Method: "PUT", Path: prefix + "/tasks/:id/enable", Handle: s.enableTask},
	}
	// tribe routes
	if s.tribeManager != nil {
		routes = append(routes, []api.Route{
			api.Route{Method: "GET", Path: prefix + "/tribe/agreements", Handle: s.getAgreements},
			api.Route{Method: "POST", Path: prefix + "/tribe/agreements", Handle: s.addAgreement},
			api.Route{Method: "GET", Path: prefix + "/tribe/agreements/:name", Handle: s.getAgreement},
			api.Route{Method: "DELETE", Path: prefix + "/tribe/agreements/:name", Handle: s.deleteAgreement},
			api.Route{Method: "PUT", Path: prefix + "/tribe/agreements/:name/join", Handle: s.joinAgreement},
			api.Route{Method: "DELETE", Path: prefix + "/tribe/agreements/:name/leave", Handle: s.leaveAgreement},
			api.Route{Method: "GET", Path: prefix + "/tribe/members", Handle: s.getMembers},
			api.Route{Method: "GET", Path: prefix + "/tribe/member/:name", Handle: s.getMember},
		}...)
	}
	return routes
}

func (s *apiV1) BindMetricManager(metricManager api.Metrics) {
	s.metricManager = metricManager
}

func (s *apiV1) BindTaskManager(taskManager api.Tasks) {
	s.taskManager = taskManager
}

func (s *apiV1) BindTribeManager(tribeManager api.Tribe) {
	s.tribeManager = tribeManager
}

func (s *apiV1) BindConfigManager(configManager api.Config) {
	s.configManager = configManager
}
