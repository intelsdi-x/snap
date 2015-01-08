package control

import (
	"sync"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestSubscriptions(t *testing.T) {
	Convey("When an AP's metrics do have subscriptions", t, func() {
		m := metrics(make(map[string]*metric))
		ap := &availablePlugin{
			Metrics: &m,
		}
		for _, n := range [][]string{
			[]string{"test", "foo"},
			[]string{"test", "bar"},
			[]string{"test", "qux"},
		} {
			m := newMetric(&metricOpts{n})
			(*ap.Metrics)[m.Key()] = m
		}
		var wg sync.WaitGroup
		wg.Add(3)
		go func() {
			for i := 0; i < 3; i++ {
				(*ap.Metrics)["test.foo"].Subscribe()
			}
			defer wg.Done()
		}()
		go func() {
			for i := 0; i < 5; i++ {
				(*ap.Metrics)["test.bar"].Subscribe()
			}
			defer wg.Done()
		}()
		go func() {
			for i := 0; i < 2; i++ {
				(*ap.Metrics)["test.qux"].Subscribe()
			}
			defer wg.Done()
		}()
		wg.Wait()
		Convey("then it accurately represents the amount of subscriptions", func() {
			So(ap.Subscriptions(), ShouldEqual, 5)
		})
		Convey("when one or more of those metric's subscription counts change", func() {
			(*ap.Metrics)["test.bar"].Unsubscribe()
			Convey("It accurately represents the subscription count", func() {
				So(ap.Subscriptions(), ShouldEqual, 4)
			})
		})
		Convey("when a different metric becomes the max subscription holder", func() {
			(*ap.Metrics)["test.foo"].Subscribe()
			(*ap.Metrics)["test.foo"].Subscribe()
			(*ap.Metrics)["test.foo"].Subscribe()
			Convey("It accurately represents the subscription count", func() {
				So(ap.Subscriptions(), ShouldEqual, 6)
			})
		})
	})
	Convey("When an AP's metrics do not have subscriptions", t, func() {
		m := metrics(make(map[string]*metric))
		ap := &availablePlugin{
			Metrics: &m,
		}
		for _, n := range [][]string{
			[]string{"test", "foo"},
			[]string{"test", "bar"},
			[]string{"test", "qux"},
		} {
			m := newMetric(&metricOpts{n})
			(*ap.Metrics)[m.Key()] = m
		}
		Convey("then it reports 0 subscriptions", func() {
			So(ap.Subscriptions(), ShouldEqual, 0)
		})
	})
	Convey("When an AP has no metrics", t, func() {
		m := metrics(make(map[string]*metric))
		ap := &availablePlugin{
			Metrics: &m,
		}
		Convey("then it reports 0 subscriptions", func() {
			So(ap.Subscriptions(), ShouldEqual, 0)
		})
	})
}
