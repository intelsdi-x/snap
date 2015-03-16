package cpolicy

import (
	"bytes"
	"encoding/gob"

	"github.com/intelsdilabs/pulse/core/ctypes"
)

// A rule validating against string-typed config
type stringRule struct {
	key      string
	required bool
	default_ *string
}

// Returns a new string-typed rule. Arguments are key(string), required(bool), default(string).
func NewStringRule(key string, req bool, opts ...string) (*stringRule, error) {
	// Return error if key is empty
	if key == "" {
		return nil, EmptyKeyError
	}

	var def *string
	if len(opts) > 0 {
		def = &opts[0]
	}

	return &stringRule{
		key:      key,
		required: req,
		default_: def,
	}, nil
}

func (s *stringRule) GobEncode() ([]byte, error) {
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

func (s *stringRule) GobDecode(buf []byte) error {
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
func (s *stringRule) Key() string {
	return s.key
}

// Validates a config value against this rule.
func (s *stringRule) Validate(cv ctypes.ConfigValue) error {
	// Check that type is correct
	if cv.Type() != "string" {
		return wrongType(s.key, cv.Type(), "string")
	}
	return nil
}

// Returns a default value is it exists.
func (s *stringRule) Default() ctypes.ConfigValue {
	if s.default_ != nil {
		return &ctypes.ConfigValueStr{Value: *s.default_}
	}
	return nil
}

// Indicates this rule is required.
func (s *stringRule) Required() bool {
	return s.required
}
