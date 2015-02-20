package ctree

import (
	"sync"
)

type ConfigTree struct {
	root  *node
	mutex *sync.Mutex
}

func New() *ConfigTree {
	return &ConfigTree{
		root:  &node{},
		mutex: &sync.Mutex{},
	}
}

func (ct *ConfigTree) Add(ns []string) {
	f, remain := ns[0], ns[1:]
	if ct.root == nil {
		ct.root = &node{
			key: f,
		}
		ct.root.add(remain)
	} else {
	}
}

//func (ct *ConfigTree) Get() Node {}

type Node interface {
}

type node struct {
	nodes []node
	key   string
	Node  Node
}

func (n *node) add(ns []string) {
	f, remain := ns[0], ns[1:]
	for _, nd := range n.nodes {
		if f == nd.key {
			nd.add(remain)
			return
		}
	}
	newNode := node{
		key: f,
	}
	newNode.add(remain)
	n.nodes = append(n.nodes, newNode)
}
