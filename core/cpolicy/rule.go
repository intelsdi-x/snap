package cpolicy

// TODO, make second opts value in New<>Rule be description for rule (used in documentation)

import (
	"errors"

	"github.com/intelsdilabs/pulse/core/ctypes"
)

var (
	EmptyKeyError     = errors.New("key cannot be empty")
	WrongTypeError    = errors.New("type mismatch")
	UnderMinimumError = errors.New("value is under minimum required")
	OverMaximumError  = errors.New("value is under minimum required")
)

// A rule used to process ConfigData
type Rule interface {
	Key() string
	Validate(ctypes.ConfigValue) error
	Default() ctypes.ConfigValue
	Required() bool
}
