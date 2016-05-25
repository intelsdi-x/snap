// Package jsonutil provides a collection of types implementing the json.Unmarshaler and json.Marshaler interface.
package jsonutil

import (
	"encoding/json"
	"time"
)

// Duration is a wrapper around time.Duration which implements json.Unmarshaler and json.Marshaler.
// It marshals and unmarshals the duration as a string in the format accepted by time.ParseDuration and returned by time.Duration.String.
type Duration struct {
	time.Duration
}

// MarshalJSON implements the json.Marshaler interface. The duration is a quoted-string in the format accepted by time.ParseDuration and returned by time.Duration.String.
func (d Duration) MarshalJSON() ([]byte, error) {
	return []byte(`"` + d.String() + `"`), nil
}

// UnmarshalJSON implements the json.Unmarshaler interface. The duration is expected to be a quoted-string of a duration in the format accepted by time.ParseDuration.
func (d *Duration) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}

	tmp, err := time.ParseDuration(s)
	if err != nil {
		return err
	}

	d.Duration = tmp

	return nil
}
