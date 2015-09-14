package core

import (
	"fmt"
	"time"

	"github.com/intelsdi-x/pulse/core/cdata"
)

type Plugin interface {
	TypeName() string
	Name() string
	Version() int
}

type PluginType int

func ToPluginType(name string) (PluginType, error) {
	pts := map[string]PluginType{
		"collector": 0,
		"processor": 1,
		"publisher": 2,
	}
	t, ok := pts[name]
	if !ok {
		return -1, fmt.Errorf("invalid plugin type name given %s", name)
	}
	return t, nil
}

func (pt PluginType) String() string {
	return []string{
		"collector",
		"processor",
		"publisher",
	}[pt]
}

const (
	// List of plugin type
	CollectorPluginType PluginType = iota
	ProcessorPluginType
	PublisherPluginType
)

type AvailablePlugin interface {
	Plugin
	HitCount() int
	LastHit() time.Time
	ID() uint32
}

// the public interface for a plugin
// this should be the contract for
// how mgmt modules know a plugin
type CatalogedPlugin interface {
	Plugin
	IsSigned() bool
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
