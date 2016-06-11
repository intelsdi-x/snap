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

	"github.com/intelsdi-x/snap/core/ctypes"
)

const (
	StringType = "string"
)

// A rule validating against string-typed config
type StringRule struct {
	rule

	key      string
	required bool
	default_ *string
}

// Returns a new string-typed rule. Arguments are key(string), required(bool), default(string).
func NewStringRule(key string, req bool, opts ...string) (*StringRule, error) {
	// Return error if key is empty
	if key == "" {
		return nil, EmptyKeyError
	}

	var def *string
	if len(opts) > 0 {
		def = &opts[0]
	}

	return &StringRule{
		key:      key,
		required: req,
		default_: def,
	}, nil
}

func (s *StringRule) Type() string {
	return StringType
}

// MarshalJSON marshals a StringRule into JSON
func (s *StringRule) MarshalJSON() ([]byte, error) {
	return json.Marshal(&struct {
		Key      string             `json:"key"`
		Required bool               `json:"required"`
		Default  ctypes.ConfigValue `json:"default"`
		Type     string             `json:"type"`
	}{
		Key:      s.key,
		Required: s.required,
		Default:  s.Default(),
		Type:     StringType,
	})
}

//GobEncode encodes a StringRule in to a GOB
func (s *StringRule) GobEncode() ([]byte, error) {
	w := new(bytes.Buffer)
	encoder := gob.NewEncoder(w)
	if err := encoder.Encode(s.key); err != nil {
		return nil, err
	}
	if err := encoder.Encode(s.required); err != nil {
		return nil, err
	}
	if s.default_ == nil {
		encoder.Encode(false)
	} else {
		encoder.Encode(true)
		if err := encoder.Encode(&s.default_); err != nil {
			return nil, err
		}
	}
	return w.Bytes(), nil
}

// GobDecode decodes a GOB into a StringRule
func (s *StringRule) GobDecode(buf []byte) error {
	r := bytes.NewBuffer(buf)
	decoder := gob.NewDecoder(r)
	if err := decoder.Decode(&s.key); err != nil {
		return err
	}
	if err := decoder.Decode(&s.required); err != nil {
		return err
	}
	var is_default_set bool
	decoder.Decode(&is_default_set)
	if is_default_set {
		return decoder.Decode(&s.default_)
	}
	return nil
}

// Returns the key
func (s *StringRule) Key() string {
	return s.key
}

// Validates a config value against this rule.
func (s *StringRule) Validate(cv ctypes.ConfigValue) error {
	// Check that type is correct
	if cv.Type() != StringType {
		return wrongType(s.key, cv.Type(), StringType)
	}
	return nil
}

// Returns a default value is it exists.
func (s *StringRule) Default() ctypes.ConfigValue {
	if s.default_ != nil {
		return ctypes.ConfigValueStr{Value: *s.default_}
	}
	return nil
}

// Indicates this rule is required.
func (s *StringRule) Required() bool {
	return s.required
}

func (s *StringRule) Minimum() ctypes.ConfigValue {
	return nil
}

func (s *StringRule) Maximum() ctypes.ConfigValue {
	return nil
}
