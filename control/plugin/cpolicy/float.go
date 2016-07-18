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
	FloatType = "float"
)

// FloatRule A rule validating against string-typed config
type FloatRule struct {
	rule

	key      string
	required bool
	default_ *float64
	minimum  *float64
	maximum  *float64
}

// MarshalJSON marshals a FloatRule into JSON
func (f *FloatRule) MarshalJSON() ([]byte, error) {
	return json.Marshal(&struct {
		Key      string             `json:"key"`
		Required bool               `json:"required"`
		Default  ctypes.ConfigValue `json:"default,omitempty"`
		Minimum  ctypes.ConfigValue `json:"minimum,omitempty"`
		Maximum  ctypes.ConfigValue `json:"maximum,omitempty"`
		Type     string             `json:"type"`
	}{
		Key:      f.key,
		Required: f.required,
		Default:  f.Default(),
		Minimum:  f.Minimum(),
		Maximum:  f.Maximum(),
		Type:     FloatType,
	})
}

func (s *FloatRule) Type() string {
	return FloatType
}

// GobEncode encodes a FloatRule into a GOB
func (f *FloatRule) GobEncode() ([]byte, error) {
	w := new(bytes.Buffer)
	encoder := gob.NewEncoder(w)
	if err := encoder.Encode(f.key); err != nil {
		return nil, err
	}
	if err := encoder.Encode(f.required); err != nil {
		return nil, err
	}
	if f.default_ == nil {
		encoder.Encode(false)
	} else {
		encoder.Encode(true)
		if err := encoder.Encode(&f.default_); err != nil {
			return nil, err
		}
	}
	if f.minimum == nil {
		encoder.Encode(false)
	} else {
		encoder.Encode(true)
		if err := encoder.Encode(f.minimum); err != nil {
			return nil, err
		}
	}
	if f.maximum == nil {
		encoder.Encode(false)
	} else {
		encoder.Encode(true)
		if err := encoder.Encode(f.maximum); err != nil {
			return nil, err
		}
	}
	return w.Bytes(), nil
}

// GobDecode decodes a GOB into a FloatRule
func (f *FloatRule) GobDecode(buf []byte) error {
	r := bytes.NewBuffer(buf)
	decoder := gob.NewDecoder(r)
	if err := decoder.Decode(&f.key); err != nil {
		return err
	}
	if err := decoder.Decode(&f.required); err != nil {
		return err
	}
	var is_default_set bool
	decoder.Decode(&is_default_set)
	if is_default_set {
		return decoder.Decode(&f.default_)
	}
	var is_minimum_set bool
	decoder.Decode(&is_minimum_set)
	if is_minimum_set {
		if err := decoder.Decode(&f.minimum); err != nil {
			return err
		}
	}
	var is_maximum_set bool
	decoder.Decode(&is_maximum_set)
	if is_maximum_set {
		if err := decoder.Decode(&f.maximum); err != nil {
			return err
		}
	}
	return nil
}

// NewFloatRule returns a new float-typed rule. Arguments are key(string), required(bool), default(float64)
func NewFloatRule(key string, req bool, opts ...float64) (*FloatRule, error) {
	// Return error if key is empty
	if key == "" {
		return nil, EmptyKeyError
	}

	f := &FloatRule{
		key:      key,
		required: req,
	}

	if len(opts) > 0 {
		f.default_ = &opts[0]
	}
	return f, nil
}

// Key Returns the key
func (f *FloatRule) Key() string {
	return f.key
}

// Validate Validates a config value against this rule.
func (f *FloatRule) Validate(cv ctypes.ConfigValue) error {
	// Check that type is correct
	if cv.Type() != FloatType {
		return wrongType(f.key, cv.Type(), FloatType)
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

// Default returns the rule's default value
func (f *FloatRule) Default() ctypes.ConfigValue {
	if f.default_ != nil {
		return ctypes.ConfigValueFloat{Value: *f.default_}
	}
	return nil
}

// Required a bool indicating whether the rule is required
func (f *FloatRule) Required() bool {
	return f.required
}

// SetMinimum sets the minimum allowable value for this rule
func (f *FloatRule) SetMinimum(m float64) {
	f.minimum = &m
}

// SetMaximum sets the maximum allowable value for this rule
func (f *FloatRule) SetMaximum(m float64) {
	f.maximum = &m
}

func (i *FloatRule) Minimum() ctypes.ConfigValue {
	if i.minimum != nil {
		return ctypes.ConfigValueFloat{Value: *i.minimum}
	}
	return nil
}

func (i *FloatRule) Maximum() ctypes.ConfigValue {
	if i.maximum != nil {
		return ctypes.ConfigValueFloat{Value: *i.maximum}
	}
	return nil
}
