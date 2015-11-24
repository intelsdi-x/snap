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

// TODO, make second opts value in New<>Rule be description for rule (used in documentation)

import (
	"errors"
	"fmt"

	"github.com/intelsdi-x/snap/core/ctypes"
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
	Type() string
	Minimum() ctypes.ConfigValue
	Maximum() ctypes.ConfigValue
}

type rule struct {
	Description string
}

func wrongType(key, inType, reqType string) error {
	return errors.New(fmt.Sprintf("type mismatch (%s wanted type '%s' but provided type '%s')", key, reqType, inType))
}
