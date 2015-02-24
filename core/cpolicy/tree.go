package cpolicy

import (
	"github.com/intelsdilabs/pulse/pkg/ctree"
)

// Allows adding of config policy by namespace and retrieving of policy from a tree
// at a specific namespace (merging the relevant hiearchy). Uses pkg.ConfigTree.
type ConfigPolicyTree struct {
	cTree *ctree.ConfigTree
}

// Returns a new ConfigDataTree.
func NewTree() *ConfigPolicyTree {
	return &ConfigPolicyTree{
		cTree: ctree.New(),
	}
}

// Adds a ConfigDataNode at the provided namespace.
func (c *ConfigPolicyTree) Add(ns []string, cpn *ConfigPolicyNode) {
	c.cTree.Add(ns, cpn)
}

// Returns a ConfigDataNode that is a merged version of the namespace provided.
func (c *ConfigPolicyTree) Get(ns []string) *ConfigPolicyNode {
	// Automatically freeze on first Get
	if !c.cTree.Frozen() {
		c.cTree.Freeze()
	}

	n := c.cTree.Get(ns)
	if n == nil {
		return nil
	} else {
		switch t := n.(type) {
		case ConfigPolicyNode:
			return &t
		default:
			return t.(*ConfigPolicyNode)
		}
	}
}

// Freezes the ConfigDataTree from future writes (adds) and triggers compression
// of tree into read-performant version.
func (c *ConfigPolicyTree) Freeze() {
	c.cTree.Freeze()
}
