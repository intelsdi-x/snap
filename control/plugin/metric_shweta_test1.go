package plugin

import (
	"testing"
	"time"
	//"bytes"
	//"encoding/gob"
	//"encoding/json"
	//"errors"
	"fmt"
	//"time"

	//"github.com/intelsdi-x/pulse/core/cdata"
	//"github.com/intelsdi-x/pulse/pkg/logger"
	. "github.com/smartystreets/goconvey/convey"
)

func TestMetricMarshal(t *testing.T) {
	if PulsePath != "" {
		Convey("Metric.go", t, func() {
			Convey("Marshalling the  GOB plugin metric type", func() {
				metrics := []PluginMetricType{
					*NewPluginMetricType([]string{"intel", "pluginName", "metricName"}, 1),
				}

				marshalData, contentTypes, err := MarshallPluginMetricTypes("pulse.*", metrics)

				So(contentTypes, ShouldResemble, "pulse.gob")
				So(err, ShouldBeNil)
				So(len(marshalData), ShouldBeGreaterThan, 0)

				Convey("Un-Marshalling GOB Metric Type", func() {
					unMarshalData, err := UnmarshallPluginMetricTypes("pulse.gob", marshalData)

					//fmt.Printf("%+v, %+v,%+v", marshalData, contentTypes, err)

					So(unMarshalData[0].Data(), ShouldResemble, int(1))
					So(err, ShouldBeNil)

				})

			})
			Convey("Marshalling the JSON plugin Metric type", func() {
				// metrics := []PluginMetricType{
				// 	*NewPluginMetricType([]string{"intel", "pluginName", "metricName"}, 1),
				// }
				metric := *NewPluginMetricType([]string{"intel", "pluginName", "metricName"}, 1)
				metric.Version_ = 1
				metric.LastAdvertisedTime_ = time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC)
				metrics := []PluginMetricType{metric, metric}

				marshalData, contentTypes, err := MarshallPluginMetricTypes("pulse.json", metrics)

				So(contentTypes, ShouldResemble, "pulse.json")
				So(err, ShouldBeNil)
				So(len(marshalData), ShouldBeGreaterThan, 0)
				So(metric.Version(), ShouldResemble, 1)
				So(metric.Namespace(), ShouldResemble, []string{"intel", "pluginName", "metricName"})
				So(metric.LastAdvertisedTime(), ShouldResemble, time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC))

				Convey("Marshallin nil Metric Type", func() {
					marshalData, contentTypes, err := MarshallPluginMetricTypes("pulse.*", nil)
					//fmt.Printf("%+v, %+v,%+v", marshalData, contentTypes, err)
					So(contentTypes, ShouldResemble, "")
					So(marshalData, ShouldResemble, []byte(nil))
					So(err.Error(), ShouldResemble, "attempt to marshall empty slice of metrics: pulse.*")

				})

				Convey("Un-Marshallin  Metric Type", func() {
					unMarshalData, err := UnmarshallPluginMetricTypes("pulse.json", marshalData)

					//fmt.Printf("%+v, %+v,%+v", marshalData, contentTypes, err)

					So(unMarshalData[0].Data(), ShouldResemble, float64(1))
					So(err, ShouldBeNil)

				})

			})
			Convey("Marshalling an unknown plugin Metric type", func() {
				metrics := []PluginMetricType{
					*NewPluginMetricType([]string{"intel", "pluginName", "metricName"}, 1),
				}

				marshalData, contentTypes, err := MarshallPluginMetricTypes("unknown_content_type", metrics)
				fmt.Printf("%+v, %+v,%+v", marshalData, contentTypes, err)
				So(contentTypes, ShouldResemble, "")
				So(err.Error(), ShouldResemble, "invlaid pulse content type: unknown_content_type")
				So(marshalData, ShouldResemble, []byte(nil))

				Convey("Un-Marshallin Unknow Metric Type", func() {
					unMarshalData, err := UnmarshallPluginMetricTypes("pulse.json", marshalData)

					//fmt.Printf("%+v, %+v,%+v", marshalData, contentTypes, err)

					So(unMarshalData[0].Data(), ShouldResemble, float64(1))
					So(err, ShouldBeNil)

				})

			})

		})
	}
}
