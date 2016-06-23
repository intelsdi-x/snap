/*
http://www.apache.org/licenses/LICENSE-2.0.txt


Copyright 2016 Intel Corporation

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

	"github.com/intelsdi-x/snap/core/ctypes"
)

const (
	BoolType = "bool"
)

// BoolRule validating against string-typed config
type BoolRule struct {
	rule

	key      string
	required bool
	default_ *bool
}

// NewBoolRule Returns a new bool-typed rule. Arguments are key(string), required(bool), default(bool).
func NewBoolRule(key string, req bool, opts ...bool) (*BoolRule, error) {
	// Return error if key is empty
	if key == "" {
		return nil, EmptyKeyError
	}

	b := &BoolRule{
		key:      key,
		required: req,
	}
	if len(opts) != 0 {
		b.default_ = &opts[0]
	}
	return b, nil
}

// Type Returns a type of Rule
func (b *BoolRule) Type() string {
	return BoolType
}

// MarshalJSON marshals a BoolRule into JSON
func (b *BoolRule) MarshalJSON() ([]byte, error) {
	return json.Marshal(&struct {
		Key      string             `json:"key"`
		Required bool               `json:"required"`
		Default  ctypes.ConfigValue `json:"default,omitempty"`
		Type     string             `json:"type"`
	}{
		Key:      b.key,
		Required: b.required,
		Default:  b.Default(),
		Type:     BoolType,
	})
}

//GobEncode encodes a BoolRule in to a GOB
func (b *BoolRule) GobEncode() ([]byte, error) {
	w := new(bytes.Buffer)
	encoder := gob.NewEncoder(w)
	if err := encoder.Encode(b.key); err != nil {
		return nil, err
	}
	if err := encoder.Encode(b.required); err != nil {
		return nil, err
	}
	if b.default_ == nil {
		encoder.Encode(false)
	} else {
		encoder.Encode(true)
		if err := encoder.Encode(&b.default_); err != nil {
			return nil, err
		}
	}
	return w.Bytes(), nil
}

// GobDecode decodes a GOB into a BoolRule
func (b *BoolRule) GobDecode(buf []byte) error {
	r := bytes.NewBuffer(buf)
	decoder := gob.NewDecoder(r)
	if err := decoder.Decode(&b.key); err != nil {
		return err
	}
	if err := decoder.Decode(&b.required); err != nil {
		return err
	}
	var isDefaultSet bool
	decoder.Decode(&isDefaultSet)
	if isDefaultSet {
		return decoder.Decode(&b.default_)
	}
	return nil
}

// Key returns the key
func (b *BoolRule) Key() string {
	return b.key
}

// Validate validates a config value against this rule.
func (b *BoolRule) Validate(cv ctypes.ConfigValue) error {
	// Check that type is correct
	if cv.Type() != BoolType {
		return wrongType(b.key, cv.Type(), BoolType)
	}
	return nil
}

// Default returns a default value is it exists.
func (b *BoolRule) Default() ctypes.ConfigValue {
	if b.default_ != nil {
		return ctypes.ConfigValueBool{Value: *b.default_}
	}
	return nil
}

// Required indicates this rule is required.
func (b *BoolRule) Required() bool {
	return b.required
}

// Minimum returns Minimum possible value of BoolRule
func (b *BoolRule) Minimum() ctypes.ConfigValue {
	return nil
}

// Maximum returns Maximum possible value of BoolRule
func (b *BoolRule) Maximum() ctypes.ConfigValue {
	return nil
}
