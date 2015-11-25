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

package cdata

import (
	"bytes"
	"encoding/gob"

	"github.com/intelsdi-x/snap/pkg/ctree"
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

func (c *ConfigDataTree) GobEncode() ([]byte, error) {
	w := new(bytes.Buffer)
	encoder := gob.NewEncoder(w)
	if err := encoder.Encode(c.cTree); err != nil {
		return nil, err
	}
	return w.Bytes(), nil
}

func (c *ConfigDataTree) GobDecode(buf []byte) error {
	r := bytes.NewBuffer(buf)
	decoder := gob.NewDecoder(r)
	return decoder.Decode(&c.cTree)
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
	}
	switch t := n.(type) {
	case ConfigDataNode:
		return &t
	default:
		return t.(*ConfigDataNode)

	}
}

// Freezes the ConfigDataTree from future writes (adds) and triggers compression
// of tree into read-performant version.
func (c *ConfigDataTree) Freeze() {
	c.cTree.Freeze()
}
