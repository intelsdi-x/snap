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
			_mt, err := mc.GetMetric(ns, -1)
			So(_mt, ShouldResemble, mt)
			So(err, ShouldBeNil)
		})
	})
	Convey("metricCatalog.GetMetric()", t, func() {
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
				_mt, err := mc.GetMetric(v, -1)
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
			m, err := mc.GetMetric(core.NewNamespace("foo", "bar"), -1)
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
			m, err := mc.GetMetric(core.NewNamespace("foo", "bar"), 2)
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
			_, err := mc.GetMetric(core.NewNamespace("foo", "bar"), 7)
			So(err.Error(), ShouldContainSubstring, "Metric not found:")
		})
	})
	Convey("metricCatalog.GetMetrics()", t, func() {
		mc := newMetricCatalog()
		ts := time.Now()
		Convey("add multiple metricTypes and get them back", func() {
			nss := []core.Namespace{
				core.NewNamespace("mock", "test", "1"),
				core.NewNamespace("mock", "test", "2"),
				core.NewNamespace("mock", "test", "3"),
				core.NewNamespace("mock", "cgroups").AddDynamicElement("cgroup_id", "cgroup id").AddStaticElement("out"),
				core.NewNamespace("mock", "cgroups").AddDynamicElement("cgroup_id", "cgroup id").AddStaticElement("in"),
			}

			lp := new(loadedPlugin)

			// add metricTypes to the catalog
			mt := []*metricType{}
			for i, ns := range nss {
				mt = append(mt, newMetricType(ns, ts, lp))
				mc.Add(mt[i])
			}
			Convey("validate adding metrics to the catalog", func() {
				for _, ns := range nss {
					// check if metric is in metric catalog
					_mt, err := mc.GetMetric(ns, -1)
					So(_mt, ShouldNotBeEmpty)
					So(err, ShouldBeNil)
				}
			})
			Convey("get a metricType explicitly (no wildcards)", func() {
				for k, ns := range nss {
					_mts, err := mc.GetMetrics(ns, -1)
					So(err, ShouldBeNil)
					So(len(_mts), ShouldEqual, 1)
					So(_mts[0], ShouldEqual, mt[k])
				}
			})
			Convey("get metricTypes with an asterisk at the end", func() {
				_mts, err := mc.GetMetrics(core.NewNamespace("mock", "test", "*"), -1)
				So(err, ShouldBeNil)
				// `/mock/test/*` should return 3 metrics:
				// `/mock/test/1`, `/mock/test/2`, and `/mock/test/3`
				So(len(_mts), ShouldEqual, 3)
			})
			Convey("get metricTypes with an asterisk in the middle of namespace", func() {
				_mts, err := mc.GetMetrics(core.NewNamespace("mock", "*", "2"), 0)
				So(err, ShouldBeNil)
				// `/mock/*/2 should return only 1 metric: `/mock/test/2`
				So(len(_mts), ShouldEqual, 1)
				So(_mts[0].Namespace(), ShouldResemble, core.NewNamespace("mock", "test", "2"))
			})
			Convey("get a metricType with specified instance of dynamic metric", func() {
				_mts, err := mc.GetMetrics(core.NewNamespace("mock", "cgroups", "A", "in"), -1)
				So(err, ShouldBeNil)
				So(len(_mts), ShouldEqual, 1)
				So(_mts[0].Namespace(), ShouldResemble, core.NewNamespace("mock", "cgroups").AddDynamicElement("cgroup_id", "cgroup id").AddStaticElement("in"))
			})
			Convey("get all metricTypes for specific instance of dynamic metric", func() {
				_mts, err := mc.GetMetrics(core.NewNamespace("mock", "cgroups", "A", "*"), -1)
				So(err, ShouldBeNil)
				// there are two dynamic metrics in the catalog: `/mock/cgroups/*/in` and /mock/cgroups/*/out`
				So(len(_mts), ShouldEqual, 2)

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
				_mts, err := mc.GetMetrics(core.NewNamespace("foo", "bar"), -1)
				So(err, ShouldBeNil)
				So(len(_mts), ShouldEqual, 1)
				So(_mts[0], ShouldEqual, m35)
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
				_mts, err := mc.GetMetrics(core.NewNamespace("foo", "bar"), 2)
				So(err, ShouldBeNil)
				So(len(_mts), ShouldEqual, 1)
				So(_mts[0], ShouldEqual, m2)
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
				_, err := mc.GetMetrics(core.NewNamespace("foo", "bar"), 7)
				So(err.Error(), ShouldContainSubstring, "Metric not found:")
			})
		})
	})
	Convey("metricCatalog.GetVersions()", t, func() {
		mc := newMetricCatalog()
		ts := time.Now()
		Convey("it returns all avaivailable versions", func() {
			lp2 := new(loadedPlugin)
			lp2.Meta.Version = 2
			lp35 := new(loadedPlugin)
			lp35.Meta.Version = 35
			m2 := newMetricType(core.NewNamespace("foo", "bar"), ts, lp2)
			mc.Add(m2)
			m35 := newMetricType(core.NewNamespace("foo", "bar"), ts, lp35)
			mc.Add(m35)
			_mts, err := mc.GetVersions(core.NewNamespace("foo", "bar"))
			So(err, ShouldBeNil)
			// all versions of /foo/bar should be returned
			So(len(_mts), ShouldEqual, 2)
		})
		Convey("it returns metric not found if metricType doesn't exist", func() {
			_mts, err := mc.GetVersions(core.NewNamespace("foo", "invalid"))
			So(_mts, ShouldBeEmpty)
			So(err.Error(), ShouldContainSubstring, "Metric not found:")
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
		Convey("remove a metricType from non-empty catalog", func() {
			// add metrics to the catalog to remove it later
			mt := []*metricType{}
			for i, ns := range nss {
				mt = append(mt, newMetricType(ns, ts, new(loadedPlugin)))
				mc.Add(mt[i])
			}
			Convey("validate adding metrics to the catalog", func() {
				for _, ns := range nss {
					// check if metric is in metric catalog
					_mt, err := mc.GetMetric(ns, -1)
					So(_mt, ShouldNotBeEmpty)
					So(err, ShouldBeNil)
				}
			})

			// remove a single metric from the catalog
			mc.Remove(nss[0])

			Convey("validate removing a single metric from the catalog", func() {
				_mt, err := mc.GetMetric(core.NewNamespace("mock", "test", "1"), -1)
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
					_mt, err := mc.GetMetric(ns, -1)
					So(_mt, ShouldBeNil)
					So(err, ShouldNotBeNil)
					So(err.Error(), ShouldContainSubstring, "Metric not found:")
				}
			})
		})
	})
	Convey("metricCatalog.Fetch()", t, func() {
		mc := newMetricCatalog()
		ts := time.Now()
		nss := []core.Namespace{
			core.NewNamespace("mock", "test", "1"),
			core.NewNamespace("mock", "test", "2"),
			core.NewNamespace("mock", "test", "3"),
			core.NewNamespace("mock", "cgroups", "*", "in"),
			core.NewNamespace("mock", "cgroups", "*", "out"),
		}
		Convey("retrive all metrics types under a given namespace", func() {

			Convey("add metrics types to the catalog", func() {
				mt := []*metricType{}
				for i, ns := range nss {
					mt = append(mt, newMetricType(ns, ts, new(loadedPlugin)))
					mc.Add(mt[i])
				}
				Convey("validate adding metrics to the catalog", func() {
					for _, ns := range nss {
						// check if metric is in metric catalog
						_mt, err := mc.GetMetric(ns, -1)
						So(_mt, ShouldNotBeEmpty)
						So(err, ShouldBeNil)
					}
				})
				Convey("fetch a single metric type from the catalog", func() {
					_mt, err := mc.Fetch(core.NewNamespace("mock", "test", "1"))
					So(err, ShouldBeNil)
					So(_mt, ShouldNotBeNil)
					So(len(_mt), ShouldEqual, 1)
					So(_mt[0].Namespace(), ShouldResemble, core.NewNamespace("mock", "test", "1"))

				})
				Convey("fetch a group of metrics types from the catalog", func() {
					_mts, err := mc.Fetch(core.NewNamespace("mock", "cgroups"))
					So(err, ShouldBeNil)
					So(_mts, ShouldNotBeNil)
					// two metrics types are under /mock/cgroups/
					So(len(_mts), ShouldEqual, 2)
				})
				Convey("fetch all metrics types from the catalog", func() {
					_mts, err := mc.Fetch(core.NewNamespace())
					So(err, ShouldBeNil)
					So(_mts, ShouldNotBeNil)
					// all the cataloged metrics types should be returned
					So(len(_mts), ShouldEqual, len(nss))
				})
				Convey("try to fetch a metric type from the catalog which not exist", func() {
					_mts, err := mc.Fetch(core.NewNamespace("mock", "invalid", "name"))
					So(err, ShouldNotBeNil)
					So(_mts, ShouldBeEmpty)
					So(err.Error(), ShouldContainSubstring, "Metrics not found below a given namespace: /mock/invalid/name")
				})
				Convey("try to fetch all metrics types from empty catalog", func() {
					// remove all metrics types from the catalog
					for _, ns := range nss {
						mc.Remove(ns)
					}
					_mts, err := mc.Fetch(core.NewNamespace())
					So(err, ShouldNotBeNil)
					So(_mts, ShouldBeEmpty)
					So(err.Error(), ShouldContainSubstring, "Metric catalog is empty (no plugin loaded)")
				})
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
			m, err2 := mc.GetMetric(core.NewNamespace("test1"), -1)
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
			m, err2 := mc.GetMetric(core.NewNamespace("test1"), -1)
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
