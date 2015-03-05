package facter

import (
	"reflect"
	"testing"

	"github.com/intelsdilabs/pulse/control/plugin"
	. "github.com/smartystreets/goconvey/convey"
)

const FACTER_NAME = "intel/facter"

func TestFacterGetMetrics(t *testing.T) {

	Convey("Facter Plugin Tests", t, func() {

		Convey("Facter plugin constants tests", func() {
			Convey("Facter name should resemble intel/facter", func() {
				So(Name, ShouldResemble, FACTER_NAME)
			})
			Convey("Facter tyoe should be plugin.CollectorPluginType", func() {
				So(Type, ShouldEqual, plugin.CollectorPluginType)
			})
		})

		Convey("GetMetricTypes tests", func() {
			facter := NewFacterPlugin()
			var pluginArgs plugin.GetMetricTypesArgs
			var reply plugin.GetMetricTypesReply
			Convey("GetMetricsTypes returns not error", func() {
				err := facter.GetMetricTypes(pluginArgs, &reply)
				So(err, ShouldBeNil)
				Convey("metricTypesReply should contain more than zero metrics", func() {
					So(len(reply.MetricTypes), ShouldBeGreaterThan, 0)
				})
				Convey("metricTypesReply contains metric namespace \"intel/facter/kernel\"", func() {
					expectedTimestamp := reply.MetricTypes[0].LastAdvertisedTimestamp()
					expectedNamespace := []string{"intel", "facter", "kernel"}
					expectedMetricType := plugin.NewMetricType(expectedNamespace, expectedTimestamp)
					//					Printf("\n expected: %v\n", expectedMetricType)
					success := false
					for idx, elem := range reply.MetricTypes {
						if reflect.DeepEqual(expectedMetricType, elem) {
							So(reply.MetricTypes[idx], ShouldResemble, expectedMetricType)
							success = true
							break
						}
					}
					if !success {
						// ShouldContain compares through pointers - SO THIS WILL FAIL
						So(reply.MetricTypes, ShouldContain, expectedMetricType)
					}

				})
			})

		})
	})
}
