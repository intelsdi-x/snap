package cdata

import (
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
