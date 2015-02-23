package cpolicy

// TODO, make second opts value in New<>Rule be description for rule (used in documentation)

import (
	"errors"
	"github.com/intelsdilabs/pulse/core/ctypes"
)

var (
	EmptyKeyError  = errors.New("key cannot be empty")
	WrongTypeError = errors.New("type mismatch")
)

// A rule used to process ConfigData
type Rule interface {
	Key() string
	Validate(ctypes.ConfigValue) error
}

// A rule validating against string-typed config
type stringRule struct {
	key      string
	required bool
	default_ *string
}

// Returns a new string-typed rule. Arguments are key(string), required(bool), default(string).
func NewStringRule(key string, req bool, opts ...string) (Rule, error) {
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

// Returns the key
func (s *stringRule) Key() string {
	return s.key
}

// Validates a config value against this rule.
func (s *stringRule) Validate(cv ctypes.ConfigValue) error {
	// Check that type is correct
	if cv.Type() != "string" {
		return WrongTypeError
	}
	return nil
}
