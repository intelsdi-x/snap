package opentsdb

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestValid(t *testing.T) {
	dataPoint := DataPoint{"Temperature", 1442028635, 23.1, map[string]StringValue{"host": "abc"}}

	Convey("Valid assertions", t, func() {
		So(dataPoint.Valid(), ShouldBeTrue)
	})
}

func TestInvalid_1(t *testing.T) {
	dataPoint := DataPoint{}

	Convey("Invalid assertions", t, func() {
		So(dataPoint.Valid(), ShouldBeFalse)

		dataPoint = DataPoint{
			Metric: "test",
			Value:  123,
		}
		So(dataPoint.Valid(), ShouldBeFalse)

		dataPoint = DataPoint{
			Metric:    "test",
			Value:     123,
			Timestamp: 12345,
		}
		So(dataPoint.Valid(), ShouldBeFalse)

		dataPoint := DataPoint{
			Metric: "test",
			Value:  123,
			Tags: map[string]StringValue{
				"host": "abc",
			},
		}
		So(dataPoint.Valid(), ShouldBeFalse)

		dataPoint = DataPoint{
			Value:     123,
			Timestamp: 12345,
			Tags: map[string]StringValue{
				"host": "abc",
			},
		}
		So(dataPoint.Valid(), ShouldBeFalse)
	})
}
