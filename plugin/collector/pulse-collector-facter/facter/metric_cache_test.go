package facter

import (
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
)

// allows to use fake facter within tests
func withFakeFacter(c *metricCache, mockFacts facts, f func()) func() {

	// getFactsMock
	getFactsMock := func(names []string, _ time.Duration, _ *cmdConfig) (*facts, *time.Time, error) {
		now := time.Now()
		return &mockFacts, &now, nil
	}

	return func() {
		// set mock
		c.getFacts = getFactsMock
		// set reset function to restore original version of getFacts
		Reset(func() {
			c.getFacts = getFacts
		})
		f()
	}
}

func TestCacheUpdate(t *testing.T) {

	// enough time to be treaeted as stale value in cache
	longAgo := time.Now().Add(-(2 * defaultCacheTTL))

	Convey("facter cache update works at all", t, func() {

		c := newMetricCache(defaultCacheTTL, defaultFacterDeadline)

		Convey("empty for start", func() {
			So(c.data, ShouldBeEmpty)
		})

		Convey("filled after first updateCacheAll", func() {
			err := c.updateCacheAll()
			So(err, ShouldBeNil)
			So(c.data, ShouldNotBeEmpty)
		})

	})

	Convey("cache update policy", t, func() {

		c := newMetricCache(defaultCacheTTL, defaultFacterDeadline)

		Convey("missing value always force refresh", func() {
			// make sure is empty
			So(c.data, ShouldBeEmpty)
			namesToUpdate := c.getNamesToUpdate([]string{"foo"})
			// we exepct not to be empty
			So(namesToUpdate, ShouldContain, "foo")
		})

		Convey("existing fresh key needn't be refreshed", func() {
			// add fresh key
			c.data["foo"] = entry{value: 1, lastUpdate: time.Now()}

			// find out what's need to be refreshed
			namesToUpdate := c.getNamesToUpdate([]string{"foo"})
			So(namesToUpdate, ShouldNotContain, "foo")
		})

		Convey("stale key need to be refreshed", func() {
			// add stale key
			c.data["foo"] = entry{value: 1, lastUpdate: longAgo}

			// find out what's need to be refreshed
			namesToUpdate := c.getNamesToUpdate([]string{"foo"})
			So(namesToUpdate, ShouldContain, "foo")
		})

	})

	Convey("cache synchronization", t, func() {

		c := newMetricCache(defaultCacheTTL, defaultFacterDeadline)

		Convey("when not synchronized cache is empty and asked for existing fact",
			withFakeFacter(c, facts{"foo": 1}, func() {
				// make sure it is empty
				So(c.data, ShouldBeEmpty)
				err := c.synchronizeCache([]string{"foo"})
				So(err, ShouldBeNil)
				fact, exists := c.data["foo"]
				So(exists, ShouldEqual, true)
				So(fact.value, ShouldEqual, 1) // because there is no such value in factor, we have nil
			}))

		Convey("cache value with faked facter for foo",
			withFakeFacter(c, facts{"foo": 1}, func() {
				err := c.synchronizeCache([]string{"foo"})
				So(err, ShouldBeNil)
				fact, exists := c.data["foo"]
				So(exists, ShouldEqual, true)
				So(fact.value, ShouldEqual, 1)

				Convey("cache which not need to be resynchronized",
					withFakeFacter(c, facts{"foo": 2}, func() {
						err := c.synchronizeCache([]string{"foo"})
						So(err, ShouldBeNil)
						fact, _ := c.data["foo"]
						So(fact.value, ShouldEqual, 1) // still returns 1
					}))

				Convey("cache which needs to be resynchronized",
					withFakeFacter(c, facts{"foo": 2}, func() {

						// invalidate value in cache by overriding lastUpdate field
						fact := c.data["foo"]
						fact.lastUpdate = longAgo
						So(fact.value, ShouldEqual, 1) // still 1 because already set in outer convey
						c.data["foo"] = fact

						// synchronize and check
						err := c.synchronizeCache([]string{"foo"})
						fact = c.data["foo"]
						So(err, ShouldBeNil)
						So(fact.value, ShouldEqual, 2) // still returns 2
					}))
			}))

		// what about that is having nil returned to pulse is good way to handle this ?
		Convey("refresh for no available metric - stores nil in cache",
			withFakeFacter(c, facts{}, func() {
				err := c.synchronizeCache([]string{"foo"})
				So(err, ShouldBeNil)
				fact, exists := c.data["foo"]
				So(exists, ShouldEqual, true)
				So(fact.value, ShouldBeNil)
			}))

	})

}
