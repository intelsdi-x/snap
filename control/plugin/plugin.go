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

package plugin

// WARNING! Do not import "fmt" and print from a plugin to stdout!
import (
	"crypto/rsa"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/intelsdi-x/snap/control/plugin/cpolicy"
)

// Plugin type
type PluginType int

// Returns string for matching enum plugin type
func (p PluginType) String() string {
	return types[p]
}

const (
	CollectorPluginType PluginType = iota
	ProcessorPluginType
	PublisherPluginType
	StreamCollectorPluginType
)

type RoutingStrategyType int

// Returns string for matching enum RoutingStrategy type
func (p RoutingStrategyType) String() string {
	return routingStrategyTypes[p]
}

const (
	// DefaultRouting is a least recently used strategy.
	DefaultRouting RoutingStrategyType = iota
	// StickyRouting is a one-to-one strategy.
	// Using this strategy a tasks requests are sent to the same running instance of a plugin.
	StickyRouting
	// ConfigRouting is routing to plugins based on the config provided to the plugin.
	// Using this strategy enables a running database plugin that has the same connection info between
	// two tasks to be shared.
	ConfigRouting
)

// Plugin response states
type PluginResponseState int

const (
	PluginSuccess PluginResponseState = iota
	PluginFailure
)

type RPCType int

const (
	// IMPORTANT: keep consistency across snap-plugin-lib, GRPC must be equal 2
	NativeRPC  RPCType = 0
	GRPC       RPCType = 2
	STREAMGRPC RPCType = 3
)

var (
	// Timeout settings
	// How much time must elapse before a lack of Ping results in a timeout
	PingTimeoutDurationDefault = time.Second * 10

	// Array matching plugin type enum to a string
	// note: in string representation we use lower case
	types = [...]string{
		"collector",
		"processor",
		"publisher",
		"streamCollector",
	}

	routingStrategyTypes = [...]string{
		"least-recently-used",
		"sticky",
		"config",
	}
)

type Plugin interface {
	GetConfigPolicy() (*cpolicy.ConfigPolicy, error)
}

// PluginMeta for plugin
type PluginMeta struct {
	Name       string
	Version    int
	Type       PluginType
	RPCType    RPCType
	RPCVersion int
	// AcceptedContentTypes are types accepted by this plugin in priority order.
	// snap.* means any snap type.
	AcceptedContentTypes []string
	// ReturnedContentTypes are content types returned in priority order.
	// This is only applicable on processors.
	ReturnedContentTypes []string
	// ConcurrencyCount is the max number concurrent calls the plugin may take.
	// If there are 5 tasks using the plugin and concurrency count is 2 there
	// will be 3 plugins running.
	ConcurrencyCount int
	// Exclusive results in a single instance of the plugin running regardless
	// the number of tasks using the plugin.
	Exclusive bool
	// Unsecure results in unencrypted communication with this plugin.
	Unsecure bool
	// CacheTTL will override the default cache TTL for the provided plugin.
	CacheTTL time.Duration
	// RoutingStrategy will override the routing strategy this plugin requires.
	// The default routing strategy round-robin.
	RoutingStrategy RoutingStrategyType
}

// Arguments passed to startup of Plugin
type Arg struct {
	// Plugin log level
	LogLevel log.Level
	// Ping timeout duration
	PingTimeoutDuration time.Duration

	NoDaemon bool
	// The listen port
	listenPort string

	// enable pprof
	Pprof bool
}

func NewArg(logLevel int, pprof bool) Arg {
	return Arg{
		LogLevel:            log.Level(logLevel),
		PingTimeoutDuration: PingTimeoutDurationDefault,
		Pprof:               pprof,
	}
}

// Response from started plugin
type Response struct {
	Meta          PluginMeta
	ListenAddress string
	PprofAddress  string
	Token         string
	Type          PluginType
	// State is a signal from plugin to control that it passed
	// its own loading requirements
	State        PluginResponseState
	ErrorMessage string
	PublicKey    *rsa.PublicKey
}
