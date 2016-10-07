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
	"strings"
	"time"

	"github.com/intelsdi-x/snap/control/plugin/cpolicy"
	"github.com/intelsdi-x/snap/core/cdata"
)

var (
	// Standard Tags are in added to the metric by the framework on plugin load.
	// STD_TAG_PLUGIN_RUNNING_ON describes where the plugin is running (hostname).
	STD_TAG_PLUGIN_RUNNING_ON = "plugin_running_on"
	nsPriorityList            = []string{"/", "|", "%", ":", "-", ";", "_", "^", ">", "<", "+", "=", "&", "㊽", "Ä", "大", "小", "ᵹ", "☍", "ヒ"}
)

// Metric represents a snap metric collected or to be collected
type Metric interface {
	RequestedMetric
	Config() *cdata.ConfigDataNode
	LastAdvertisedTime() time.Time
	Data() interface{}
	Tags() map[string]string
	Timestamp() time.Time
	Description() string
	Unit() string
}

type Namespace []NamespaceElement

// String returns the string representation of the namespace with "/" joining
// the elements of the namespace.  A leading "/" is added.
func (n Namespace) String() string {
	ns := n.Strings()
	s := n.getSeparator()
	return s + strings.Join(ns, s)
}

// Strings returns an array of strings that represent the elements of the
// namespace.
func (n Namespace) Strings() []string {
	var ns []string
	for _, namespaceElement := range n {
		ns = append(ns, namespaceElement.Value)
	}
	return ns
}

// getSeparator returns the highest suitable separator from the nsPriorityList.
// Otherwise the core separator is returned.
func (n Namespace) getSeparator() string {
	smap := initSeparatorMap()

	for _, e := range n {
		// look at each char
		for _, r := range e.Value {
			ch := fmt.Sprintf("%c", r)
			if v, ok := smap[ch]; ok && !v {
				smap[ch] = true
			}
		}
	}

	// Go through our separator list
	for _, s := range nsPriorityList {
		if v, ok := smap[s]; ok && !v {
			return s
		}
	}
	return Separator
}

// initSeparatorMap populates the local map of nsPriorityList.
func initSeparatorMap() map[string]bool {
	m := map[string]bool{}

	for _, s := range nsPriorityList {
		m[s] = false
	}
	return m
}

// IsDynamic returns true if there is any element of the namespace which is
// dynamic.  If the namespace is dynamic the second return value will contain
// an array of namespace elements (indexes) where there are dynamic namespace
// elements. A dynamic component of the namespace are those elements that
// contain variable data.
func (n Namespace) IsDynamic() (bool, []int) {
	var idx []int
	ret := false
	for i := range n {
		if n[i].IsDynamic() {
			ret = true
			idx = append(idx, i)
		}
	}
	return ret, idx
}

// NewNamespace takes an array of strings and returns a Namespace.  A Namespace
// is an array of NamespaceElements.  The provided array of strings is used to
// set the corresponding Value fields in the array of NamespaceElements.
func NewNamespace(ns ...string) Namespace {
	n := make([]NamespaceElement, len(ns))
	for i, ns := range ns {
		n[i] = NamespaceElement{Value: ns}
	}
	return n
}

// AddDynamicElement adds a dynamic element to the given Namespace.  A dynamic
// NamespaceElement is defined by having a nonempty Name field.
func (n Namespace) AddDynamicElement(name, description string) Namespace {
	nse := NamespaceElement{Name: name, Description: description, Value: "*"}
	return append(n, nse)
}

// AddStaticElement adds a static element to the given Namespace.  A static
// NamespaceElement is defined by having an empty Name field.
func (n Namespace) AddStaticElement(value string) Namespace {
	nse := NamespaceElement{Value: value}
	return append(n, nse)
}

// AddStaticElements adds a static elements to the given Namespace.  A static
// NamespaceElement is defined by having an empty Name field.
func (n Namespace) AddStaticElements(values ...string) Namespace {
	for _, value := range values {
		n = append(n, NamespaceElement{Value: value})
	}
	return n
}

func (n Namespace) Element(idx int) NamespaceElement {
	if idx >= 0 && idx < len(n) {
		return n[idx]
	}
	return NamespaceElement{}
}

// NamespaceElement provides meta data related to the namespace.  This is of particular importance when
// the namespace contains data.
type NamespaceElement struct {
	Value       string
	Description string
	Name        string
}

// NewNamespaceElement tasks a string and returns a NamespaceElement where the
// Value field is set to the provided string argument.
func NewNamespaceElement(e string) NamespaceElement {
	if e != "" {
		return NamespaceElement{Value: e}
	}
	return NamespaceElement{}
}

// IsDynamic returns true if the namespace element contains data.  A namespace
// element that has a nonempty Name field is considered dynamic.
func (n *NamespaceElement) IsDynamic() bool {
	if n.Name != "" {
		return true
	}
	return false
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
	Description() string
	Unit() string
}
