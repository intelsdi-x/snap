package routing

import (
	"time"
)

type SelectablePluginPool interface {
}

type SelectablePlugin interface {
	HitCount() int
	LastHit() time.Time
	String() string
}
