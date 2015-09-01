package opentsdb

import (
	"fmt"
	"strconv"
)

const (
	EmptyString = ""
)

type DataPoint struct {
	Metric    StringValue            `json:"metric"`
	Timestamp int64                  `json:"timestamp"`
	Value     interface{}            `json:"value"`
	Tags      map[string]StringValue `json:"tags"`
}

// Valid verifies the mandatory fields of the Datapoint.
func (d *DataPoint) Valid() bool {
	if d.Metric == EmptyString || d.Value == nil || d.Timestamp == 0 || len(d.Tags) == 0 {
		return false
	}

	if _, err := strconv.ParseFloat(fmt.Sprint(d.Value), 64); err != nil {
		return false
	}
	return true
}
