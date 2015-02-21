package core

import (
	"github.com/intelsdilabs/pulse/pkg/ctree"
)

type ConfigDataTree struct {
	cTree *ctree.ConfigTree
}

func NewConfigDataTree() *ConfigDataTree {
	return &ConfigDataTree{cTree: ctree.New()}
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
	Value string
}

func (c ConfigDataNode) Merge(n ctree.Node) ctree.Node {
	// Because Add only allows the ConfigDataNode type we
	// are safe to convert ctree.Node interface to ConfigDataNode
	// cd := n.(ConfigDataNode)
	c.Value = c.Value + "/" + n.(*ConfigDataNode).Value
	return c
}
