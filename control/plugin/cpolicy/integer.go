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

	"github.com/intelsdi-x/snap/core/ctypes"
)

const (
	IntegerType = "integer"
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

// NewIntegerRule returns a new int-typed rule. Arguments are key(string), required(bool), default(int)
func NewIntegerRule(key string, req bool, opts ...int) (*IntRule, error) {
	// Return error if key is empty
	if key == "" {
		return nil, EmptyKeyError
	}

	i := &IntRule{
		key:      key,
		required: req,
	}

	if len(opts) > 0 {
		i.default_ = &opts[0]
	}
	return i, nil
}

func (i *IntRule) Type() string {
	return IntegerType
}

// MarshalJSON marshals a IntRule into JSON
func (i *IntRule) MarshalJSON() ([]byte, error) {
	return json.Marshal(&struct {
		Key      string             `json:"key"`
		Required bool               `json:"required"`
		Default  ctypes.ConfigValue `json:"default,omitempty"`
		Minimum  ctypes.ConfigValue `json:"minimum,omitempty"`
		Maximum  ctypes.ConfigValue `json:"maximum,omitempty"`
		Type     string             `json:"type"`
	}{
		Key:      i.key,
		Required: i.required,
		Default:  i.Default(),
		Minimum:  i.Minimum(),
		Maximum:  i.Maximum(),
		Type:     IntegerType,
	})
}

// GobEncode encodes a IntRule into a GOB
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

// GobDecode decodes a GOB into a IntRule
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

// Key Returns the key
func (i *IntRule) Key() string {
	return i.key
}

// Validate Validates a config value against this rule.
func (i *IntRule) Validate(cv ctypes.ConfigValue) error {
	// Check that type is correct
	// when unmarshalling JSON numbers are converted to floats which is the reason
	// we are checking for integer below.
	// http://golang.org/pkg/encoding/json/#Marshal
	if cv.Type() != IntegerType && cv.Type() != FloatType {
		return wrongType(i.key, cv.Type(), IntegerType)
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

// Default return this rules default value
func (i *IntRule) Default() ctypes.ConfigValue {
	if i.default_ != nil {
		return ctypes.ConfigValueInt{Value: *i.default_}
	}
	return nil
}

// Required returns a boolean indicating if this rule is required
func (i *IntRule) Required() bool {
	return i.required
}

// SetMinimum sets the minimum allowed value
func (i *IntRule) SetMinimum(m int) {
	i.minimum = &m
}

// SetMaximum sets the maximum allowed value
func (i *IntRule) SetMaximum(m int) {
	i.maximum = &m
}

func (i *IntRule) Minimum() ctypes.ConfigValue {
	if i.minimum != nil {
		return ctypes.ConfigValueInt{Value: *i.minimum}
	}
	return nil
}

func (i *IntRule) Maximum() ctypes.ConfigValue {
	if i.maximum != nil {
		return ctypes.ConfigValueInt{Value: *i.maximum}
	}
	return nil
}
