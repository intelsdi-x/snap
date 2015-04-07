package mttrie

import (
	"errors"

	"github.com/intelsdilabs/pulse/core"
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

The result of a Get query like so: Get([]{"root", "foo"})
would return a slice of the MetricTypes found in nodes a & b.
Get collects all children of a given node and returns the values
in all leaves.

This query is needed primarily for the REST interface, where it
can be used to make efficient lookups of Metric Types in a RESTful
manner:

GET /metric/root/foo -> trie.Get([]string{"root", "foo"}) ->
    [a,b]

*/

// ErrNotFound is returned when Get cannot find the given namespace
var ErrNotFound = errors.New("namespace not found in trie")

type mttNode struct {
	children map[string]*mttNode
	mt       core.MetricType
}

// The root in the trie
type MTTrie struct {
	*mttNode
}

// New() returns an empty trie
func New() *MTTrie {
	m := &mttNode{
		children: map[string]*mttNode{},
	}
	return &MTTrie{m}
}

// Add adds a node with the given namespace with the
// given MetricType
func (mtt *mttNode) Add(ns []string, mt core.MetricType) {
	node, index := mtt.walk(ns)
	if index == len(ns) {
		node.mt = mt
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
	node.mt = mt
}

// Get collects all children below a given namespace
// and concatenates their metric types into a single slice
func (mtt *mttNode) Get(ns []string) ([]core.MetricType, error) {
	node, index := mtt.walk(ns)
	if index != len(ns) {
		return nil, ErrNotFound
	}

	var c []*mttNode
	if node.children == nil {
		c = append(c, node)
	} else {
		c = gatherChildren(c, node)
	}

	mts := make([]core.MetricType, len(c))
	for i, cc := range c {
		mts[i] = cc.mt
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

func gatherChildren(children []*mttNode, node *mttNode) []*mttNode {
	for _, child := range node.children {
		if child.children != nil {
			children = gatherChildren(children, child)
			continue
		}
		children = append(children, child)
	}
	return children
}
