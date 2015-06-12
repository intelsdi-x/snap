package routing

import (
	"errors"
	"math/rand"

	log "github.com/Sirupsen/logrus"
)

var (
	ErrorCouldNotSelect = errors.New("could not select a plugin (round robin strategy)")
)

type RoundRobinStrategy struct {
}

func (r *RoundRobinStrategy) Select(spp SelectablePluginPool, spa []SelectablePlugin) (SelectablePlugin, error) {
	var h int = -1
	var index int = -1
	for i, sp := range spa {
		// look for the lowest hit count
		if sp.HitCount() < h || h == -1 {
			index = i
			h = sp.HitCount()
		}
		// on a hitcount tie we randomly choose one
		if sp.HitCount() == h {
			if rand.Intn(1) == 1 {
				index = i
				h = sp.HitCount()
			}
		}
	}
	if index > -1 {
		log.WithFields(log.Fields{
			"module":    "control-routing",
			"block":     "select",
			"strategy":  "round-robin",
			"pool size": len(spa),
			"index":     spa[index].String(),
			"hitcount":  spa[index].HitCount(),
		}).Debug("plugin selected")
		return spa[index], nil
	}
	log.WithFields(log.Fields{
		"module":   "control-routing",
		"block":    "select",
		"strategy": "round-robin",
		"error":    ErrorCouldNotSelect,
	}).Debug("error selecting")
	return nil, ErrorCouldNotSelect
}
