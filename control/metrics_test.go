// +build legacy

/*
http://www.apache.org/licenses/LICENSE-2.0.txt


Copyright 2015-2016 Intel Corporation

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package control

import (
	"testing"
	"time"

	"github.com/intelsdi-x/snap/control/plugin"
	"github.com/intelsdi-x/snap/core"

	. "github.com/smartystreets/goconvey/convey"
)

func TestMetricType(t *testing.T) {
	Convey("newMetricType()", t, func() {
		Convey("returns a metricType", func() {
			mt := newMetricType(core.NewNamespace("test"), time.Now(), new(loadedPlugin))
			So(mt, ShouldHaveSameTypeAs, new(metricType))
		})
	})
	Convey("metricType.Namespace()", t, func() {
		Convey("returns the namespace of a metricType", func() {
			ns := core.NewNamespace("test")
			mt := newMetricType(ns, time.Now(), new(loadedPlugin))
			So(mt.Namespace(), ShouldHaveSameTypeAs, ns)
			So(mt.Namespace(), ShouldResemble, ns)
		})
	})
	Convey("metricType.Version()", t, func() {
		Convey("returns the namespace of a metricType", func() {
			ns := core.NewNamespace("test")
			lp := &loadedPlugin{Meta: plugin.PluginMeta{Version: 1}}
			mt := newMetricType(ns, time.Now(), lp)
			So(mt.Version(), ShouldEqual, 1)
		})
	})
	Convey("metricType.LastAdvertisedTimestamp()", t, func() {
		Convey("returns the LastAdvertisedTimestamp for the metricType", func() {
			ts := time.Now()
			mt := newMetricType(core.NewNamespace("test"), ts, new(loadedPlugin))
			So(mt.LastAdvertisedTime(), ShouldHaveSameTypeAs, ts)
			So(mt.LastAdvertisedTime(), ShouldResemble, ts)
		})
	})
	Convey("metricType.Key()", t, func() {
		Convey("returns the key for the metricType", func() {
			ts := time.Now()
			lp := new(loadedPlugin)
			mt := newMetricType(core.NewNamespace("foo", "bar"), ts, lp)
			key := mt.Key()
			So(key, ShouldEqual, "/foo/bar/0")
		})
		Convey("returns the key for the queried version", func() {
			ts := time.Now()
			lp2 := new(loadedPlugin)
			lp2.Meta.Version = 2
			mt := newMetricType(core.NewNamespace("foo", "bar"), ts, lp2)
			key := mt.Key()
			So(key, ShouldEqual, "/foo/bar/2")
		})
	})
}

func TestMetricMatching(t *testing.T) {
	Convey("metricCatalog.GetQueriedNamespaces()", t, func() {
		Convey("verify query support for static metrics", func() {
			mc := newMetricCatalog()
			ns := []core.Namespace{
				core.NewNamespace("mock", "foo", "bar"),
				core.NewNamespace("mock", "foo", "baz"),
				core.NewNamespace("mock", "asdf", "bar"),
				core.NewNamespace("mock", "asdf", "baz"),
				core.NewNamespace("mock", "test", "1"),
				core.NewNamespace("mock", "test", "2"),
				core.NewNamespace("mock", "test", "3"),
				core.NewNamespace("mock", "test", "4"),
			}
			lp := new(loadedPlugin)
			ts := time.Now()
			mt := []*metricType{
				newMetricType(ns[0], ts, lp),
				newMetricType(ns[1], ts, lp),
				newMetricType(ns[2], ts, lp),
				newMetricType(ns[3], ts, lp),
				newMetricType(ns[4], ts, lp),
				newMetricType(ns[5], ts, lp),
				newMetricType(ns[6], ts, lp),
				newMetricType(ns[7], ts, lp),
			}

			for _, v := range mt {
				mc.Add(v)
			}

			Convey("match /mock/foo/*", func() {
				mc.UpdateQueriedNamespaces(core.NewNamespace("mock", "foo", "*"))
				nss, err := mc.GetQueriedNamespaces(core.NewNamespace("mock", "foo", "*"))
				So(err, ShouldBeNil)
				So(len(nss), ShouldEqual, 2)
				So(nss, ShouldResemble, []core.Namespace{
					core.NewNamespace("mock", "foo", "bar"),
					core.NewNamespace("mock", "foo", "baz"),
				})

			})
			Convey("match /mock/test/*", func() {
				mc.UpdateQueriedNamespaces(core.NewNamespace("mock", "test", "*"))
				nss, err := mc.GetQueriedNamespaces(core.NewNamespace("mock", "test", "*"))
				So(err, ShouldBeNil)
				So(len(nss), ShouldEqual, 4)
				So(nss, ShouldResemble, []core.Namespace{
					core.NewNamespace("mock", "test", "1"),
					core.NewNamespace("mock", "test", "2"),
					core.NewNamespace("mock", "test", "3"),
					core.NewNamespace("mock", "test", "4"),
				})
			})
			Convey("match /mock/*/bar", func() {
				mc.UpdateQueriedNamespaces(core.NewNamespace("mock", "*", "bar"))
				nss, err := mc.GetQueriedNamespaces(core.NewNamespace("mock", "*", "bar"))
				So(err, ShouldBeNil)
				So(len(nss), ShouldEqual, 2)
				So(nss, ShouldResemble, []core.Namespace{
					core.NewNamespace("mock", "foo", "bar"),
					core.NewNamespace("mock", "asdf", "bar"),
				})
			})
			Convey("match /mock/*", func() {
				mc.UpdateQueriedNamespaces(core.NewNamespace("mock", "*"))
				nss, err := mc.GetQueriedNamespaces(core.NewNamespace("mock", "*"))
				So(err, ShouldBeNil)
				So(len(nss), ShouldEqual, len(ns))
				So(nss, ShouldResemble, ns)
			})
			Convey("match /mock/(foo|asdf)/baz", func() {
				mc.UpdateQueriedNamespaces(core.NewNamespace("mock", "(foo|asdf)", "baz"))
				nss, err := mc.GetQueriedNamespaces(core.NewNamespace("mock", "(foo|asdf)", "baz"))
				So(err, ShouldBeNil)
				So(len(nss), ShouldEqual, 2)
				So(nss, ShouldResemble, []core.Namespace{
					core.NewNamespace("mock", "foo", "baz"),
					core.NewNamespace("mock", "asdf", "baz"),
				})
			})
			Convey("match /mock/test/(1|2|3)", func() {
				mc.UpdateQueriedNamespaces(core.NewNamespace("mock", "test", "(1|2|3)"))
				nss, err := mc.GetQueriedNamespaces(core.NewNamespace("mock", "test", "(1|2|3)"))
				So(err, ShouldBeNil)
				So(len(nss), ShouldEqual, 3)
				So(nss, ShouldResemble, []core.Namespace{
					core.NewNamespace("mock", "test", "1"),
					core.NewNamespace("mock", "test", "2"),
					core.NewNamespace("mock", "test", "3"),
				})
			})
			Convey("invalid matching", func() {
				mc.UpdateQueriedNamespaces(core.NewNamespace("mock", "not", "exist", "metric"))
				nss, err := mc.GetQueriedNamespaces(core.NewNamespace("mock", "not", "exist", "metric"))
				So(err, ShouldNotBeNil)
				So(nss, ShouldBeEmpty)
				So(err.Error(), ShouldContainSubstring, "Metric not found:")
			})
		})
		Convey("verify query support for dynamic metrics", func() {
			mc := newMetricCatalog()
			ns := []core.Namespace{
				core.NewNamespace("mock", "cgroups", "*", "bar"),
				core.NewNamespace("mock", "cgroups", "*", "baz"),
				core.NewNamespace("mock", "cgroups", "*", "in"),
				core.NewNamespace("mock", "cgroups", "*", "out"),
				core.NewNamespace("mock", "cgroups", "*", "test", "1"),
				core.NewNamespace("mock", "cgroups", "*", "test", "2"),
				core.NewNamespace("mock", "cgroups", "*", "test", "3"),
				core.NewNamespace("mock", "cgroups", "*", "test", "4"),
			}
			lp := new(loadedPlugin)
			ts := time.Now()
			mt := []*metricType{
				newMetricType(ns[0], ts, lp),
				newMetricType(ns[1], ts, lp),
				newMetricType(ns[2], ts, lp),
				newMetricType(ns[3], ts, lp),
				newMetricType(ns[4], ts, lp),
				newMetricType(ns[5], ts, lp),
				newMetricType(ns[6], ts, lp),
				newMetricType(ns[7], ts, lp),
			}

			for _, v := range mt {
				mc.Add(v)
			}
			// check if metrics were added to metric catalog
			So(len(mc.Keys()), ShouldEqual, len(ns))

			Convey("match /mock/cgroups/*", func() {
				mc.UpdateQueriedNamespaces(core.NewNamespace("mock", "cgroups", "*"))
				nss, err := mc.GetQueriedNamespaces(core.NewNamespace("mock", "cgroups", "*"))
				So(err, ShouldBeNil)
				So(len(nss), ShouldEqual, len(ns))
				So(nss, ShouldResemble, ns)
			})
			Convey("match /mock/cgroups/*/bar", func() {
				mc.UpdateQueriedNamespaces(core.NewNamespace("mock", "cgroups", "*", "bar"))
				nss, err := mc.GetQueriedNamespaces(core.NewNamespace("mock", "cgroups", "*", "bar"))
				So(err, ShouldBeNil)
				So(len(nss), ShouldEqual, 1)
				So(nss, ShouldResemble, []core.Namespace{
					core.NewNamespace("mock", "cgroups", "*", "bar"),
				})
			})
			Convey("match /mock/cgroups/*/(bar|baz)", func() {
				mc.UpdateQueriedNamespaces(core.NewNamespace("mock", "cgroups", "*", "(bar|baz)"))
				nss, err := mc.GetQueriedNamespaces(core.NewNamespace("mock", "cgroups", "*", "(bar|baz)"))
				So(err, ShouldBeNil)
				So(len(nss), ShouldEqual, 2)
				So(nss, ShouldResemble, []core.Namespace{
					core.NewNamespace("mock", "cgroups", "*", "bar"),
					core.NewNamespace("mock", "cgroups", "*", "baz"),
				})
			})
			Convey("match /mock/cgroups/*/test/*", func() {
				mc.UpdateQueriedNamespaces(core.NewNamespace("mock", "cgroups", "*", "test", "*"))
				nss, err := mc.GetQueriedNamespaces(core.NewNamespace("mock", "cgroups", "*", "test", "*"))
				So(err, ShouldBeNil)
				So(len(nss), ShouldEqual, 4)
				So(nss, ShouldResemble, []core.Namespace{
					core.NewNamespace("mock", "cgroups", "*", "test", "1"),
					core.NewNamespace("mock", "cgroups", "*", "test", "2"),
					core.NewNamespace("mock", "cgroups", "*", "test", "3"),
					core.NewNamespace("mock", "cgroups", "*", "test", "4"),
				})
			})
			Convey("match /mock/cgroups/*/test/(1|2|3)", func() {
				mc.UpdateQueriedNamespaces(core.NewNamespace("mock", "cgroups", "*", "test", "(1|2|3)"))
				nss, err := mc.GetQueriedNamespaces(core.NewNamespace("mock", "cgroups", "*", "test", "(1|2|3)"))
				So(err, ShouldBeNil)
				So(len(nss), ShouldEqual, 3)
				So(nss, ShouldResemble, []core.Namespace{
					core.NewNamespace("mock", "cgroups", "*", "test", "1"),
					core.NewNamespace("mock", "cgroups", "*", "test", "2"),
					core.NewNamespace("mock", "cgroups", "*", "test", "3"),
				})
			})
		})

	})
}

func TestMetricCatalog(t *testing.T) {
	Convey("newMetricCatalog()", t, func() {
		Convey("returns a metricCatalog", func() {
			mc := newMetricCatalog()
			So(mc, ShouldHaveSameTypeAs, new(metricCatalog))
		})
	})
	Convey("metricCatalog.Add()", t, func() {
		Convey("adds a metricType to the metricCatalog", func() {
			ns := core.NewNamespace("test")
			mt := newMetricType(ns, time.Now(), new(loadedPlugin))
			mt.description = "some description"
			mc := newMetricCatalog()
			mc.Add(mt)
			_mt, err := mc.Get(ns, -1)
			So(_mt, ShouldResemble, mt)
			So(err, ShouldBeNil)
		})
	})
	Convey("metricCatalog.Get()", t, func() {
		mc := newMetricCatalog()
		ts := time.Now()
		Convey("add multiple metricTypes and get them back", func() {
			ns := []core.Namespace{
				core.NewNamespace("test1"),
				core.NewNamespace("test2"),
				core.NewNamespace("test3"),
			}
			lp := new(loadedPlugin)
			mt := []*metricType{
				newMetricType(ns[0], ts, lp),
				newMetricType(ns[1], ts, lp),
				newMetricType(ns[2], ts, lp),
			}
			for _, v := range mt {
				mc.Add(v)
			}
			for k, v := range ns {
				_mt, err := mc.Get(v, -1)
				So(_mt, ShouldEqual, mt[k])
				So(err, ShouldBeNil)
			}
		})
		Convey("it returns the latest version", func() {
			lp2 := new(loadedPlugin)
			lp2.Meta.Version = 2
			lp35 := new(loadedPlugin)
			lp35.Meta.Version = 35
			m2 := newMetricType(core.NewNamespace("foo", "bar"), ts, lp2)
			mc.Add(m2)
			m35 := newMetricType(core.NewNamespace("foo", "bar"), ts, lp35)
			mc.Add(m35)
			m, err := mc.Get(core.NewNamespace("foo", "bar"), -1)
			So(err, ShouldBeNil)
			So(m, ShouldEqual, m35)
		})
		Convey("it returns the queried version", func() {
			lp2 := new(loadedPlugin)
			lp2.Meta.Version = 2
			lp35 := new(loadedPlugin)
			lp35.Meta.Version = 35
			m2 := newMetricType(core.NewNamespace("foo", "bar"), ts, lp2)
			mc.Add(m2)
			m35 := newMetricType(core.NewNamespace("foo", "bar"), ts, lp35)
			mc.Add(m35)
			m, err := mc.Get(core.NewNamespace("foo", "bar"), 2)
			So(err, ShouldBeNil)
			So(m, ShouldEqual, m2)
		})
		Convey("it returns metric not found if version doesn't exist", func() {
			lp2 := new(loadedPlugin)
			lp2.Meta.Version = 2
			lp35 := new(loadedPlugin)
			lp35.Meta.Version = 35
			m2 := newMetricType(core.NewNamespace("foo", "bar"), ts, lp2)
			mc.Add(m2)
			m35 := newMetricType(core.NewNamespace("foo", "bar"), ts, lp35)
			mc.Add(m35)
			_, err := mc.Get(core.NewNamespace("foo", "bar"), 7)
			So(err.Error(), ShouldContainSubstring, "Metric not found:")
		})
	})
	Convey("metricCatalog.Table()", t, func() {
		Convey("returns a copy of the table", func() {
			mc := newMetricCatalog()
			mt := newMetricType(core.NewNamespace("foo", "bar"), time.Now(), &loadedPlugin{})
			mc.Add(mt)
			//TODO test tree
			//So(mc.Table(), ShouldHaveSameTypeAs, map[string][]*metricType{})
			//So(mc.Table()["foo.bar"], ShouldResemble, []*metricType{mt})
		})
	})
	Convey("metricCatalog.Next()", t, func() {
		ns := core.NewNamespace("test")
		mt := newMetricType(ns, time.Now(), new(loadedPlugin))
		mc := newMetricCatalog()
		Convey("returns false on empty table", func() {
			ok := mc.Next()
			So(ok, ShouldEqual, false)
		})
		Convey("returns true on populated table", func() {
			mc.Add(mt)
			ok := mc.Next()
			So(ok, ShouldEqual, true)
		})
	})
	Convey("metricCatalog.Item()", t, func() {
		ns := []core.Namespace{
			core.NewNamespace("test1"),
			core.NewNamespace("test2"),
			core.NewNamespace("test3"),
		}
		lp := new(loadedPlugin)
		t := time.Now()
		mt := []*metricType{
			newMetricType(ns[0], t, lp),
			newMetricType(ns[1], t, lp),
			newMetricType(ns[2], t, lp),
		}
		mc := newMetricCatalog()
		for _, v := range mt {
			mc.Add(v)
		}
		Convey("return first key and item in table", func() {
			mc.Next()
			key, item := mc.Item()
			So(key, ShouldEqual, ns[0].Key())
			So(item, ShouldResemble, []*metricType{mt[0]})
		})
		Convey("return second key and item in table", func() {
			mc.Next()
			mc.Next()
			key, item := mc.Item()
			So(key, ShouldEqual, ns[1].Key())
			So(item, ShouldResemble, []*metricType{mt[1]})
		})
		Convey("return third key and item in table", func() {
			mc.Next()
			mc.Next()
			mc.Next()
			key, item := mc.Item()
			So(key, ShouldEqual, ns[2].Key())
			So(item, ShouldResemble, []*metricType{mt[2]})
		})
	})

	Convey("metricCatalog.Remove()", t, func() {
		mc := newMetricCatalog()
		ts := time.Now()
		nss := []core.Namespace{
			core.NewNamespace("mock", "test", "1"),
			core.NewNamespace("mock", "test", "2"),
			core.NewNamespace("mock", "test", "3"),
			core.NewNamespace("mock", "cgroups", "*", "in"),
			core.NewNamespace("mock", "cgroups", "*", "out"),
		}
		Convey("removes a metricType from the catalog", func() {
			// adding metrics to the catalog
			mt := []*metricType{}
			for i, ns := range nss {
				mt = append(mt, newMetricType(ns, ts, new(loadedPlugin)))
				mc.Add(mt[i])
			}
			Convey("validate adding metrics to the catalog", func() {
				for _, ns := range nss {
					// check if metric is in metric catalog
					_mt, err := mc.Get(ns, -1)
					So(_mt, ShouldNotBeEmpty)
					So(err, ShouldBeNil)
				}
			})

			// remove a single metric from the catalog
			mc.Remove(nss[0])

			Convey("validate removing a single metric from the catalog", func() {
				_mt, err := mc.Get(core.NewNamespace("mock", "test", "1"), -1)
				So(_mt, ShouldBeNil)
				So(err, ShouldNotBeNil)

			})

			// remove the rest metrics from the catalog
			for _, ns := range nss[1:] {
				mc.Remove(ns)
			}

			Convey("validate removing all metrics from the catalog", func() {
				for _, ns := range nss {
					// check if metric is in metric catalog
					_mt, err := mc.Get(ns, -1)
					So(_mt, ShouldBeNil)
					So(err, ShouldNotBeNil)
					So(err.Error(), ShouldContainSubstring, "Metric not found:")
				}
			})
		})
	})
}

func TestSubscribe(t *testing.T) {
	ns := []core.Namespace{
		core.NewNamespace("test1"),
		core.NewNamespace("test2"),
		core.NewNamespace("test3"),
	}
	lp := new(loadedPlugin)
	ts := time.Now()
	mt := []*metricType{
		newMetricType(ns[0], ts, lp),
		newMetricType(ns[1], ts, lp),
		newMetricType(ns[2], ts, lp),
	}
	mc := newMetricCatalog()
	for _, v := range mt {
		mc.Add(v)
	}
	Convey("when the metric is not in the table", t, func() {
		Convey("then it returns an error", func() {
			err := mc.Subscribe([]string{"test4"}, -1)
			So(err.Error(), ShouldContainSubstring, "Metric not found:")
		})
	})
	Convey("when the metric is in the table", t, func() {
		Convey("then it gets correctly increments the count", func() {
			err := mc.Subscribe([]string{"test1"}, -1)
			So(err, ShouldBeNil)
			m, err2 := mc.Get(core.NewNamespace("test1"), -1)
			So(err2, ShouldBeNil)
			So(m.subscriptions, ShouldEqual, 1)
		})
	})
}

func TestUnsubscribe(t *testing.T) {
	ns := []core.Namespace{
		core.NewNamespace("test1"),
		core.NewNamespace("test2"),
		core.NewNamespace("test3"),
	}
	lp := new(loadedPlugin)
	ts := time.Now()
	mt := []*metricType{
		newMetricType(ns[0], ts, lp),
		newMetricType(ns[1], ts, lp),
		newMetricType(ns[2], ts, lp),
	}
	mc := newMetricCatalog()
	for _, v := range mt {
		mc.Add(v)
	}
	Convey("when the metric is in the table", t, func() {
		Convey("then its subscription count is decremented", func() {
			err := mc.Subscribe([]string{"test1"}, -1)
			So(err, ShouldBeNil)
			err1 := mc.Unsubscribe([]string{"test1"}, -1)
			So(err1, ShouldBeNil)
			m, err2 := mc.Get(core.NewNamespace("test1"), -1)
			So(err2, ShouldBeNil)
			So(m.subscriptions, ShouldEqual, 0)
		})
	})
	Convey("when the metric is not in the table", t, func() {
		Convey("then it returns metric not found error", func() {
			err := mc.Unsubscribe([]string{"test4"}, -1)
			So(err.Error(), ShouldContainSubstring, "Metric not found:")
		})
	})
	Convey("when the metric's count is already 0", t, func() {
		Convey("then it returns negative subscription count error", func() {
			err := mc.Unsubscribe([]string{"test1"}, -1)
			So(err, ShouldResemble, errNegativeSubCount)
		})
	})
}

func TestSubscriptionCount(t *testing.T) {
	m := newMetricType(core.NewNamespace("test"), time.Now(), &loadedPlugin{})
	Convey("it returns the subscription count", t, func() {
		m.Subscribe()
		So(m.SubscriptionCount(), ShouldEqual, 1)
		m.Subscribe()
		m.Subscribe()
		So(m.SubscriptionCount(), ShouldEqual, 3)
		m.Unsubscribe()
		So(m.SubscriptionCount(), ShouldEqual, 2)
	})
}

func TestMetricNamespaceValidation(t *testing.T) {
	Convey("validateMetricNamespace()", t, func() {
		Convey("validation passes", func() {
			ns := core.NewNamespace("mock", "foo", "bar")
			err := validateMetricNamespace(ns)
			So(err, ShouldBeNil)
		})
		Convey("contains not allowed characters", func() {
			ns := core.NewNamespace("mock", "foo", "(bar)")
			err := validateMetricNamespace(ns)
			So(err, ShouldNotBeNil)
		})
		Convey("contains unacceptable wildcardat at the end", func() {
			ns := core.NewNamespace("mock", "foo", "*")
			err := validateMetricNamespace(ns)
			So(err, ShouldNotBeNil)
		})
	})
}

func TestMetricStaticDynamicNamespace(t *testing.T) {
	Convey("validateStaticDynamic()", t, func() {
		Convey("has static elements only", func() {
			ns := core.NewNamespace("mock", "foo", "bar")
			err := validateMetricNamespace(ns)
			So(err, ShouldBeNil)
		})
		Convey("had both static and dynamic elements", func() {
			ns := core.NewNamespace("mock", "foo", "*", "bar")
			ns[2].Name = "dynamic element"
			err := validateMetricNamespace(ns)
			So(err, ShouldBeNil)
		})
		Convey("has name for a static element", func() {
			ns := core.NewNamespace("mock", "foo")
			ns[0].Name = "static element"
			err := validateMetricNamespace(ns)
			So(err, ShouldNotBeNil)
		})
		Convey("has * but no name", func() {
			ns := core.NewNamespace("mock", "foo", "*", "bar")
			err := validateMetricNamespace(ns)
			So(err, ShouldNotBeNil)
		})
	})
}
