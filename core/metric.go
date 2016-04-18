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

type Namespace []NamespaceElement

func (n Namespace) Strings() []string {
	st := make([]string, len(n))
	for i, j := range n {
		st[i] = j.Value
	}
	return st
}

type NamespaceElement struct {
	Value       string
	Description string
	Name        string
}

func NewNamespaceElement(e string) NamespaceElement {
	return NamespaceElement{Value: e}
}

func NewNamespace(ns []string) Namespace {
	n := make([]NamespaceElement, len(ns))
	for i, ns := range ns {
		n[i] = NamespaceElement{Value: ns}
	}
	return n
}

// RequestedMetric is a metric requested for collection
type RequestedMetric interface {
	Namespace() Namespace
	Version() int
}

type CatalogedMetric interface {
	RequestedMetric
	LastAdvertisedTime() time.Time
	Policy() *cpolicy.ConfigPolicyNode
}

func JoinNamespace(ns Namespace) string {
	n := ""
	for i, x := range ns {
		if i == 0 {
			n = "/"
		}
		n += x.Value
		if i != len(ns)-1 {
			n += "/"
		}
	}
	return n
}

func GenerateKey(ns Namespace) string {
	n := ""
	for i, x := range ns {
		n += x.Value
		if i < len(ns)-1 {
			n += "."
		}
	}
	return n
}
