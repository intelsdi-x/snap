package core

import (
	"sync"

	"github.com/intelsdilabs/pulse/pkg/ctree"
)

type ConfigDataTree struct {
	cTree *ctree.ConfigTree
}

func NewConfigDataTree() *ConfigDataTree {
	return &ConfigDataTree{
		cTree: ctree.New(),
	}
}

func (c *ConfigDataTree) Add(ns []string, cdn *ConfigDataNode) {
	c.cTree.Add(ns, cdn)
}

func (c *ConfigDataTree) Get(ns []string) *ConfigDataNode {
	// Automatically freeze on first Get
	if !c.cTree.Frozen() {
		c.cTree.Freeze()
	}

	n := c.cTree.Get(ns)
	if n == nil {
		return nil
	} else {
		cd := n.(ConfigDataNode)
		return &cd
	}
}

func (c *ConfigDataTree) Freeze() {
	c.cTree.Freeze()
}

type ConfigDataNode struct {
	mutex *sync.Mutex
	table map[string]ConfigValue
}

func NewConfigDataNode() *ConfigDataNode {
	return &ConfigDataNode{
		mutex: new(sync.Mutex),
		table: make(map[string]ConfigValue),
	}
}

func (c *ConfigDataNode) Table() map[string]ConfigValue {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	return c.table
}

func (c *ConfigDataNode) AddItem(k string, v ConfigValue) {
	// And empty is a noop
	if k == "" {
		return
	}
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.table[k] = v
}

func (c ConfigDataNode) Merge(n ctree.Node) ctree.Node {
	// Because Add only allows the ConfigDataNode type we
	// are safe to convert ctree.Node interface to ConfigDataNode
	cd := n.(*ConfigDataNode)
	t := cd.Table()
	// For the table in the passed ConfigDataNode(converted) add each item to
	// this ConfigDataNode overwritting where needed.
	for k, v := range t {
		c.AddItem(k, v)
	}
	// Return modified version of ConfigDataNode(as ctree.Node)
	return c
}
