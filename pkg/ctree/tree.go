/*
http://www.apache.org/licenses/LICENSE-2.0.txt


Copyright 2015 Intel Corporation

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package ctree

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"strings"

	log "github.com/Sirupsen/logrus"
)

// ConfigTree struct type
type ConfigTree struct {
	// Debug turns on verbose logging of the tree functions to stdout
	Debug bool

	freezeFlag bool
	root       *node
}

// New returns a new instance of ConfigTree
func New() *ConfigTree {
	return &ConfigTree{}
}

func (c *ConfigTree) log(s string) {
	if c.Debug {
		log.Print(s)
	}
}

// GobEncode returns the encoded ConfigTree. Otherwise,
// an error is returned
func (c *ConfigTree) GobEncode() ([]byte, error) {
	//todo throw an error if not frozen
	w := new(bytes.Buffer)
	encoder := gob.NewEncoder(w)
	if c.root == nil {
		c.root = &node{}
		// c.root.setKeys([]string{})
	}
	if err := encoder.Encode(c.root); err != nil {
		return nil, err
	}
	if err := encoder.Encode(c.freezeFlag); err != nil {
		return nil, err
	}
	return w.Bytes(), nil
}

// GobDecode decodes the ConfigTree.
func (c *ConfigTree) GobDecode(buf []byte) error {
	r := bytes.NewBuffer(buf)
	decoder := gob.NewDecoder(r)
	if err := decoder.Decode(&c.root); err != nil {
		return err
	}
	return decoder.Decode(&c.freezeFlag)
}

// MarshalJSON marshals ConfigTree
func (c *ConfigTree) MarshalJSON() ([]byte, error) {
	return json.Marshal(&struct {
		Root *node `json:"root"`
	}{
		Root: c.root,
	})
}

// Add adds a new tree node
func (c *ConfigTree) Add(ns []string, inNode Node) {
	c.log(fmt.Sprintf("Adding %v at %s\n", inNode, ns))
	if len(ns) == 0 {
		c.log(fmt.Sprintln("ns is empty - returning with no change to tree"))
		return
	}
	f, remain := ns[0], ns[1:]
	c.log(fmt.Sprintf("first ns (%s) remain (%s)", f, remain))
	if c.root == nil {
		// Create node at root
		c.root = new(node)
		c.root.setKeys([]string{f})

		c.log(fmt.Sprintf("Root now = %v\n", c.root.keys))

		// If remain is empty then the inNode belongs at this root level
		if len(remain) == 0 {
			c.log(fmt.Sprintf("adding node at root level\n"))
			c.root.Node = inNode
			// And return since we are done
			return
		}

	} else {
		if f != c.root.keys[0] {
			panic("Can't add a new root namespace")
		}
	}
	c.root.add(remain, inNode)

}

func (c *ConfigTree) GetAll() map[string]Node {
	ret := map[string]Node{}
	if !c.Frozen() {
		panic("must freeze before getting")
	}
	if c.root == nil {
		c.log(fmt.Sprintln("ctree: no root - returning nil"))
		return nil
	}
	return c.getAll(c.root, "", ret)
}

func (c *ConfigTree) getAll(node *node, base string, results map[string]Node) map[string]Node {
	if len(node.keys) > 0 {
		if base != "" {
			base = base + "." + strings.Join(node.keys, ".")
		} else {
			base = strings.Join(node.keys, ".")
		}
		if node.Node != nil {
			results[base] = node.Node
		}
	}
	for _, child := range node.nodes {
		c.getAll(child, base, results)
	}
	return results
}

// Get returns a tree node given the namespace
func (c *ConfigTree) Get(ns []string) Node {
	c.log(fmt.Sprintf("Get on ns (%s)\n", ns))
	if !c.Frozen() {
		panic("must freeze before getting")
	}
	retNodes := new([]Node)
	// Return if no root exists (no tree without a root)
	if c.root == nil {
		c.log(fmt.Sprintln("ctree: no root - returning nil"))
		return nil
	}

	if len(c.root.keys) == 0 {
		//This will be the case when a plugin returns an empty configPolicyTree
		return nil
	}

	rootKeyLength := len(c.root.keys)

	if len(ns) < rootKeyLength {
		c.log(fmt.Sprintln("ns less than root key length - returning nil"))
		return nil
	}

	match, remain := ns[:rootKeyLength], ns[rootKeyLength:]
	if bytes.Compare(nsToByteArray(match), c.root.keysBytes) != 0 {
		c.log(fmt.Sprintf("no match versus root key (match:'%s' != root:'%s')\n", string(nsToByteArray(match)), string(c.root.keysBytes)))
		return nil
	}
	c.log(fmt.Sprintf("Match root key (match:'%s' == root:'%s')\n", string(nsToByteArray(match)), string(c.root.keysBytes)))

	if c.root.Node != nil {
		c.log(fmt.Sprintf("adding root node (not nil) to nodes to merge (%v)\n", c.root.Node))
		*retNodes = append(*retNodes, c.root.Node)
	}

	c.log(fmt.Sprintf("children to get from (%d)\n", len(c.root.nodes)))
	for _, child := range c.root.nodes {
		childNodes := child.get(remain)
		*retNodes = append(*retNodes, *childNodes...)
	}
	if len(*retNodes) == 0 {
		// There are no child nodes with configs so we return
		return nil
	}

	c.log(fmt.Sprintf("nodes to merge count (%d)\n", len(*retNodes)))
	// Call Node.Merge() sequentially on the retNodes
	rn := (*retNodes)[0]
	for _, n := range (*retNodes)[1:] {
		rn = rn.Merge(n)
	}
	return rn
}

// Freeze sets the ConfigTree's freezeFlag to true
func (c *ConfigTree) Freeze() {
	if !c.freezeFlag {
		c.freezeFlag = true
		c.compact()
	}
}

// Frozen returns the bool value of ConfigTree freezeFlag
func (c *ConfigTree) Frozen() bool {
	return c.freezeFlag
}

func (c *ConfigTree) compact() {
	if c.root != nil {
		c.root.compact()
	}
}

// Print prints out the ConfigTree
func (c *ConfigTree) Print() {
	c.root.print("")
}

// Node interface
type Node interface {
	Merge(Node) Node
}

type node struct {
	nodes     []*node
	keys      []string
	keysBytes []byte
	Node      Node
}

// MarshalJSON marshals the ConfigTree.
func (n *node) MarshalJSON() ([]byte, error) {
	return json.Marshal(&struct {
		Nodes     []*node  `json:"nodes"`
		Keys      []string `json:"keys"`
		KeysBytes []byte   `json:"keysbytes"`
		Node      Node     `json:"node"`
	}{
		Nodes:     n.nodes,
		Keys:      n.keys,
		KeysBytes: n.keysBytes,
		Node:      n.Node,
	})
}

// GobEncode encodes every member of node struct instance
func (n *node) GobEncode() ([]byte, error) {
	w := new(bytes.Buffer)
	encoder := gob.NewEncoder(w)
	if err := encoder.Encode(n.nodes); err != nil {
		return nil, err
	}
	if err := encoder.Encode(n.keys); err != nil {
		return nil, err
	}
	if err := encoder.Encode(n.keysBytes); err != nil {
		return nil, err
	}
	if err := encoder.Encode(&n.Node); err != nil {
		return nil, err
	}
	return w.Bytes(), nil
}

// GobDecode decodes every member of node struct instance
func (n *node) GobDecode(buf []byte) error {
	r := bytes.NewBuffer(buf)
	decoder := gob.NewDecoder(r)
	if err := decoder.Decode(&n.nodes); err != nil {
		return err
	}
	if err := decoder.Decode(&n.keys); err != nil {
		return err
	}
	if err := decoder.Decode(&n.keysBytes); err != nil {
		return err
	}
	return decoder.Decode(&n.Node)
}

func (n *node) setKeys(k []string) {
	n.keys = k
	n.keysBytes = nsToByteArray(n.keys)
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
	newNode := new(node)
	newNode.setKeys([]string{f})
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
			n.setKeys(append(n.keys, n.nodes[0].keys...))

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
