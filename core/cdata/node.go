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
	"encoding/json"
	"fmt"
	"sync"

	"github.com/intelsdi-x/snap/core/ctypes"
	"github.com/intelsdi-x/snap/pkg/ctree"
)

// Represents a set of configuration data
type ConfigDataNode struct {
	mutex *sync.Mutex
	table map[string]ctypes.ConfigValue
}

// GobEcode encodes a ConfigDataNode in go binary format
func (c *ConfigDataNode) GobEncode() ([]byte, error) {
	w := new(bytes.Buffer)
	encoder := gob.NewEncoder(w)
	if err := encoder.Encode(&c.table); err != nil {
		return nil, err
	}
	return w.Bytes(), nil
}

// GobDecode decodes a GOB into a ConfigDataNode
func (c *ConfigDataNode) GobDecode(buf []byte) error {
	r := bytes.NewBuffer(buf)
	c.mutex = new(sync.Mutex)
	decoder := gob.NewDecoder(r)
	return decoder.Decode(&c.table)
}

// MarshalJSON marshals a ConfigDataNode into JSON
func (c *ConfigDataNode) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.table)
}

// UnmarshalJSON unmarshals JSON into a ConfigDataNode
func (c *ConfigDataNode) UnmarshalJSON(data []byte) error {
	t := map[string]interface{}{}
	c.table = map[string]ctypes.ConfigValue{}
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.UseNumber()
	if err := dec.Decode(&t); err != nil {
		return err
	}

	for k, i := range t {
		switch t := i.(type) {
		case string:
			c.table[k] = ctypes.ConfigValueStr{Value: t}
		case bool:
			c.table[k] = ctypes.ConfigValueBool{Value: t}
		case json.Number:
			if v, err := t.Int64(); err == nil {
				c.table[k] = ctypes.ConfigValueInt{Value: int(v)}
				continue
			}
			if v, err := t.Float64(); err == nil {
				c.table[k] = ctypes.ConfigValueFloat{Value: v}
				continue
			}
		default:
			return fmt.Errorf("Error Unmarshalling JSON ConfigDataNode. Key: %v Type: %v is unsupported.", k, t)
		}
	}
	c.mutex = new(sync.Mutex)
	return nil
}

// Returns a new and empty node.
func NewNode() *ConfigDataNode {
	return &ConfigDataNode{
		mutex: new(sync.Mutex),
		table: make(map[string]ctypes.ConfigValue),
	}
}

func FromTable(table map[string]ctypes.ConfigValue) *ConfigDataNode {
	return &ConfigDataNode{
		mutex: new(sync.Mutex),
		table: table,
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
	// this ConfigDataNode overwriting where needed.
	for k, v := range t {
		c.AddItem(k, v)
	}
	// Return modified version of ConfigDataNode(as ctree.Node)
	return c
}

// Merges a ConfigDataNode with this one but does not overwrite any
// conflicting values. Any conflicts are decided by the callers value.
func (c *ConfigDataNode) ReverseMergeInPlace(n ctree.Node) ctree.Node {
	cd := n.(*ConfigDataNode)
	new_table := make(map[string]ctypes.ConfigValue)
	// Lock here since we are modifying c.table
	c.mutex.Lock()
	defer c.mutex.Unlock()
	t := cd.Table()
	t2 := c.table
	for k, v := range t {
		new_table[k] = v
	}
	for k, v := range t2 {
		new_table[k] = v
	}
	c.table = new_table
	return c
}

// Merges a ConfigDataNode with a copy of the current ConfigDataNode and returns
// the copy.  The merge does not overwrite any conflicting values.
// Any conflicts are decided by the callers value.
func (c *ConfigDataNode) ReverseMerge(n ctree.Node) *ConfigDataNode {
	cd := n.(*ConfigDataNode)
	copy := NewNode()
	t2 := c.table
	for k, v := range cd.Table() {
		copy.table[k] = v
	}
	for k, v := range t2 {
		copy.table[k] = v
	}
	return copy
}

// ApplyDefaults will set default values if the given ConfigDataNode doesn't
// already have a value for the given configuration.
func (c *ConfigDataNode) ApplyDefaults(defaults map[string]ctypes.ConfigValue) {
	// Lock here since we are modifying c.table
	c.mutex.Lock()
	defer c.mutex.Unlock()
	for name, def := range defaults {
		if _, ok := c.table[name]; !ok {
			c.table[name] = def
		}
	}
}

// Deletes a field in ConfigDataNode. If the field does not exist Delete is
// considered a no-op
func (c ConfigDataNode) DeleteItem(k string) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	delete(c.table, k)
}
