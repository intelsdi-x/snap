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

package rbody

import (
	"fmt"

	"github.com/intelsdi-x/snap/core/serror"
)

const (
	ErrorType = "error"
)

// Unsuccessful generic response to a failed API call
type Error struct {
	ErrorMessage string            `json:"message"`
	Fields       map[string]string `json:"fields"`
}

func FromSnapError(pe serror.SnapError) *Error {
	e := &Error{ErrorMessage: pe.Error(), Fields: make(map[string]string)}
	// Convert into string format
	for k, v := range pe.Fields() {
		e.Fields[k] = fmt.Sprint(v)
	}
	return e
}

func FromSnapErrors(errs []serror.SnapError) *Error {
	fields := make(map[string]string)
	var msg string
	for i, err := range errs {
		for k, v := range err.Fields() {
			fields[fmt.Sprintf("%s_err_%d", k, i)] = fmt.Sprint(v)
		}
		msg = msg + fmt.Sprintf("error %d: %s ", i, err.Error())
	}
	return &Error{
		ErrorMessage: msg,
		Fields:       fields,
	}
}

func FromError(err error) *Error {
	e := &Error{ErrorMessage: err.Error(), Fields: make(map[string]string)}
	return e
}

func (e *Error) Error() string {
	return e.ErrorMessage
}

func (e *Error) ResponseBodyMessage() string {
	return e.ErrorMessage
}

func (e *Error) ResponseBodyType() string {
	return ErrorType
}
