package core

import "time"

type AvailablePlugin interface {
	Name() string
	Version() int
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
	LoadedTimestamp() int64
}

// the collection of cataloged plugins used
// by mgmt modules
type PluginCatalog []CatalogedPlugin

type PluginType int

const (
	// List of plugin type
	CollectorPluginType PluginType = iota
	PublisherPluginType
	ProcessorPluginType
)

type Plugin interface {
	Name() string
	Version() int
}
