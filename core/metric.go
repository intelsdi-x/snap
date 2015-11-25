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
	"strings"
	"time"

	"github.com/intelsdi-x/snap/control/plugin/cpolicy"
	"github.com/intelsdi-x/snap/core/cdata"
)

type Label struct {
	Index int    `json:"index"`
	Name  string `json:"name"`
}

// Metric represents a snap metric collected or to be collected
type Metric interface {
	RequestedMetric
	Config() *cdata.ConfigDataNode
	LastAdvertisedTime() time.Time
	Data() interface{}
	Source() string
	Labels() []Label
	Tags() map[string]string
	Timestamp() time.Time
}

// RequestedMetric is a metric requested for collection
type RequestedMetric interface {
	Namespace() []string
	Version() int
}

type CatalogedMetric interface {
	RequestedMetric
	LastAdvertisedTime() time.Time
	Policy() *cpolicy.ConfigPolicyNode
}

func JoinNamespace(ns []string) string {
	return "/" + strings.Join(ns, "/")
}
