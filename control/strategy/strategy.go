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

// Package strategy provides basic interfaces for routing to availble
// plugins and caching metric data.
package strategy

import (
	"time"

	"github.com/intelsdi-x/snap/core"
)

type SelectablePlugin interface {
	HitCount() int
	LastHit() time.Time
	String() string
}

type RoutingAndCaching interface {
	Select([]SelectablePlugin) (SelectablePlugin, error)
	CheckCache(mts []core.Metric) ([]core.Metric, []core.Metric)
	UpdateCache(mts []core.Metric)
	CacheHits(string, int) (uint64, error)
	CacheMisses(string, int) (uint64, error)
	AllCacheHits() uint64
	AllCacheMisses() uint64
	CacheTTL() time.Duration
	String() string
}
