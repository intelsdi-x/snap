/*
# testing
go test -v github.com/intelsdilabs/pulse/plugin/collector/pulse-collector-facter/facter
*/
package facter

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/intelsdilabs/pulse/control/plugin"
	. "github.com/smartystreets/goconvey/convey"
)

// fact expected to be available on every system
// can be allways received from Facter for test purposes
const existingFact = "kernel"

var existingNamespace = []string{vendor, prefix, existingFact}

// allows to use fake facter within tests - that returns error every time
func withBrokenFacter(facter *Facter, f func()) func() {

	// getFactsMock
	getFactsMock := func(_ []string, _ time.Duration, _ *cmdConfig) (*facts, *time.Time, error) {
		return nil, nil, errors.New("dummy error")
	}

	return func() {
		// set mock
		facter.metricCache.getFacts = getFactsMock
		// set reset function to restore original version of getFacts
		Reset(func() {
			facter.metricCache.getFacts = getFacts
		})
		f()
	}
}

func TestFacterCollectMetrics(t *testing.T) {
	Convey("TestFacterCollect tests", t, func() {

		Convey("asked for nothgin returns nothing", func() {
			f := NewFacter()
			metricTypes := []plugin.PluginMetricType{}
			metrics, err := f.CollectMetrics(metricTypes)
			So(err, ShouldBeNil)
			So(metrics, ShouldBeEmpty)
		})

		Convey("asked for somehting returns somthing", func() {
			f := NewFacter()
			metricTypes := []plugin.PluginMetricType{
				*plugin.NewPluginMetricType(
					existingNamespace,
				),
			}
			metrics, err := f.CollectMetrics(metricTypes)
			So(err, ShouldBeNil)
			So(metrics, ShouldNotBeEmpty)
		})

		Convey("ask for inappriopriate metrics", func() {
			f := NewFacter()
			Convey("wrong number of parts", func() {
				_, err := f.CollectMetrics(
					[]plugin.PluginMetricType{
						*plugin.NewPluginMetricType(
							[]string{"where are my other parts"},
						),
					},
				)
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldContainSubstring, "segments")
			})

			Convey("wrong vendor", func() {
				_, err := f.CollectMetrics(
					[]plugin.PluginMetricType{
						*plugin.NewPluginMetricType(
							[]string{"nonintelvendor", prefix, existingFact},
						),
					},
				)
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldContainSubstring, "expected vendor")
			})

			Convey("wrong prefix", func() {
				_, err := f.CollectMetrics(
					[]plugin.PluginMetricType{
						*plugin.NewPluginMetricType(
							[]string{vendor, "this is wrong prefix", existingFact},
						),
					},
				)
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldContainSubstring, "expected prefix")
			})

		})
	})
}

func TestFacterReturnsErrors(t *testing.T) {

	f := NewFacter()

	Convey("returns errors as expected when cmd isn't working", t, withBrokenFacter(f, func() {

		_, err := f.CollectMetrics([]plugin.PluginMetricType{
			*plugin.NewPluginMetricType(
				existingNamespace,
			),
		},
		)
		So(err, ShouldNotBeNil)

		_, err = f.GetMetricTypes()
		So(err, ShouldNotBeNil)
	}))
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
