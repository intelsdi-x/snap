package core

import (
	"sync"

	"github.com/intelsdilabs/pulse/core/ctypes"
	"github.com/intelsdilabs/pulse/pkg/ctree"
)

// Allows adding of config data by namespace and retrieving of data from tree
// at a specific namespace (merging the relevant hiearchy). Uses pkg.ConfigTree.
type ConfigDataTree struct {
	cTree *ctree.ConfigTree
}

// Returns a new ConfigDataTree.
func NewTree() *ConfigDataTree {
	return &ConfigDataTree{
		cTree: ctree.New(),
	}
}

// Adds a ConfigDataNode at the provided namespace.
func (c *ConfigDataTree) Add(ns []string, cdn *ConfigDataNode) {
	c.cTree.Add(ns, cdn)
}

// Returns a ConfigDataNode that is a merged version of the namespace provided.
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

// Freezes the ConfigDataTree from future writes (adds) and triggers compression
// of tree into read-performant version.
func (c *ConfigDataTree) Freeze() {
	c.cTree.Freeze()
}

// Represents a set of configuration data
type ConfigDataNode struct {
	mutex *sync.Mutex
	table map[string]ctypes.ConfigValue
}

// Returns a new and empty node.
func NewNode() *ConfigDataNode {
	return &ConfigDataNode{
		mutex: new(sync.Mutex),
		table: make(map[string]ctypes.ConfigValue),
	}
}

// Returns the table of configuration items [key(string) / value(core.ConfigValue)].
func (c *ConfigDataNode) Table() map[string]ctypes.ConfigValue {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	return c.table
}

// Adds an item to the ConfigDataNode.
func (c *ConfigDataNode) AddItem(k string, v ctypes.ConfigValue) {
	// And empty is a noop
	if k == "" {
		return
	}
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.table[k] = v
}

// Merges a ConfigDataNode on top of this one (overwriting items where it occurs).
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
