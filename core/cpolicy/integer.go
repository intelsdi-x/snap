package cpolicy

import (
	"github.com/intelsdilabs/pulse/core/ctypes"
)

// A rule validating against string-typed config
type intRule struct {
	key      string
	required bool
	default_ *int
	minimum  *int
	maximum  *int
}

// Returns a new int-typed rule. Arguments are key(string), required(bool), default(int), min(int), max(int)
func NewIntegerRule(key string, req bool, opts ...int) (*intRule, error) {
	// Return error if key is empty
	if key == "" {
		return nil, EmptyKeyError
	}

	options := make([]*int, 1)
	for i, o := range opts {
		options[i] = &o
	}

	return &intRule{
		key:      key,
		required: req,
		default_: options[0],
	}, nil
}

// Returns the key
func (i *intRule) Key() string {
	return i.key
}

// Validates a config value against this rule.
func (i *intRule) Validate(cv ctypes.ConfigValue) error {
	// Check that type is correct
	if cv.Type() != "integer" {
		return WrongTypeError
	}
	// Check minimum. Type should be safe now because of the check above.
	if i.minimum != nil && cv.(ctypes.ConfigValueInt).Value < *i.minimum {
		return UnderMinimumError
	}
	// Check maximum. Type should be safe now because of the check above.
	if i.maximum != nil && cv.(ctypes.ConfigValueInt).Value > *i.maximum {
		return OverMaximumError
	}
	return nil
}

func (i *intRule) Default() ctypes.ConfigValue {
	if i.default_ != nil {
		return &ctypes.ConfigValueInt{Value: *i.default_}
	}
	return nil
}

func (i *intRule) Required() bool {
	return i.required
}

func (i *intRule) SetMinimum(m int) {
	i.minimum = &m
}

func (i *intRule) SetMaximum(m int) {
	i.maximum = &m
}
