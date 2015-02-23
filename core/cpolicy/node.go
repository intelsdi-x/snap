package cpolicy

import (
	"errors"
	"fmt"
	"sync"

	"github.com/intelsdilabs/pulse/core/ctypes"
	"github.com/intelsdilabs/pulse/pkg/ctree"
)

type ProcessingErrors struct {
	errors []error
	mutex  *sync.Mutex
}

func newProcessingErrors() *ProcessingErrors {
	return &ProcessingErrors{
		errors: []error{},
		mutex:  &sync.Mutex{},
	}
}

func (p *ProcessingErrors) Errors() []error {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	return p.errors
}

func (p *ProcessingErrors) HasErrors() bool {
	return len(p.errors) > 0
}

func (p *ProcessingErrors) AddError(e error) {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	p.errors = append(p.errors, e)
}

type ConfigPolicyNode struct {
	rules map[string]Rule
	mutex *sync.Mutex
}

func NewPolicyNode() *ConfigPolicyNode {
	return &ConfigPolicyNode{
		rules: make(map[string]Rule),
		mutex: &sync.Mutex{},
	}
}

// Adds a rule to this policy node
func (p *ConfigPolicyNode) Add(rules ...Rule) {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	for _, r := range rules {
		p.rules[r.Key()] = r
	}
}

// Validates and returns a processed policy node or nil and error if validation has failed
func (c *ConfigPolicyNode) Process(m map[string]ctypes.ConfigValue) (*map[string]ctypes.ConfigValue, *ProcessingErrors) {
	pErrors := newProcessingErrors()
	// Loop through each rule and process
	for key, rule := range c.rules {
		// items exists for rule
		if cv, ok := m[key]; ok {
			println(cv)

			// Validate versus matching data
			e := rule.Validate(cv)
			if e != nil {
				pErrors.AddError(e)
			}
		} else {
			// If it was required add error
			if rule.Required() {
				e := errors.New(fmt.Sprintf("required key missing (%s)", key))
				pErrors.AddError(e)
			} else {
				// If default returns we should add it
				cv := rule.Default()
				if cv != nil {
					m[key] = cv
				}

			}
		}
	}

	if pErrors.HasErrors() {
		return nil, pErrors
	}
	return &m, pErrors
}

// Merges a provided policy node over (overwriting) this one and returns the copy.
func (c ConfigPolicyNode) Merge() ctree.Node {
	return nil
}
