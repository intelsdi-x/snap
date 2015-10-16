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
	PluginPath() string
	LoadedTimestamp() *time.Time
}

// the collection of cataloged plugins used
// by mgmt modules
type PluginCatalog []CatalogedPlugin

type SubscribedPlugin interface {
	Plugin
	Config() *cdata.ConfigDataNode
}
