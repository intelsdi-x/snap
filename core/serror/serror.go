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

package serror

type SnapError interface {
	error
	Fields() map[string]interface{}
	SetFields(map[string]interface{})
}

type Fields map[string]interface{}

type snapError struct {
	err    error
	fields Fields
}

// New returns an initialized SnapError.
// The variadic signature allows fields to optionally
// be added at construction.
func New(e error, fields ...map[string]interface{}) *snapError {
	// Catch someone trying to wrap a serror around a serror.
	// We throw a panic to make them fix this.
	if _, ok := e.(SnapError); ok {
		panic("You are trying to wrap a snapError around a snapError. Don't do this.")
	}

	p := &snapError{
		err:    e,
		fields: make(map[string]interface{}),
	}

	// insert fields into new snapError
	for _, f := range fields {
		for k, v := range f {
			p.fields[k] = v
		}
	}

	return p
}

func (p *snapError) SetFields(f map[string]interface{}) {
	p.fields = f
}

func (p *snapError) Fields() map[string]interface{} {
	return p.fields
}

func (p *snapError) Error() string {
	return p.err.Error()
}

func (p *snapError) String() string {
	return p.Error()
}
