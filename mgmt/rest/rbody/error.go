package rbody

import (
	"fmt"

	"github.com/intelsdi-x/pulse/core/perror"
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

func FromError(err error) *Error {
	e := &Error{ErrorMessage: err.Error(), Fields: make(map[string]string)}
	return e
}

func (e *Error) ResponseBodyMessage() string {
	return e.ResponseBodyMessage()
}

func (e *Error) ResponseBodyType() string {
	return "error"
}
