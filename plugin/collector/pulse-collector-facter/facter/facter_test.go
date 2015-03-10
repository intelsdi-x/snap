/*
# testing
go test -v github.com/intelsdilabs/pulse/plugin/collector/pulse-collector-facter/facter
*/
package facter

import (
	"reflect"
	"testing"
	"time"

	"github.com/intelsdilabs/pulse/control/plugin"
	. "github.com/smartystreets/goconvey/convey"
)

// allows to create fake facter
func withFakeFacter(facter *Facter, output stringmap, f func()) func() {

	// getFactsMock
	getFactsMock := func(keys []string, facterTimeout time.Duration) (*stringmap, *time.Time, error) {
		now := time.Now()
		return &output, &now, nil
	}

	return func() {
		// set mock
		facter.getFacts = getFactsMock
		// set reset function to restore original version of getFacts
		Reset(func() {
			facter.getFacts = getFacts
		})
		f()
	}
}

// TODO:
func TestCacheUpdate(t *testing.T) {

	// enough time to be treaeted as stale value
	longAgo := time.Now().Add(-(2 * DefaultCacheTTL))

	Convey("Facter cache update works", t, func() {

		f := NewFacterPlugin()

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

		f := NewFacterPlugin()

		Convey("empty cache always refresh given value", func() {
			// make sure is empty
			So(f.cache, ShouldBeEmpty)
			namesToUpdate := f.getNamesToUpdate([]string{"foo"})
			So(namesToUpdate, ShouldContain, "foo")
		})

		Convey("existing fresh key needn't be refreshed", func() {
			f.cache["foo"] = fact{value: 1, lastUpdate: time.Now()}
			namesToUpdate := f.getNamesToUpdate([]string{"foo"})
			So(namesToUpdate, ShouldNotContain, "foo")
		})

		Convey("stale key need to be refreshed", func() {
			// add stale key
			f.cache["foo"] = fact{value: 1, lastUpdate: longAgo}

			namesToUpdate := f.getNamesToUpdate([]string{"foo"})
			So(namesToUpdate, ShouldContain, "foo")
		})

	})

	Convey("cache synchronization", t, func() {

		f := NewFacterPlugin()

		// Convey("when not synchronized cache is empty", func() {
		// 	// make sure it is empty
		// 	So(f.cache, ShouldBeEmpty)
		// 	err := f.synchronizeCache([]string{"bar"})
		// 	So(err, ShouldBeNil)
		// 	fact, exists := f.cache["bar"]
		// 	So(exists, ShouldEqual, true)
		// 	So(fact.value, ShouldBeNil) // because there is no such value in factor, we have nil
		// })

		Convey("cache value with faked facter for foo", withFakeFacter(f, stringmap{"foo": 1}, func() {
			err := f.synchronizeCache([]string{"foo"})
			So(err, ShouldBeNil)
			fact, exists := f.cache["foo"]
			So(exists, ShouldEqual, true)
			So(fact.value, ShouldEqual, 1)

			// Convey("cache which not need to be resynchronized", withFakeFacter(f, stringmap{"foo": 2}, func() {
			// 	err := f.synchronizeCache([]string{"foo"})
			// 	So(err, ShouldBeNil)
			// 	fact, _ := f.cache["foo"]
			// 	So(fact.value, ShouldEqual, 1) // still returns 1
			// }))

			Convey("cache which needs to be resynchronized", withFakeFacter(f, stringmap{"foo": 2}, func() {

				// invalidate value in cache by overriding lastUpdate field
				fact := f.cache["foo"]
				fact.lastUpdate = longAgo
				So(fact.value, ShouldEqual, 1) // still because of outer convey
				f.cache["foo"] = fact

				// synchronize and check
				err := f.synchronizeCache([]string{"foo"})
				fact = f.cache["foo"]
				So(err, ShouldBeNil)
				So(fact.value, ShouldEqual, 2) // still returns 2
			}))

		}))

		//
		// Convey("asked for no existing fact", func() {
		//
		// })

	})

}

func TestFacterGetMetrics(t *testing.T) {

	// TODO:not implemented! - fullfill GetMetricTypes
	Convey("GetMetricTypes tests", t, func() {

		facter := NewFacterPlugin()
		var pluginArgs plugin.GetMetricTypesArgs
		var reply plugin.GetMetricTypesReply
		Convey("GetMetricsTypes returns no error", func() {
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
}

func TestFacterCollect(t *testing.T) {
	// TODO: time outs after 5 seconds because of goroutine
	Convey("TestFacterCollect tests", t, func() {

		f := NewFacterPlugin()
		Convey("update ache", func() {
			f.synchronizeCache([]string{"foo"})
		})

		Convey("Collect returns nil", func() {
			facter := NewFacterPlugin()
			var pluginArgs plugin.CollectorArgs
			var reply plugin.CollectorReply
			So(facter.Collect(pluginArgs, &reply), ShouldBeNil)
		})
	})
}

func TestFacterPluginMeta(t *testing.T) {
	Convey("PluginMeta tests", t, func() {
		meta := Meta()
		Convey("Meta is not nil", func() {
			So(meta, ShouldNotBeNil)
		})
		Convey("Name should be Intel Facter Plugin (c) 2015 Intel Corporation", func() {
			So(meta.Name, ShouldResemble, "Intel Facter Plugin (c) 2015 Intel Corporation")
		})
		Convey("Version should be 1", func() {
			So(meta.Version, ShouldEqual, 1)
		})
		Convey("Type should be plugin.CollectorPluginType", func() {
			So(meta.Type, ShouldResemble, plugin.CollectorPluginType)
		})
	})
}

func TestFacterConfigPolicy(t *testing.T) {
	Convey("TestFacterConfigPolicy tests", t, func() {
		Convey("TestFacterConfigPolicy returns proper object", func() {
			pluginPolicy := new(plugin.ConfigPolicy)
			So(ConfigPolicy(), ShouldResemble, pluginPolicy)
		})
	})
}
