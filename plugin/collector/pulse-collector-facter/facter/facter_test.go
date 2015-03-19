/*
# testing
go test -v github.com/intelsdilabs/pulse/plugin/collector/pulse-collector-facter/facter
*/
package facter

import (
	"reflect"
	"testing"

	"github.com/intelsdilabs/pulse/control/plugin"
	"github.com/intelsdilabs/pulse/control/plugin/cpolicy"
	. "github.com/smartystreets/goconvey/convey"
)

func TestFacterCollect(t *testing.T) {
	Convey("TestFacterCollect tests", t, func() {

		Convey("Collect executes error for empty request", func() {
			f := NewFacter()
			// ok. even for emtyp request ?
			metricTypes := []plugin.PluginMetricType{}
			metrics, err := f.CollectMetrics(metricTypes)
			So(err, ShouldBeNil)
			So(metrics, ShouldBeEmpty)

		})
	})
}

func TestFacterGetMetrics(t *testing.T) {

	Convey("GetMetricTypes tests", t, func() {

		f := NewFacter()

		Convey("GetMetricsTypes returns no error", func() {
			// exectues without error
			metricTypes, err := f.GetMetricTypes()
			So(err, ShouldBeNil)

			Convey("metricTypesReply should contain more than zero metrics", func() {
				So(metricTypes, ShouldNotBeEmpty)
			})

			Convey("metricTypesReply contains metric namespace \"intel/facter/kernel\"", func() {

				// we exepect that all metric has the same advertised timestamp (because its get together)
				expectedNamespace := []string{"intel", "facter", "kernel"}

				// we are looking for this one in reply
				expectedMetricType := plugin.NewPluginMetricType(expectedNamespace)

				found := false
				for _, metricType := range metricTypes {
					if reflect.DeepEqual(expectedMetricType, metricType) {
						found = true
						break
					}
				}
				if !found {
					t.Error("It was expected to find intel/facter/kernel metricType (but it wasn't there)")
				}
			})
		})
	})
}

func TestFacterPluginMeta(t *testing.T) {
	Convey("PluginMeta tests", t, func() {
		meta := Meta()
		Convey("Meta is not nil", func() {
			So(meta, ShouldNotBeNil)
		})
		Convey("Name should be right", func() {
			So(meta.Name, ShouldEqual, "Intel Fact Gathering Plugin")
		})
		Convey("Version should be 1", func() {
			So(meta.Version, ShouldEqual, 1)
		})
		Convey("Type should be plugin.CollectorPluginType", func() {
			So(meta.Type, ShouldEqual, plugin.CollectorPluginType)
		})
	})
}

func TestFacterConfigPolicy(t *testing.T) {
	Convey("config policy has right type", t, func() {
		expectedCPT := ConfigPolicyTree()
		gotCPT := cpolicy.NewTree()
		So(expectedCPT, ShouldHaveSameTypeAs, gotCPT)
	})
}
