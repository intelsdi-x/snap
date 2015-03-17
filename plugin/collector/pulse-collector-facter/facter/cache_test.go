package facter

import (
	"reflect"
	"testing"
	"time"

	"github.com/intelsdilabs/pulse/control/plugin"
	. "github.com/smartystreets/goconvey/convey"
)

func TestCacheUpdate(t *testing.T) {

	// enough time to be treaeted as stale value in cache
	longAgo := time.Now().Add(-(2 * defaultCacheTTL))

	Convey("facter cache update works at all", t, func() {

		f := NewFacter()

		Convey("empty for start", func() {
			So(f.cache, ShouldBeEmpty)
		})

		Convey("filled after first updateCacheAll", func() {
			err := f.updateCacheAll()
			So(err, ShouldBeNil)
			So(f.cache, ShouldNotBeEmpty)
		})

	})

	Convey("cache update policy", t, func() {

		f := NewFacter()

		Convey("missing value always force refresh", func() {
			// make sure is empty
			So(f.cache, ShouldBeEmpty)
			namesToUpdate := f.getNamesToUpdate([]string{"foo"})
			// we exepct not to be empty
			So(namesToUpdate, ShouldContain, "foo")
		})

		Convey("existing fresh key needn't be refreshed", func() {
			// add fresh key
			f.cache["foo"] = entry{value: 1, lastUpdate: time.Now()}

			// find out what's need to be refreshed
			namesToUpdate := f.getNamesToUpdate([]string{"foo"})
			So(namesToUpdate, ShouldNotContain, "foo")
		})

		Convey("stale key need to be refreshed", func() {
			// add stale key
			f.cache["foo"] = entry{value: 1, lastUpdate: longAgo}

			// find out what's need to be refreshed
			namesToUpdate := f.getNamesToUpdate([]string{"foo"})
			So(namesToUpdate, ShouldContain, "foo")
		})

	})

	Convey("cache synchronization", t, func() {

		f := NewFacter()

		Convey("when not synchronized cache is empty and asked for existing fact",
			withFakeFacter(f, facts{"foo": 1}, func() {
				// make sure it is empty
				So(f.cache, ShouldBeEmpty)
				err := f.synchronizeCache([]string{"foo"})
				So(err, ShouldBeNil)
				fact, exists := f.cache["foo"]
				So(exists, ShouldEqual, true)
				So(fact.value, ShouldEqual, 1) // because there is no such value in factor, we have nil
			}))

		Convey("cache value with faked facter for foo",
			withFakeFacter(f, facts{"foo": 1}, func() {
				err := f.synchronizeCache([]string{"foo"})
				So(err, ShouldBeNil)
				fact, exists := f.cache["foo"]
				So(exists, ShouldEqual, true)
				So(fact.value, ShouldEqual, 1)

				Convey("cache which not need to be resynchronized",
					withFakeFacter(f, facts{"foo": 2}, func() {
						err := f.synchronizeCache([]string{"foo"})
						So(err, ShouldBeNil)
						fact, _ := f.cache["foo"]
						So(fact.value, ShouldEqual, 1) // still returns 1
					}))

				Convey("cache which needs to be resynchronized",
					withFakeFacter(f, facts{"foo": 2}, func() {

						// invalidate value in cache by overriding lastUpdate field
						fact := f.cache["foo"]
						fact.lastUpdate = longAgo
						So(fact.value, ShouldEqual, 1) // still 1 because already set in outer convey
						f.cache["foo"] = fact

						// synchronize and check
						err := f.synchronizeCache([]string{"foo"})
						fact = f.cache["foo"]
						So(err, ShouldBeNil)
						So(fact.value, ShouldEqual, 2) // still returns 2
					}))
			}))

		// what about that is having nil returned to pulse is good way to handle this ?
		Convey("refresh for no available metric - stores nil in cache",
			withFakeFacter(f, facts{}, func() {
				err := f.synchronizeCache([]string{"foo"})
				So(err, ShouldBeNil)
				fact, exists := f.cache["foo"]
				So(exists, ShouldEqual, true)
				So(fact.value, ShouldBeNil)
			}))

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
