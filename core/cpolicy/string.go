package cpolicy

import (
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

func (s *stringRule) Default() ctypes.ConfigValue {
	if s.default_ != nil {
		return &ctypes.ConfigValueStr{Value: *s.default_}
	}
	return nil
}

func (s *stringRule) Required() bool {
	return s.required
}
