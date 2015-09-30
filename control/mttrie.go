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

package control

import (
	"errors"
	"fmt"

	"github.com/intelsdi-x/pulse/core"
	"github.com/intelsdi-x/pulse/core/perror"
)

/*
Given a trie like this:

        root
         /\
        /  \
      foo  bar
     /\      /\
    /  \    /  \
   a    b  c    d

The result of a collect query like so: Fetch([]string{"root", "foo"})
would return a slice of the MetricTypes found in nodes a & b.
Get collects all children of a given node and returns the values
in all leaves.

This query is needed primarily for the REST interface, where it
can be used to make efficient lookups of Metric Types in a RESTful
manner:

FETCH /metric/root/foo -> trie.Fetch([]string{"root", "foo"}) ->
    [a,b]

*/

// ErrNotFound is returned when Get cannot find the given namespace
var ErrNotFound = errors.New("metric not found")

type mttNode struct {
	children map[string]*mttNode
	mts      map[int]*metricType
}

// The root in the trie
type MTTrie struct {
	*mttNode
}

// New() returns an empty trie
func NewMTTrie() *MTTrie {
	m := &mttNode{
		children: map[string]*mttNode{},
	}
	return &MTTrie{m}
}

// Handy print out of the tr(i)e
func (m *MTTrie) String() string {
	out := ""
	for _, mt := range m.gatherMetricTypes() {
		out += fmt.Sprintf("%s => %s\n", mt.Key(), mt.Plugin.Key())
	}
	return out
}

func (m *MTTrie) gatherMetricTypes() []metricType {
	mts := make([]metricType, 0)
	children := make([]*mttNode, 0)
	for _, node := range m.children {
		children = gatherChildren(children, node)
	}
	for _, c := range children {
		for _, mt := range c.mts {
			mts = append(mts, *mt)
		}
	}
	return mts
}

// Remove all metrics from the catalog if they match a loadedPlugin
func (m *MTTrie) DeleteByPlugin(lp *loadedPlugin) {
	for _, mt := range m.gatherMetricTypes() {
		if mt.Plugin.Key() == lp.Key() {
			// Remove this metric
			m.RemoveMetric(mt)
		}
	}
}

// Removes a specific metric by namespace and version from the tree
func (m *MTTrie) RemoveMetric(mt metricType) {
	a, _ := m.find(mt.Namespace())
	if a != nil {
		for v, x := range a.mts {
			if mt.Version() == x.Version() {
				// Delete the metric from the node
				delete(a.mts, v)
			}
		}
	}
}

// Add adds a node with the given namespace with the
// given MetricType
func (mtt *mttNode) Add(mt *metricType) {
	ns := mt.Namespace()
	node, index := mtt.walk(ns)
	if index == len(ns) {
		if node.mts == nil {
			node.mts = make(map[int]*metricType)
		}
		node.mts[mt.Version()] = mt
		return
	}
	// walk through the remaining namespace and build out the
	// new branch in the trie.
	for _, n := range ns[index:] {
		if node.children == nil {
			node.children = make(map[string]*mttNode)
		}
		node.children[n] = &mttNode{}
		node = node.children[n]
	}
	node.mts = make(map[int]*metricType)
	node.mts[mt.Version()] = mt
}

// Collect collects all children below a given namespace
// and concatenates their metric types into a single slice
func (mtt *mttNode) Fetch(ns []string) ([]*metricType, perror.PulseError) {
	node, err := mtt.find(ns)
	if err != nil {
		return nil, err
	}

	var children []*mttNode
	if node.mts != nil {
		children = append(children, node)
	}
	if node.children != nil {
		children = gatherChildren(children, node)
	}

	var mts []*metricType
	for _, child := range children {
		for _, mt := range child.mts {
			mts = append(mts, mt)
		}
	}

	return mts, nil
}

// Remove removes all children below a given namespace
func (mtt *mttNode) Remove(ns []string) perror.PulseError {
	_, err := mtt.find(ns)
	if err != nil {
		return err
	}

	//remove node from parent
	parent, err := mtt.find(ns[:len(ns)-1])
	if err != nil {
		return err
	}
	delete(parent.children, ns[len(ns)-1:][0])

	return nil
}

// Get works like fetch, but only returns the MT at the given node
// and does not gather the node's children.
func (mtt *mttNode) Get(ns []string) ([]*metricType, perror.PulseError) {
	node, err := mtt.find(ns)
	if err != nil {
		return nil, err
	}
	if node.mts == nil {
		pe := perror.New(errorMetricNotFound(ns))
		pe.SetFields(map[string]interface{}{
			"name": core.JoinNamespace(ns),
		})
		return nil, pe
	}
	var mts []*metricType
	for _, mt := range node.mts {
		mts = append(mts, mt)
	}
	return mts, nil
}

// walk returns the last leaf / branch present
// in the trie and the index in the namespace that the last node exists.
func (mtt *mttNode) walk(ns []string) (*mttNode, int) {
	parent := mtt
	var pp *mttNode
	for i, n := range ns {
		if parent.children == nil {
			return parent, i
		}
		pp = parent
		parent = parent.children[n]
		if parent == nil {
			return pp, i
		}
	}
	return parent, len(ns)
}

func (mtt *mttNode) find(ns []string) (*mttNode, perror.PulseError) {
	node, index := mtt.walk(ns)
	if index != len(ns) {
		pe := perror.New(errorMetricNotFound(ns))
		pe.SetFields(map[string]interface{}{
			"name": core.JoinNamespace(ns),
		})
		return nil, pe
	}
	return node, nil
}

func gatherChildren(children []*mttNode, node *mttNode) []*mttNode {
	for _, child := range node.children {
		if child.children != nil {
			children = gatherChildren(children, child)
		}
		children = append(children, child)
	}
	return children
}
