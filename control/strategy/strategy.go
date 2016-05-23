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

// Package strategy provides basic interfaces for routing to available
// plugins and caching metric data.
package strategy

import (
	"errors"
	"time"

	"github.com/intelsdi-x/snap/core"
)

type MapAvailablePlugin map[uint32]AvailablePlugin

var (
	ErrCouldNotSelect = errors.New("could not select a plugin")
)

type RoutingAndCaching interface {
	Select(availablePlugins []AvailablePlugin, id string) (AvailablePlugin, error)
	Remove(availablePlugins []AvailablePlugin, id string) (AvailablePlugin, error)
	CheckCache(metrics []core.Metric, id string) ([]core.Metric, []core.Metric)
	UpdateCache(metrics []core.Metric, id string)
	CacheHits(ns string, ver int, id string) (uint64, error)
	CacheMisses(ns string, ver int, id string) (uint64, error)
	AllCacheHits() uint64
	AllCacheMisses() uint64
	CacheTTL(taskID string) (time.Duration, error)
	String() string
}

// Values returns slice of map values
func (sm MapAvailablePlugin) Values() []AvailablePlugin {
	values := []AvailablePlugin{}
	for _, v := range sm {
		values = append(values, v)
	}
	return values
}
