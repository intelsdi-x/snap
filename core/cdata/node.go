package cdata

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/intelsdi-x/pulse/core/ctypes"
	"github.com/intelsdi-x/pulse/pkg/ctree"
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
	return json.Marshal(&struct {
		Table map[string]ctypes.ConfigValue `json:"table"`
	}{
		Table: c.table,
	})
}

// UnmarshalJSON unmarshals JSON into a ConfigDataNode
func (c *ConfigDataNode) UnmarshalJSON(data []byte) error {
	t := map[string]map[string]interface{}{}
	c.table = map[string]ctypes.ConfigValue{}
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.UseNumber()
	if err := dec.Decode(&t); err != nil {
		return err
	}

	for k, i := range t["table"] {
		switch t := i.(map[string]interface{})["Value"].(type) {
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
				fmt.Printf("%v is a float64\n", k)
				c.table[k] = ctypes.ConfigValueFloat{Value: v}
				continue
			}
		default:
			return fmt.Errorf("Error Unmarshalling ConfigDataNode into JSON.  Type %v is unsupported.", k)
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
	// this ConfigDataNode overwritting where needed.
	for k, v := range t {
		c.AddItem(k, v)
	}
	// Return modified version of ConfigDataNode(as ctree.Node)
	return c
}
