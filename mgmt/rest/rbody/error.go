package rbody

import (
	"fmt"

	"github.com/intelsdi-x/pulse/core/perror"
)

const (
	ErrorType = "error"
)

// Unsuccessful generic response to a failed API call
type Error struct {
	ErrorMessage string            `json:"message"`
	Fields       map[string]string `json:"fields"`
}

func FromPulseError(pe perror.PulseError) *Error {
	e := &Error{ErrorMessage: pe.Error(), Fields: make(map[string]string)}
	// Convert into string format
	for k, v := range pe.Fields() {
		e.Fields[k] = fmt.Sprint(v)
	}
	return e
}

func FromPulseErrors(errs []perror.PulseError) *Error {
	fields := make(map[string]string)
	var msg string
	for i, err := range errs {
		for k, v := range err.Fields() {
			fields[fmt.Sprintf("%s_err_%d", k, i)] = fmt.Sprint(v)
		}
		msg = msg + fmt.Sprintf("message @ error %d: %s ", i, err.Error())
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
