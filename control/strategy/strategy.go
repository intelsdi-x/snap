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

var (
	ErrCouldNotSelect = errors.New("could not select a plugin")
)

type SelectablePlugin interface {
	HitCount() int
	LastHit() time.Time
	String() string
	Kill(r string) error
	ID() uint32
}

type RoutingAndCaching interface {
	Select(selectablePlugins []SelectablePlugin, taskID string) (SelectablePlugin, error)
	Remove(selectablePlugins []SelectablePlugin, taskID string) (SelectablePlugin, error)
	CheckCache(metrics []core.Metric, taskID string) ([]core.Metric, []core.Metric)
	UpdateCache(metrics []core.Metric, taskID string)
	CacheHits(ns string, ver int, taskID string) (uint64, error)
	CacheMisses(ns string, ver int, taskID string) (uint64, error)
	AllCacheHits() uint64
	AllCacheMisses() uint64
	CacheTTL(taskID string) (time.Duration, error)
	String() string
}
