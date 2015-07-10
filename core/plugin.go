package core

import (
	"time"

	"github.com/intelsdi-x/pulse/core/cdata"
)

type Plugin interface {
	Name() string
	Version() int
}

type PluginType int

const (
	// List of plugin type
	CollectorPluginType PluginType = iota
	PublisherPluginType
	ProcessorPluginType
)

type AvailablePlugin interface {
	Plugin
	HitCount() int
	LastHit() time.Time
	TypeName() string
	ID() int
}

// the public interface for a plugin
// this should be the contract for
// how mgmt modules know a plugin
type CatalogedPlugin interface {
	Plugin
	TypeName() string
	Status() string
	LoadedTimestamp() *time.Time
}

// the collection of cataloged plugins used
// by mgmt modules
type PluginCatalog []CatalogedPlugin

type SubscribedPlugin interface {
	Plugin
	Config() *cdata.ConfigDataNode
}
