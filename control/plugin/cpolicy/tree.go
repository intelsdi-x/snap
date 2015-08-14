package cpolicy

import (
	"bytes"
	"encoding/gob"
	"encoding/json"

	"github.com/intelsdi-x/pulse/pkg/ctree"
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

func (c *ConfigPolicyTree) GobEncode() ([]byte, error) {
	//todo throw an error if not frozen
	w := new(bytes.Buffer)
	encoder := gob.NewEncoder(w)
	if err := encoder.Encode(c.cTree); err != nil {
		return nil, err
	}
	return w.Bytes(), nil
}

func (c *ConfigPolicyTree) GobDecode(buf []byte) error {
	r := bytes.NewBuffer(buf)
	decoder := gob.NewDecoder(r)
	return decoder.Decode(&c.cTree)
}

// UnmarshalJSON unmarshals JSON into a ConfigPolicyTree
func (c *ConfigPolicyTree) UnmarshalJSON(data []byte) error {
	m := map[string]map[string]interface{}{}
	decoder := json.NewDecoder(bytes.NewReader(data))
	if err := decoder.Decode(&m); err != nil {
		return err
	}
	c.cTree = ctree.New()
	if ctree, ok := m["PolicyTree"]["ctree"]; ok {
		if root, ok := ctree.(map[string]interface{}); ok {
			if node, ok := root["root"]; ok {
				if n, ok := node.(map[string]interface{}); ok {
					return unmarshalJSON(n, &[]string{}, c.cTree)
				}
			}
		}
	}
	return nil
}

func unmarshalJSON(m map[string]interface{}, keys *[]string, tree *ctree.ConfigTree) error {
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
					addRulesToConfigPolicyNode(rules, cpn)
				}
				tree.Add(*keys, cpn)
			}
		}
	}
	if val, ok := m["nodes"]; ok {
		if nodes, ok := val.([]interface{}); ok {
			for _, node := range nodes {
				if n, ok := node.(map[string]interface{}); ok {
					unmarshalJSON(n, keys, tree)
				}
			}
		}
	}
	return nil
}

// MarshalJSON marshals a ConfigPolicyTree into JSON
func (c *ConfigPolicyTree) MarshalJSON() ([]byte, error) {
	return json.Marshal(&struct {
		CTree *ctree.ConfigTree `json:"ctree"`
	}{
		CTree: c.cTree,
	})
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
		return NewPolicyNode()
	}
	switch t := n.(type) {
	case ConfigPolicyNode:
		return &t
	default:
		return t.(*ConfigPolicyNode)

	}
}

// Freezes the ConfigDataTree from future writes (adds) and triggers compression
// of tree into read-performant version.
func (c *ConfigPolicyTree) Freeze() {
	c.cTree.Freeze()
}
