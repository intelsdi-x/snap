/*
http://www.apache.org/licenses/LICENSE-2.0.txt


Copyright 2015 Intel Coporation

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

func (r *RoundRobinStrategy) String() string {
	return "round-robin"
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
			"_module":   "control-routing",
			"block":     "select",
			"strategy":  "round-robin",
			"pool size": len(spa),
			"index":     spa[index].String(),
			"hitcount":  spa[index].HitCount(),
		}).Debug("plugin selected")
		return spa[index], nil
	}
	log.WithFields(log.Fields{
		"_module":  "control-routing",
		"block":    "select",
		"strategy": "round-robin",
		"error":    ErrorCouldNotSelect,
	}).Debug("error selecting")
	return nil, ErrorCouldNotSelect
}
