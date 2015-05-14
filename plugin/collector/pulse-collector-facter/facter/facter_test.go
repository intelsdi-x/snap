// +build linux integration

/*
# testing
go test -v github.com/intelsdi-x/pulse/plugin/collector/pulse-collector-facter/facter
*/
package facter

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/intelsdi-x/pulse/control/plugin"
	. "github.com/smartystreets/goconvey/convey"
)

// fact expected to be available on every system
// can be allways received from Facter for test purposes
const someFact = "kernel"
const someValue = "linux 1234"

var existingNamespace = []string{vendor, prefix, someFact}

func TestFacterCollectMetrics(t *testing.T) {
	Convey("TestFacterCollect tests", t, func() {

		f := NewFacter()
		// always return at least one metric
		f.getFacts = func(_ []string, _ time.Duration, _ *cmdConfig) (facts, error) {
			return facts{someFact: someValue}, nil
		}

		Convey("asked for nothgin returns nothing", func() {
			metricTypes := []plugin.PluginMetricType{}
			metrics, err := f.CollectMetrics(metricTypes)
			So(err, ShouldBeNil)
			So(metrics, ShouldBeEmpty)
		})

		Convey("asked for somehting returns something", func() {
			metricTypes := []plugin.PluginMetricType{
				plugin.PluginMetricType{
					Namespace_: existingNamespace,
				},
			}
			metrics, err := f.CollectMetrics(metricTypes)
			So(err, ShouldBeNil)
			So(metrics, ShouldNotBeEmpty)
			So(len(metrics), ShouldEqual, 1)

			// check just one metric
			metric := metrics[0]
			So(metric.Namespace()[2], ShouldResemble, someFact)
			So(metric.Data().(string), ShouldEqual, someValue)
		})

		Convey("ask for inappriopriate metrics", func() {
			Convey("wrong number of parts", func() {
				_, err := f.CollectMetrics(
					[]plugin.PluginMetricType{
						plugin.PluginMetricType{
							Namespace_: []string{"where are my other parts"},
						},
					},
				)
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldContainSubstring, "segments")
			})

			Convey("wrong vendor", func() {
				_, err := f.CollectMetrics(
					[]plugin.PluginMetricType{
						plugin.PluginMetricType{
							Namespace_: []string{"nonintelvendor", prefix, someFact},
						},
					},
				)
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldContainSubstring, "expected vendor")
			})

			Convey("wrong prefix", func() {
				_, err := f.CollectMetrics(
					[]plugin.PluginMetricType{
						plugin.PluginMetricType{
							Namespace_: []string{vendor, "this is wrong prefix", someFact},
						},
					},
				)
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldContainSubstring, "expected prefix")
			})

		})
	})
}

func TestFacterInvalidBehavior(t *testing.T) {

	Convey("returns errors as expected when cmd isn't working", t, func() {
		f := NewFacter()
		// mock that getFacts returns error every time
		f.getFacts = func(_ []string, _ time.Duration, _ *cmdConfig) (facts, error) {
			return nil, errors.New("dummy error")
		}

		_, err := f.CollectMetrics([]plugin.PluginMetricType{
			plugin.PluginMetricType{
				Namespace_: existingNamespace,
			},
		},
		)
		So(err, ShouldNotBeNil)

		_, err = f.GetMetricTypes()
		So(err, ShouldNotBeNil)
	})
	Convey("returns not as much values as asked", t, func() {

		f := NewFacter()
		// mock that getFacts returns error every time
		//returns zero elements even when asked for one
		f.getFacts = func(_ []string, _ time.Duration, _ *cmdConfig) (facts, error) {
			return nil, nil
		}

		_, err := f.CollectMetrics([]plugin.PluginMetricType{
			plugin.PluginMetricType{
				Namespace_: existingNamespace,
			},
		},
		)
		So(err, ShouldNotBeNil)
		So(err.Error(), ShouldContainSubstring, "more/less")
	})

}

func TestFacterGetMetricsTypes(t *testing.T) {

	Convey("GetMetricTypes functionallity", t, func() {

		f := NewFacter()

		Convey("GetMetricsTypes returns no error", func() {
			// exectues without error
			metricTypes, err := f.GetMetricTypes()
			So(err, ShouldBeNil)
			Convey("metricTypesReply should contain more than zero metrics", func() {
				So(metricTypes, ShouldNotBeEmpty)
			})

			Convey("at least one metric contains metric namespace \"intel/facter/kernel\"", func() {

				expectedNamespaceStr := strings.Join(existingNamespace, "/")

				found := false
				for _, metricType := range metricTypes {
					// join because we cannot compare slices
					if strings.Join(metricType.Namespace(), "/") == expectedNamespaceStr {
						found = true
						break
					}
				}
				if !found {
					t.Error("It was expected to find at least on intel/facter/kernel metricType (but it wasn't there)")
				}
			})
		})
	})
}
