package cpolicy

import (
	"bytes"
	"encoding/gob"
	"errors"
	"fmt"

	"github.com/intelsdi-x/pulse/core/ctypes"
)

// A rule validating against string-typed config
type IntRule struct {
	rule

	key      string
	required bool
	default_ *int
	minimum  *int
	maximum  *int
}

// Returns a new int-typed rule. Arguments are key(string), required(bool), default(int), min(int), max(int)
func NewIntegerRule(key string, req bool, opts ...int) (*IntRule, error) {
	// Return error if key is empty
	if key == "" {
		return nil, EmptyKeyError
	}

	options := make([]*int, 1)
	for i, o := range opts {
		options[i] = &o
	}

	return &IntRule{
		key:      key,
		required: req,
		default_: options[0],
	}, nil
}

func (i *IntRule) GobEncode() ([]byte, error) {
	w := new(bytes.Buffer)
	encoder := gob.NewEncoder(w)
	if err := encoder.Encode(i.key); err != nil {
		return nil, err
	}
	if err := encoder.Encode(i.required); err != nil {
		return nil, err
	}
	if i.default_ == nil {
		encoder.Encode(false)
	} else {
		encoder.Encode(true)
		if err := encoder.Encode(&i.default_); err != nil {
			return nil, err
		}
	}
	if i.minimum == nil {
		encoder.Encode(false)
	} else {
		encoder.Encode(true)
		if err := encoder.Encode(i.minimum); err != nil {
			return nil, err
		}
	}
	if i.maximum == nil {
		encoder.Encode(false)
	} else {
		encoder.Encode(true)
		if err := encoder.Encode(i.maximum); err != nil {
			return nil, err
		}
	}
	return w.Bytes(), nil
}

func (i *IntRule) GobDecode(buf []byte) error {
	r := bytes.NewBuffer(buf)
	decoder := gob.NewDecoder(r)
	if err := decoder.Decode(&i.key); err != nil {
		return err
	}
	if err := decoder.Decode(&i.required); err != nil {
		return err
	}
	var is_default_set bool
	decoder.Decode(&is_default_set)
	if is_default_set {
		return decoder.Decode(&i.default_)
	}
	var is_minimum_set bool
	decoder.Decode(&is_minimum_set)
	if is_minimum_set {
		if err := decoder.Decode(&i.minimum); err != nil {
			return err
		}
	}
	var is_maximum_set bool
	decoder.Decode(&is_maximum_set)
	if is_maximum_set {
		if err := decoder.Decode(&i.maximum); err != nil {
			return err
		}
	}
	return nil
}

// Returns the key
func (i *IntRule) Key() string {
	return i.key
}

// Validates a config value against this rule.
func (i *IntRule) Validate(cv ctypes.ConfigValue) error {
	// Check that type is correct
	if cv.Type() != "integer" {
		return wrongType(i.key, cv.Type(), "integer")
	}
	// Check minimum. Type should be safe now because of the check above.
	if i.minimum != nil && cv.(ctypes.ConfigValueInt).Value < *i.minimum {
		return errors.New(fmt.Sprintf("value is under minimum (%s value %d < %d)", i.key, cv.(ctypes.ConfigValueInt).Value, *i.minimum))
	}
	// Check maximum. Type should be safe now because of the check above.
	if i.maximum != nil && cv.(ctypes.ConfigValueInt).Value > *i.maximum {
		return errors.New(fmt.Sprintf("value is over maximum (%s value %d > %d)", i.key, cv.(ctypes.ConfigValueInt).Value, *i.maximum))
	}
	return nil
}

func (i *IntRule) Default() ctypes.ConfigValue {
	if i.default_ != nil {
		return &ctypes.ConfigValueInt{Value: *i.default_}
	}
	return nil
}

func (i *IntRule) Required() bool {
	return i.required
}

func (i *IntRule) SetMinimum(m int) {
	i.minimum = &m
}

func (i *IntRule) SetMaximum(m int) {
	i.maximum = &m
}
