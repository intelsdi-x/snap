package plugin

import (
	"errors"
	"reflect"
	"strings"
)

// ConfigPolicy is the collection of policies which are needed to collect a metric.
// As an example, authentication may be needed to retrieve a metric.
type ConfigPolicy map[string][]*Policy

// NewConfigPolicy intializes and returns a pointer to a new ConfigPolicy
func NewConfigPolicy() *ConfigPolicy {
	m := make(map[string][]*Policy)
	c := ConfigPolicy(m)
	return &c
}

/* Add panics if bad data is given, as these are loaded at start time
   and we want to confirm that what is given in the policy is valid
   before attempting to use them at collection time.
*/
// Add adds a policy to a config policy.
func (cp *ConfigPolicy) Add(pluginName, ns string, p *Policy) {
	var (
		blankType reflect.Kind
		blankKey  string
	)
	if p.Type == blankType || p.Key == blankKey {
		panic("Type and Key are required fields on a policy")
	}
	if string(ns[0]) != "/" {
		panic("config policy namespace must begin with [/]")
	}
	s := strings.Split(ns, "/")
	n := strings.Split(pluginName, "/")
	l := len(n)
	for i, node := range n[:l-1] {
		if s[i] != node {
			panic("metric namespace must fall under plugin's namespace")
		}
	}
	if _, ok := (*cp)[ns]; !ok {
		(*cp)[ns] = []*Policy{p}
	} else {
		(*cp)[ns] = append((*cp)[ns], p)
	}
}

// Policy is the policy details.
type Policy struct {
	// Key is the name of the needed field
	// e.g. "username" or "password"
	Key string

	// type uses the Kind constants from reflect.
	// This can be used to test against the given value to
	// confirm the correct type is given before a collection attempt.
	Type reflect.Kind

	// Is this required to collect.
	Required bool

	// a default value
	Default interface{}
}

// Validate validates the type of a given policy input value.
// Leaving as `Validate` (as opposed to ValidateType)
// now as more validation may eventually occur.
func (p *Policy) Validate(pi *PolicyInput) error {
	if pi.Key != p.Key {
		return errors.New("incorrect key given [" + pi.Key + "] for policy [" + p.Key + "]")
	}
	if pi.Value != nil {
		t := reflect.TypeOf(pi.Value)
		if t.Kind() != p.Type {
			return errors.New("invalid type given for policy " + p.Key)
		}
	} else {
		return errors.New("policy input with nil value given")
	}
	return nil
}

// PolicyInput is what gets unmarshalled from an input from
// a mgmt module
type PolicyInput struct {
	Key   string      `json:"key"`
	Value interface{} `json:"value"`
}
