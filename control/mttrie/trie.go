package mttrie

import (
	"errors"

	"github.com/intelsdilabs/pulse/core"
)

var ErrNotFound = errors.New("namespace not found in trie")

type mttNode struct {
	children map[string]*mttNode
	mt       core.MetricType
}

type MTTrie struct {
	*mttNode
}

func New() *MTTrie {
	m := &mttNode{
		children: map[string]*mttNode{},
	}
	return &MTTrie{m}
}

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
