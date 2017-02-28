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
	"errors"
	"fmt"
	"sync"

	"github.com/intelsdi-x/snap/core/ctypes"
	"github.com/intelsdi-x/snap/pkg/ctree"
)

type ProcessingErrors struct {
	errors []error
	mutex  *sync.Mutex
}

func NewProcessingErrors() *ProcessingErrors {
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

func (c *ConfigPolicyNode) CopyRules() ([]Rule, error) {
	rules := []Rule{}
	for _, rule := range c.rules {
		var err error
		switch rule.(type) {
		case *BoolRule:
			var newBoolRule *BoolRule
			if rule.Default() != nil {
				newBoolRule, err = NewBoolRule(rule.Key(), rule.Required(), rule.Default().(ctypes.ConfigValueBool).Value)
			} else {
				newBoolRule, err = NewBoolRule(rule.Key(), rule.Required())
			}
			rules = append(rules, newBoolRule)
		case *StringRule:
			var newStringRule *StringRule
			if rule.Default() != nil {
				newStringRule, err = NewStringRule(rule.Key(), rule.Required(), rule.Default().(ctypes.ConfigValueStr).Value)
			} else {
				newStringRule, err = NewStringRule(rule.Key(), rule.Required())
			}
			rules = append(rules, newStringRule)
		case *FloatRule:
			var newFloatRule *FloatRule
			if rule.Default() != nil {
				newFloatRule, err = NewFloatRule(rule.Key(), rule.Required(), rule.Default().(ctypes.ConfigValueFloat).Value)
			} else {
				newFloatRule, err = NewFloatRule(rule.Key(), rule.Required())
			}
			rules = append(rules, newFloatRule)
		case *IntRule:
			var newIntRule *IntRule
			if rule.Default() != nil {
				newIntRule, err = NewIntegerRule(rule.Key(), rule.Required(), rule.Default().(ctypes.ConfigValueInt).Value)
			} else {
				newIntRule, err = NewIntegerRule(rule.Key(), rule.Required())
			}
			rules = append(rules, newIntRule)
		default:
			return []Rule{}, errors.New(fmt.Sprint("Unknown rule type"))
		}

		if err != nil {
			return []Rule{}, errors.New(fmt.Sprintf("Could not create rule %s type %s ", rule.Key(), rule.Type()))
		}
	}
	return rules, nil
}

// UnmarshalJSON unmarshals JSON into a ConfigPolicyNode
func (c *ConfigPolicyNode) UnmarshalJSON(data []byte) error {
	m := map[string]interface{}{}
	decoder := json.NewDecoder(bytes.NewReader(data))
	if err := decoder.Decode(&m); err != nil {
		return err
	}
	if rs, ok := m["rules"]; ok {
		if rules, ok := rs.(map[string]interface{}); ok {
			addRulesToConfigPolicyNode(rules, c)
		}
	}
	return nil
}

func (c *ConfigPolicyNode) MarshalJSON() ([]byte, error) {
	return json.Marshal(&struct {
		Rules map[string]Rule `json:"rules"`
	}{
		Rules: c.rules,
	})
}

func (c *ConfigPolicyNode) GobEncode() ([]byte, error) {
	w := new(bytes.Buffer)
	encoder := gob.NewEncoder(w)
	if err := encoder.Encode(&c.rules); err != nil {
		return nil, err
	}
	return w.Bytes(), nil
}

func (c *ConfigPolicyNode) GobDecode(buf []byte) error {
	c.mutex = &sync.Mutex{}
	r := bytes.NewBuffer(buf)
	decoder := gob.NewDecoder(r)
	return decoder.Decode(&c.rules)
}

// Adds a rule to this policy node
func (p *ConfigPolicyNode) Add(rules ...Rule) {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	for _, r := range rules {
		p.rules[r.Key()] = r
	}
}

type RuleTableSlice []RuleTable

type RuleTable struct {
	Name     string      `json:"name"`
	Type     string      `json:"type"`
	Default  interface{} `json:"default,omitempty"`
	Required bool        `json:"required"`
	Minimum  interface{} `json:"minimum,omitempty"`
	Maximum  interface{} `json:"maximum,omitempty"`
}

func (p *ConfigPolicyNode) RulesAsTable() RuleTableSlice {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	rt := make([]RuleTable, 0, len(p.rules))
	for _, r := range p.rules {
		rt = append(rt, RuleTable{
			Name:     r.Key(),
			Type:     r.Type(),
			Default:  r.Default(),
			Required: r.Required(),
			Minimum:  r.Minimum(),
			Maximum:  r.Maximum(),
		})
	}
	return rt
}

func (c *ConfigPolicyNode) HasRules() bool {
	if len(c.rules) > 0 {
		return true
	}
	return false
}

// Defaults returns a map[string]ctypes.ConfigValue for all of the rules that
// have defaults.
func (c *ConfigPolicyNode) Defaults() map[string]ctypes.ConfigValue {
	defaults := map[string]ctypes.ConfigValue{}
	for name, rule := range c.rules {
		if def := rule.Default(); def != nil {
			defaults[name] = def
		}
	}
	return defaults
}

// Validates and returns a processed policy node or nil and error if validation has failed
func (c *ConfigPolicyNode) Process(m map[string]ctypes.ConfigValue) (*map[string]ctypes.ConfigValue, *ProcessingErrors) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	pErrors := NewProcessingErrors()
	// Loop through each rule and process
	for key, rule := range c.rules {
		// items exists for rule
		if cv, ok := m[key]; ok {
			// Validate versus matching data
			e := rule.Validate(cv)
			if e != nil {
				pErrors.AddError(e)
			}
		} else {
			// If it was required add error
			if rule.Required() {
				e := fmt.Errorf("required key missing (%s)", key)
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

// AddDefaults validates and returns a processed policy node or nil and error if validation has failed
func (c *ConfigPolicyNode) AddDefaults(m map[string]ctypes.ConfigValue) (*map[string]ctypes.ConfigValue, *ProcessingErrors) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	pErrors := NewProcessingErrors()
	// Loop through each rule and process
	for key, rule := range c.rules {
		// items exists for rule
		if _, ok := m[key]; ok {
			pErrors.AddError(fmt.Errorf("The key \"%v\" already has a default policy for this plugin", key))
		} else {
			// If it was required add error
			if !rule.Required() {
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

// Merges a ConfigPolicyNode on top of this one (overwriting items where it occurs).
func (c ConfigPolicyNode) Merge(n ctree.Node) ctree.Node {
	// Because Add only allows the ConfigPolicyNode type we
	// are safe to convert ctree.Node interface to ConfigPolicyNode
	cd := n.(*ConfigPolicyNode)
	// For the rules in the passed ConfigPolicyNode(converted) add each rule to
	// this ConfigPolicyNode overwriting where needed.
	no := NewPolicyNode()
	for _, r := range c.rules {
		no.Add(r)
	}

	for _, r := range cd.rules {

		no.Add(r)
	}
	// Return modified version of ConfigPolicyNode(as ctree.Node)
	return no
}

// addRulesToConfigPolicyNode accepts a map of empty interfaces that will be
// marshalled into rules which will be added to the ConfigPolicyNode provided
// as the second argument.  This function is called used by the UnmarshalJSON
// for ConfigPolicy and ConfigPolicyNode.
func addRulesToConfigPolicyNode(rules map[string]interface{}, cpn *ConfigPolicyNode) error {
	for k, rule := range rules {
		if rule, ok := rule.(map[string]interface{}); ok {
			req, _ := rule["required"].(bool)
			switch rule["type"] {
			case "integer":
				r, _ := NewIntegerRule(k, req)
				if d, ok := rule["default"]; ok {
					// json encoding an int results in a float when decoding
					def_, _ := d.(float64)
					def := int(def_)
					r.default_ = &def
				}
				if m, ok := rule["minimum"]; ok {
					min_, _ := m.(float64)
					min := int(min_)
					r.minimum = &min
				}
				if m, ok := rule["maximum"]; ok {
					max_, _ := m.(float64)
					max := int(max_)
					r.maximum = &max
				}
				cpn.Add(r)
			case "string":
				r, _ := NewStringRule(k, req)
				if d, ok := rule["default"]; ok {
					def, _ := d.(string)
					if def != "" {
						r.default_ = &def
					}
				}

				cpn.Add(r)
			case "bool":
				r, _ := NewBoolRule(k, req)
				if d, ok := rule["default"]; ok {
					def, _ := d.(bool)
					r.default_ = &def
				}

				cpn.Add(r)
			case "float":
				r, _ := NewFloatRule(k, req)
				if d, ok := rule["default"]; ok {
					def, _ := d.(float64)
					r.default_ = &def
				}
				if m, ok := rule["minimum"]; ok {
					min, _ := m.(float64)
					r.minimum = &min
				}
				if m, ok := rule["maximum"]; ok {
					max, _ := m.(float64)
					r.maximum = &max
				}
				cpn.Add(r)
			default:
				return errors.New("unknown type")
			}
		}
	}
	return nil
}
