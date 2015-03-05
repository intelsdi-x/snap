package facter

import (
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
			var metricTypesReply plugin.GetMetricTypesReply
			Convey("GetMetrics test", func() {
				err := facter.GetMetricTypes(pluginArgs, &metricTypesReply)
				So(err, ShouldBeNil)
			})
		})
	})
}
