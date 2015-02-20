package ctree

import (
	"fmt"
	"strings"
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
		fmt.Println("\nNew node")
		c.root = &node{
			keys: []string{f},
		}
	} else {
		fmt.Println("Existing node")
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

func (c *ConfigTree) Get(ns []string) *[]Node {
	retNodes := new([]Node)

	if c.root == nil {
		return retNodes
	}

	rootKeyLength := len(c.root.keys)

	if len(ns) < rootKeyLength {
		return retNodes
	}

	match, remain := ns[:rootKeyLength], ns[rootKeyLength:]

	fmt.Println(rootKeyLength)
	fmt.Println(match, remain)

	if strings.Join(match, "/") == c.root.keysString {
		for _, child := c.root.nodes {
			childNodes := child.get(remain)
			if len(childNodes) > 0 {
				retNodes = append(retNodes, childNodes...)
			}
		}
	}

	return retNodes
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
}

// REMOVE ME
type DummyNode struct {
}

type node struct {
	nodes      []*node
	keys       []string
	keysString string
	Node       Node
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
		fmt.Printf("leaf: %s\n", n.keys[0])
		n.Node = inNode
		// fmt.Printf("inNode: %v\n", n.Node)
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
			fmt.Printf("MERGE! %s <= %s\n", n.keys, n.nodes[0].keys)
			n.keys = append(n.keys, n.nodes[0].keys...)
			n.keysString = strings.Join(n.keys, "/")

			n.Node = n.nodes[0].Node
			n.nodes = n.nodes[0].nodes

			fmt.Printf("now %s\n", n.keys)

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
