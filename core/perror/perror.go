/*
http://www.apache.org/licenses/LICENSE-2.0.txt


Copyright 2015 Intel Coporation

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

package perror

type PulseError interface {
	error
	Fields() map[string]interface{}
	SetFields(map[string]interface{})
}

type Fields map[string]interface{}

type pulseError struct {
	err    error
	fields Fields
}

// New returns an initialized PulseError.
// The variadic signature allows fields to optionally
// be added at construction.
func New(e error, fields ...map[string]interface{}) *pulseError {
	// Catch someone trying to wrap a pe around a pe.
	// We throw a panic to make them fix this.
	if _, ok := e.(PulseError); ok {
		panic("You are trying to wrap a pulseError around a PulseError. Don't do this.")
	}

	p := &pulseError{
		err:    e,
		fields: make(map[string]interface{}),
	}

	// insert fields into new PulseError
	for _, f := range fields {
		for k, v := range f {
			p.fields[k] = v
		}
	}

	return p
}

func (p *pulseError) SetFields(f map[string]interface{}) {
	p.fields = f
}

func (p *pulseError) Fields() map[string]interface{} {
	return p.fields
}

func (p *pulseError) Error() string {
	return p.err.Error()
}

func (p *pulseError) String() string {
	return p.Error()
}
