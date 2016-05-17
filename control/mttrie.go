/*
http://www.apache.org/licenses/LICENSE-2.0.txt


Copyright 2015-2016 Intel Corporation

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
	"fmt"
	"sort"
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

type mttNode struct {
	children map[string]*mttNode
	mts      map[int]*metricType
}

// MTTrie struct representing the root in the trie
type MTTrie struct {
	*mttNode
}

// NewMTTrie returns an empty trie
func NewMTTrie() *MTTrie {
	m := &mttNode{
		children: map[string]*mttNode{},
	}
	return &MTTrie{m}
}

// String prints out of the tr(i)e
func (m *MTTrie) String() string {
	out := ""
	for _, mt := range m.gatherMetricTypes() {
		out += fmt.Sprintf("%s => %s\n", mt.Key(), mt.Plugin.Key())
	}
	return out
}

func (m *MTTrie) gatherMetricTypes() []metricType {
	var mts []metricType
	var descendants []*mttNode
	for _, node := range m.children {
		descendants = gatherDescendants(descendants, node)
	}
	for _, c := range descendants {
		for _, mt := range c.mts {
			mts = append(mts, *mt)
		}
	}
	return mts
}

// DeleteByPlugin removes all metrics from the catalog if they match a loadedPlugin
func (m *MTTrie) DeleteByPlugin(lp *loadedPlugin) {
	for _, mt := range m.gatherMetricTypes() {
		if mt.Plugin.Key() == lp.Key() {
			// Remove this metric
			m.RemoveMetric(mt)
		}
	}
}

// RemoveMetric removes a specific metric by namespace and version from the tree
func (m *MTTrie) RemoveMetric(mt metricType) {
	a, _ := m.find(mt.Namespace().Strings())
	if a != nil {
		for v, x := range a.mts {
			if mt.Version() == x.Version() {
				// delete this metric from the node
				delete(a.mts, v)
			}
		}
	}
}

// Add adds a node with the given namespace with the given MetricType
func (mtt *mttNode) Add(mt *metricType) {
	ns := mt.Namespace()
	node, index := mtt.walk(ns.Strings())
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
		node.children[n.Value] = &mttNode{}
		node = node.children[n.Value]
	}
	node.mts = make(map[int]*metricType)
	node.mts[mt.Version()] = mt
}

// Fetch collects all children below a given namespace
// and concatenates their metric types into a single slice
func (mtt *mttNode) Fetch(ns []string) ([]*metricType, error) {
	children := mtt.fetch(ns)
	var mts []*metricType
	for _, child := range children {
		for _, mt := range child.mts {
			mts = append(mts, mt)
		}
	}
	if len(mts) == 0 {
		return nil, errorFetchMetricsNotFound(ns)
	}
	return mts, nil
}

// Remove removes all descendants nodes below a given namespace
func (mtt *mttNode) Remove(ns []string) error {
	_, err := mtt.find(ns)
	if err != nil {
		return err
	}

	parent, err := mtt.find(ns[:len(ns)-1])
	if err != nil {
		return err
	}

	delete(parent.children, ns[len(ns)-1:][0])

	return nil
}

// GetMetric works like fetch, but only returns the MT at the given node
// in the queried version (or in the latest if ver < 1) and does NOT gather the node's children.
func (mtt *mttNode) GetMetric(ns []string, ver int) (*metricType, error) {
	node, err := mtt.find(ns)
	if err != nil {
		return nil, err
	}

	mt, err := getVersion(node.mts, ver)
	if err != nil {
		return nil, errorMetricNotFound(ns, ver)
	}

	return mt, nil
}

// GetMetrics returns all MTs at the given namespace in the queried version (or in the latest if ver < 1)
// and does gather all the node's descendants if the namespace ends with an asterisk
func (mtt *mttNode) GetMetrics(ns []string, ver int) ([]*metricType, error) {
	nodes := []*mttNode{}
	mts := []*metricType{}

	// search returns all of the nodes fulfilling the 'ns'
	// even for some of them there is no metric (empty node.mts)
	nodes = mtt.search(nodes, ns)

	for _, node := range nodes {
		// choose the queried version of metric types (or the latest if ver < 1)
		// and concatenate them into a single slice
		mt, err := getVersion(node.mts, ver)

		if err != nil {
			continue
		}

		mts = append(mts, mt)
	}

	if len(mts) == 0 {
		return nil, errorMetricNotFound(ns, ver)
	}

	return mts, nil
}

// GetVersions returns all versions of MTs below the given namespace
func (mtt *mttNode) GetVersions(ns []string) ([]*metricType, error) {
	var nodes []*mttNode
	var mts []*metricType

	nodes = mtt.search(nodes, ns)

	for _, node := range nodes {
		// concatenates metric types in ALL versions into a single slice
		for _, mt := range node.mts {
			mts = append(mts, mt)
		}
	}

	if len(mts) == 0 {
		return nil, errorMetricNotFound(ns)
	}

	return mts, nil
}

// fetch collects all descendants nodes below the given namespace
func (mtt *mttNode) fetch(ns []string) []*mttNode {
	node, err := mtt.find(ns)
	if err != nil {
		return nil
	}

	var children []*mttNode
	if node.mts != nil {
		children = append(children, node)
	}
	if node.children != nil {
		children = gatherDescendants(children, node)
	}

	return children
}

// walk returns the last leaf / branch present in the trie and the index in the namespace that the last node exists.
// It is useful e.g. to locate the right place to add new metric type into tree with the given namespace
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

// search returns leaf nodes in the trie below the given namespace
func (mtt *mttNode) search(nodes []*mttNode, ns []string) []*mttNode {
	parent := mtt
	var children []*mttNode

	if parent.children == nil {
		return nodes
	}

	if len(ns) == 1 {
		// the last element of ns is under searching process

		switch ns[0] {
		case "*":
			// fetch all descendants when wildcard ends namespace
			children = parent.fetch([]string{})

		default:
			children = parent.gatherChildren(ns[0])
		}

		nodes = append(nodes, children...)
		return nodes
	}

	children = parent.gatherChildren(ns[0])

	for _, child := range children {
		nodes = child.search(nodes, ns[1:])
	}

	return nodes
}

func (mtt *mttNode) find(ns []string) (*mttNode, error) {
	node, index := mtt.walk(ns)
	if index != len(ns) {
		return nil, errorMetricNotFound(ns)
	}
	return node, nil
}

// gatherChildren returns child or children by the 'name' of a given node (direct descendant(s))
// and concatenates this direct descendant(s) into a single slice
func (mtt *mttNode) gatherChildren(name string) []*mttNode {
	var children []*mttNode
	switch name {
	case "*":
		// name of child is unspecified, so gather all children
		for _, child := range mtt.children {
			children = append(children, child)
		}
	default:
		// gather a single child with specified name
		child := mtt.children[name]

		if child == nil {
			// child with this name not exist; it might be specific instance of dynamic metric
			// so, take child named with an asterisk
			child = mtt.children["*"]

		}

		if child != nil {
			children = append(children, child)
		}

	}
	return children
}

// gatherDescendants returns all descendants of a given node
func gatherDescendants(descendants []*mttNode, node *mttNode) []*mttNode {
	for _, child := range node.children {

		if child.mts != nil {
			descendants = append(descendants, child)
		}

		if child.children != nil {
			descendants = gatherDescendants(descendants, child)
		}

	}
	return descendants
}

// getVersion returns the MT in the latest version
func getLatest(mts map[int]*metricType) *metricType {
	versions := []int{}

	// version is a key in mts map
	for ver := range mts {
		// concatenates all available versions to a single slice
		versions = append(versions, ver)
	}

	// sort and take the last element (the latest version)
	sort.Ints(versions)
	latestVersion := versions[len(versions)-1]

	return mts[latestVersion]
}

// getVersion returns the MT in the queried version (or the latest if 'ver' < 1)
func getVersion(mts map[int]*metricType, ver int) (*metricType, error) {
	if len(mts) == 0 {
		return nil, errMetricNotFound
	}

	if ver > 0 {
		// a version IS given
		if mt, exist := mts[ver]; exist {
			return mt, nil
		}
		return nil, errMetricNotFound
	}

	// ver is less than or equal to 0 get the latest
	return getLatest(mts), nil
}
