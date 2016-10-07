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

package cpolicy

import (
	"bytes"
	"encoding/gob"
	"encoding/json"

	"github.com/intelsdi-x/snap/pkg/ctree"
)

// Allows adding of config policy by namespace and retrieving of policy from a tree
// at a specific namespace (merging the relevant hiearchy). Uses pkg.ConfigTree.
type ConfigPolicy struct {
	config *ctree.ConfigTree
}

// Returns a new ConfigPolicy.
func New() *ConfigPolicy {
	return &ConfigPolicy{
		config: ctree.New(),
	}
}

func (c *ConfigPolicy) GobEncode() ([]byte, error) {
	//todo throw an error if not frozen
	w := new(bytes.Buffer)
	encoder := gob.NewEncoder(w)
	if err := encoder.Encode(c.config); err != nil {
		return nil, err
	}
	return w.Bytes(), nil
}

func (c *ConfigPolicy) GobDecode(buf []byte) error {
	r := bytes.NewBuffer(buf)
	decoder := gob.NewDecoder(r)
	return decoder.Decode(&c.config)
}

// UnmarshalJSON unmarshals JSON into a ConfigPolicy
func (c *ConfigPolicy) UnmarshalJSON(data []byte) error {
	m := map[string]map[string]interface{}{}
	decoder := json.NewDecoder(bytes.NewReader(data))
	if err := decoder.Decode(&m); err != nil {
		return err
	}
	c.config = ctree.New()
	if config, ok := m["config"]; ok {
		if node, ok := config["root"]; ok {
			if n, ok := node.(map[string]interface{}); ok {
				return unmarshalJSON(n, &[]string{}, c.config)
			}
		}
	}
	return nil
}

func unmarshalJSON(m map[string]interface{}, keys *[]string, config *ctree.ConfigTree) error {
	if val, ok := m["keys"]; ok {
		if items, ok := val.([]interface{}); ok {
			for _, i := range items {
				if key, ok := i.(string); ok {
					*keys = append(*keys, key)
				}
			}
		}
	}
	if val, ok := m["node"]; ok {
		if node, ok := val.(map[string]interface{}); ok {
			if nval, ok := node["rules"]; ok {
				cpn := NewPolicyNode()
				if rules, ok := nval.(map[string]interface{}); ok {
					err := addRulesToConfigPolicyNode(rules, cpn)
					if err != nil {
						return err
					}
				}
				config.Add(*keys, cpn)
			}
		}
	}
	if val, ok := m["nodes"]; ok {
		if nodes, ok := val.([]interface{}); ok {
			for _, node := range nodes {
				if n, ok := node.(map[string]interface{}); ok {
					unmarshalJSON(n, keys, config)
				}
			}
		}
	}
	return nil
}

// MarshalJSON marshals a ConfigPolicy into JSON
func (c *ConfigPolicy) MarshalJSON() ([]byte, error) {
	return json.Marshal(&struct {
		Config *ctree.ConfigTree `json:"config"`
	}{
		Config: c.config,
	})
}

// Adds a ConfigPolicyNode at the provided namespace.
func (c *ConfigPolicy) Add(ns []string, cpn *ConfigPolicyNode) {
	c.config.Add(ns, cpn)
}

// Returns a ConfigPolicyNode that is a merged version of the namespace provided.
func (c *ConfigPolicy) Get(ns []string) *ConfigPolicyNode {

	n := c.config.Get(ns)
	if n == nil {
		return NewPolicyNode()
	}
	switch t := n.(type) {
	case ConfigPolicyNode:
		return &t
	default:
		return t.(*ConfigPolicyNode)
	}
}

type keyNode struct {
	Key []string
	*ConfigPolicyNode
}

func (c *ConfigPolicy) GetAll() []keyNode {

	ret := make([]keyNode, 0)
	for _, node := range c.config.GetAll() {
		key := node.Key
		switch t := node.Node.(type) {
		case *ConfigPolicyNode:
			ret = append(ret, keyNode{Key: key, ConfigPolicyNode: t})
		}
	}
	return ret
}
