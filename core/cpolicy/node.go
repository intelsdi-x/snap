package cpolicy

import (
	"sync"

	"github.com/intelsdilabs/pulse/pkg/ctree"
)

type ConfigPolicyNode struct {
	rules map[string]Rule
	mutex *sync.Mutex
}

// Adds a rule to this policy node
func (p *ConfigPolicyNode) Add(r Rule) {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	p.rules[r.Key()] = r
}

// Validates and returns a processed policy node or nil and error if validation has failed
func (c *ConfigPolicyNode) Process() (*ConfigPolicyNode, error) {
	return nil, nil
}

// Merges a provided policy node over (overwriting) this one and returns the copy.
func (c ConfigPolicyNode) Merge() ctree.Node {
	return nil
}
