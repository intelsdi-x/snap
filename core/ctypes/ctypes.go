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

package ctypes

import "encoding/json"

// TODO constructors for each that have typing for value (and optionally validate)

type ConfigValue interface {
	Type() string
}

type ConfigValueInt struct {
	Value int
}

func (c ConfigValueInt) Type() string {
	return "integer"
}

func (c ConfigValueInt) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.Value)
}

type ConfigValueStr struct {
	Value string
}

func (c ConfigValueStr) Type() string {
	return "string"
}

func (c ConfigValueStr) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.Value)
}

type ConfigValueFloat struct {
	Value float64
}

func (c ConfigValueFloat) Type() string {
	return "float"
}

func (c ConfigValueFloat) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.Value)
}

type ConfigValueBool struct {
	Value bool
}

func (c ConfigValueBool) Type() string {
	return "bool"
}

func (c ConfigValueBool) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.Value)
}

// Returns a slice of string keywords for the types supported by ConfigValue.
func SupportedTypes() []string {
	// This is kind of a hack but keeps the definition of types here in
	// ctypes.go. If you create a new ConfigValue type be sure and add here
	// to return the Type() response. This will cause any depedant components
	// to acknowledge and use that type.
	t := []string{
		// String
		ConfigValueStr{}.Type(),
		// Integer
		ConfigValueInt{}.Type(),
		// Float
		ConfigValueFloat{}.Type(),
		// Bool
		ConfigValueBool{}.Type(),
	}
	return t
}
