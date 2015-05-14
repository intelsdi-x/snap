package cpolicy

import (
	"errors"
	"fmt"

	"github.com/intelsdi-x/pulse/core/ctypes"
)

// A rule validating against string-typed config
type floatRule struct {
	rule

	key      string
	required bool
	default_ *float64
	minimum  *float64
	maximum  *float64
}

// Returns a new float-typed rule. Arguments are key(string), required(bool), default(float64), min(float64), max(float64)
func NewFloatRule(key string, req bool, opts ...float64) (*floatRule, error) {
	// Return error if key is empty
	if key == "" {
		return nil, EmptyKeyError
	}

	options := make([]*float64, 1)
	for i, o := range opts {
		options[i] = &o
	}

	return &floatRule{
		key:      key,
		required: req,
		default_: options[0],
	}, nil
}

// Returns the key
func (f *floatRule) Key() string {
	return f.key
}

// Validates a config value against this rule.
func (f *floatRule) Validate(cv ctypes.ConfigValue) error {
	// Check that type is correct
	if cv.Type() != "float" {
		return wrongType(f.key, cv.Type(), "float")
	}
	// Check minimum. Type should be safe now because of the check above.
	if f.minimum != nil && cv.(ctypes.ConfigValueFloat).Value < *f.minimum {
		return errors.New(fmt.Sprintf("value is under minimum (%s value %f < %f)", f.key, cv.(ctypes.ConfigValueFloat).Value, *f.minimum))
	}
	// Check maximum. Type should be safe now because of the check above.
	if f.maximum != nil && cv.(ctypes.ConfigValueFloat).Value > *f.maximum {
		return errors.New(fmt.Sprintf("value is over maximum (%s value %f > %f)", f.key, cv.(ctypes.ConfigValueFloat).Value, *f.maximum))
	}
	return nil
}

func (f *floatRule) Default() ctypes.ConfigValue {
	if f.default_ != nil {
		return &ctypes.ConfigValueFloat{Value: *f.default_}
	}
	return nil
}

func (f *floatRule) Required() bool {
	return f.required
}

func (f *floatRule) SetMinimum(m float64) {
	f.minimum = &m
}

func (f *floatRule) SetMaximum(m float64) {
	f.maximum = &m
}
