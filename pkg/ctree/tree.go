package ctree

import (
	"bytes"
	"fmt"
	"sync"
)

type ConfigTree struct {
	freezeFlag bool
	root       *node
	mutex      *sync.Mutex
}

func New() *ConfigTree {
	return &ConfigTree{
		mutex: &sync.Mutex{},
	}
}

func (c *ConfigTree) Add(ns []string, inNode Node) {
	c.mutex.Lock()
	f, remain := ns[0], ns[1:]
	if c.root == nil {
		c.root = &node{
			keys: []string{f},
		}
	} else {
		if f != c.root.keys[0] {
			panic("Can't add a new root namespace")
		}
	}
	if len(ns) == 0 {
		c.root.Node = inNode
		return
	}
	c.root.add(remain, inNode)
	c.mutex.Unlock()
}

func (c *ConfigTree) Get(ns []string) Node {
	retNodes := new([]Node)
	// Return if no
	if c.root == nil {
		return nil
	}

	rootKeyLength := len(c.root.keys)

	if len(ns) < rootKeyLength {
		return nil
	}

	match, remain := ns[:rootKeyLength], ns[rootKeyLength:]
	if bytes.Compare(nsToByteArray(match), c.root.keysBytes) != 0 {
		return nil
	}

	if c.root.Node != nil {
		*retNodes = append(*retNodes, c.root.Node)
	}
	for _, child := range c.root.nodes {
		childNodes := child.get(remain)
		*retNodes = append(*retNodes, *childNodes...)
	}

	// Call Node.Merge() sequentially on the retNodes
	rn := (*retNodes)[0]
	for _, n := range (*retNodes)[1:] {
		rn.Merge(n)
	}
	return rn
}

func (c *ConfigTree) Freeze() {
	c.mutex.Lock()
	if !c.freezeFlag {
		c.freezeFlag = true
		c.compact()
	}
	c.mutex.Unlock()
}

func (c *ConfigTree) Frozen() bool {
	return c.freezeFlag
}

func (c *ConfigTree) compact() {
	if c.root != nil {
		c.root.compact()
	}
}

func (c *ConfigTree) print() {
	c.root.print("")
}

type Node interface {
	Merge(Node)
}

type node struct {
	nodes     []*node
	keys      []string
	keysBytes []byte
	Node      Node
}

func (n *node) print(p string) {
	s := fmt.Sprintf("%s/%s(%v)", p, n.keys, n.Node != nil)
	fmt.Println(s)
	for _, nd := range n.nodes {
		nd.print(s)
	}
}

func (n *node) add(ns []string, inNode Node) {
	if len(ns) == 0 {
		n.Node = inNode
		return
	}
	f, remain := ns[0], ns[1:]
	for _, nd := range n.nodes {
		if f == nd.keys[0] {
			nd.add(remain, inNode)
			return
		}
	}
	newNode := &node{
		keys: []string{f},
	}
	newNode.add(remain, inNode)
	n.nodes = append(n.nodes, newNode)
}

func (n *node) compact() {
	// Eval if we can merge with single child
	//
	// Only try compact if we have a single child
	if len(n.nodes) == 1 {
		if n.empty() {
			// merge single child into this node
			n.keys = append(n.keys, n.nodes[0].keys...)

			n.keysBytes = nsToByteArray(n.keys)

			n.Node = n.nodes[0].Node
			n.nodes = n.nodes[0].nodes

			n.compact()
			return
		}
	}

	// Call compact on any children
	for _, child := range n.nodes {
		child.compact()
	}
}

func (n *node) empty() bool {
	return n.Node == nil
}

func (n *node) get(ns []string) *[]Node {
	retNodes := new([]Node)

	rootKeyLength := len(n.keys)
	if len(ns) < rootKeyLength {
		return retNodes
	}

	match, remain := ns[:rootKeyLength], ns[rootKeyLength:]
	if bytes.Compare(nsToByteArray(match), n.keysBytes) == 0 {
		// If Node is present add to the return Nodes
		if !n.empty() {
			*retNodes = append(*retNodes, n.Node)
		}

		// For any existing children call get
		for _, child := range n.nodes {
			childNodes := child.get(remain)
			*retNodes = append(*retNodes, *childNodes...)
		}
	}

	return retNodes
}

func nsToByteArray(str []string) []byte {
	b := []byte{}
	for _, s := range str {
		b = append(b, []byte(s)...)
	}
	return b
}
