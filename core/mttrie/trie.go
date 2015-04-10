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
	mts      map[int]core.MetricType
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
		if node.mts == nil {
			node.mts = make(map[int]core.MetricType)
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
	node.mts = make(map[int]core.MetricType)
	node.mts[mt.Version()] = mt
}

// Collect collects all children below a given namespace
// and concatenates their metric types into a single slice
func (mtt *mttNode) Fetch(ns []string) ([]core.MetricType, error) {
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

	var mts []core.MetricType
	for _, child := range children {
		for _, mt := range child.mts {
			mts = append(mts, mt)
		}
	}

	return mts, nil
}

// Remove removes all children below a given namespace
func (mtt *mttNode) Remove(ns []string) error {
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
func (mtt *mttNode) Get(ns []string) ([]core.MetricType, error) {
	node, err := mtt.find(ns)
	if err != nil {
		return nil, err
	}
	if node.mts == nil {
		return nil, ErrNotFound
	}
	var mts []core.MetricType
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

func (mtt *mttNode) find(ns []string) (*mttNode, error) {
	node, index := mtt.walk(ns)
	if index != len(ns) {
		return nil, ErrNotFound
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
