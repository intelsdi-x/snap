package cpolicy

// TODO, make second opts value in New<>Rule be description for rule (used in documentation)

import (
	"errors"
	"fmt"

	"github.com/intelsdilabs/pulse/core/ctypes"
)

var (
	EmptyKeyError = errors.New("key cannot be empty")
)

// A rule used to process ConfigData
type Rule interface {
	Key() string
	Validate(ctypes.ConfigValue) error
	Default() ctypes.ConfigValue
	Required() bool
}

func wrongType(key, inType, reqType string) error {
	return errors.New(fmt.Sprintf("type mismatch (%s wanted type '%s' but provided type '%s')", key, inType, reqType))
}
