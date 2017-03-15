// +build medium
/*
http://www.apache.org/licenses/LICENSE-2.0.txt


Copyright 2015 Intel Corporation

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

	"github.com/intelsdi-x/snap/core"
	. "github.com/smartystreets/goconvey/convey"
)

func TestTrie(t *testing.T) {
	Convey("Make a new trie", t, func() {
		trie := NewMTTrie()
		So(trie, ShouldNotBeNil)
		So(trie, ShouldHaveSameTypeAs, &MTTrie{})
	})
	Convey("Fetch", t, func() {
		trie := NewMTTrie()
		Convey("Add and collect split namespace", func() {
			mt := newMetricType(core.NewNamespace("intel", "foo"), time.Now(), new(loadedPlugin))
			mt2 := newMetricType(core.NewNamespace("intel", "baz", "qux"), time.Now(), new(loadedPlugin))
			trie.Add(mt)
			trie.Add(mt2)

			in, err := trie.Fetch([]string{"intel"})
			So(err, ShouldBeNil)
			So(len(in), ShouldEqual, 2)
			for _, mt := range in {
				So(mt, ShouldNotBeNil)
				So(mt.Namespace(), ShouldHaveSameTypeAs, core.NewNamespace(""))
			}
		})
		Convey("Add and collect with nodes with children", func() {
			mt := newMetricType(core.NewNamespace("intel", "foo", "bar"), time.Now(), new(loadedPlugin))
			mt2 := newMetricType(core.NewNamespace("intel", "foo"), time.Now(), new(loadedPlugin))
			mt3 := newMetricType(core.NewNamespace("intel", "foo", "qux"), time.Now(), new(loadedPlugin))
			trie.Add(mt)
			trie.Add(mt2)
			trie.Add(mt3)

			in, err := trie.Fetch([]string{"intel"})
			So(err, ShouldBeNil)
			So(len(in), ShouldEqual, 3)
		})
		Convey("Add and collect at node with mt and children", func() {
			mt := newMetricType(core.NewNamespace("intel", "foo", "bar"), time.Now(), new(loadedPlugin))
			mt2 := newMetricType(core.NewNamespace("intel", "foo"), time.Now(), new(loadedPlugin))
			trie.Add(mt)
			trie.Add(mt2)

			in, err := trie.Fetch([]string{"intel", "foo"})
			So(err, ShouldBeNil)
			So(len(in), ShouldEqual, 2)
		})
		Convey("add and collect single depth namespace", func() {
			mt := newMetricType(core.NewNamespace("test"), time.Now(), new(loadedPlugin))
			trie.Add(mt)
			t, err := trie.Fetch([]string{"test"})
			So(err, ShouldBeNil)
			So(t[0].Namespace(), ShouldResemble, core.NewNamespace("test"))
		})
		Convey("add and longer length with single child", func() {
			mt := newMetricType(core.NewNamespace("d", "a", "n", "b", "a", "r"), time.Now(), new(loadedPlugin))
			trie.Add(mt)
			d, err := trie.Fetch([]string{"d", "a", "n", "b", "a", "r"})
			So(err, ShouldBeNil)
			So(d[0].Namespace(), ShouldResemble, core.NewNamespace("d", "a", "n", "b", "a", "r"))
			dd, err := trie.Fetch([]string{"d", "a", "n"})
			So(err, ShouldBeNil)
			So(dd[0].Namespace(), ShouldResemble, core.NewNamespace("d", "a", "n", "b", "a", "r"))
		})
		Convey("Multiple versions", func() {
			lp := new(loadedPlugin)
			lp.Meta.Version = 1
			mt := newMetricType(core.NewNamespace("intel", "foo"), time.Now(), lp)
			lp2 := new(loadedPlugin)
			lp2.Meta.Version = 2
			mt2 := newMetricType(core.NewNamespace("intel", "foo"), time.Now(), lp2)
			trie.Add(mt)
			trie.Add(mt2)
			n, err := trie.Fetch([]string{"intel"})
			So(err, ShouldBeNil)
			So(len(n), ShouldEqual, 2)
		})
		Convey("Fetch with error: not found", func() {
			_, err := trie.Fetch([]string{"not", "present"})
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, "No metric found below the given namespace: /not/present")
		})
		Convey("Fetch with error: depth exceeded", func() {
			lp := new(loadedPlugin)
			lp.Meta.Version = 1
			mt := newMetricType(core.NewNamespace("intel", "foo"), time.Now(), lp)
			trie.Add(mt)
			_, err := trie.Fetch([]string{"intel", "foo", "bar", "baz"})
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, "No metric found below the given namespace: /intel/foo/bar/baz")

		})
	})
}

func TestTrie_GetMetrics(t *testing.T) {
	Convey("Simply get metrics", t, func() {
		Convey("adding nodes to mttrie", func() {
			trie := NewMTTrie()
			lp2 := new(loadedPlugin)
			lp2.Meta.Version = 2

			lp5 := new(loadedPlugin)
			lp5.Meta.Version = 5

			mtstatic2 := newMetricType(core.NewNamespace("intel", "mock", "foo"), time.Now(), lp2)
			mtstatic5 := newMetricType(core.NewNamespace("intel", "mock", "foo"), time.Now(), lp5)

			mtdynamic2 := newMetricType(core.NewNamespace("intel", "mock").AddDynamicElement("host", "host id").AddStaticElement("baz"), time.Now(), lp2)
			mtdynamic5 := newMetricType(core.NewNamespace("intel", "mock").AddDynamicElement("host", "host id").AddStaticElement("baz"), time.Now(), lp5)

			trie.Add(mtstatic2)
			trie.Add(mtstatic5)
			trie.Add(mtdynamic2)
			trie.Add(mtdynamic5)

			Convey("for requested static metric", func() {
				Convey("get the latest version", func() {
					mts, err := trie.GetMetrics([]string{"intel", "mock", "foo"}, -1)
					So(err, ShouldBeNil)
					So(len(mts), ShouldEqual, 1)
					So(mts[0], ShouldEqual, mtstatic5)
				})
				Convey("get the queried version", func() {
					mts, err := trie.GetMetrics([]string{"intel", "mock", "foo"}, 2)
					So(err, ShouldBeNil)
					So(len(mts), ShouldEqual, 1)
					So(mts[0], ShouldEqual, mtstatic2)
				})
				Convey("error: the queried version of metric cannot be found", func() {
					mts, err := trie.GetMetrics([]string{"intel", "mock", "foo"}, 6)
					So(err, ShouldNotBeNil)
					So(mts, ShouldBeEmpty)
					So(err.Error(), ShouldContainSubstring, "Metric not found: /intel/mock/foo (version: 6)")
				})
				Convey("error: the queried metric cannot be found", func() {
					mts, err := trie.GetMetrics([]string{"intel", "mock", "invalid"}, -1)
					So(err, ShouldNotBeNil)
					So(mts, ShouldBeEmpty)
					So(err.Error(), ShouldContainSubstring, "Metric not found: /intel/mock/invalid (version: -1)")
				})
			})
			Convey("for requested dynamic metric", func() {
				Convey("get the latest version", func() {
					mts, err := trie.GetMetrics([]string{"intel", "mock", "*", "baz"}, -1)
					So(err, ShouldBeNil)
					So(len(mts), ShouldEqual, 1)
					So(mts[0], ShouldEqual, mtdynamic5)
				})
				Convey("get the queried version", func() {
					mts, err := trie.GetMetrics([]string{"intel", "mock", "*", "baz"}, 2)
					So(err, ShouldBeNil)
					So(len(mts), ShouldEqual, 1)
					So(mts[0], ShouldEqual, mtdynamic2)
				})
				Convey("error: the queried version of metric cannot be found", func() {
					mts, err := trie.GetMetrics([]string{"intel", "mock", "*", "baz"}, 6)
					So(err, ShouldNotBeNil)
					So(mts, ShouldBeEmpty)
					So(err.Error(), ShouldContainSubstring, "Metric not found: /intel/mock/*/baz (version: 6)")
				})
				Convey("error: the queried metric cannot be found", func() {
					mts, err := trie.GetMetrics([]string{"intel", "mock", "*", "invalid"}, -1)
					So(err, ShouldNotBeNil)
					So(mts, ShouldBeEmpty)
					So(err.Error(), ShouldContainSubstring, "Metric not found: /intel/mock/*/invalid (version: -1)")
				})
			})
		})
	})
	Convey("Query in get metrics", t, func() {
		Convey("adding nodes to mttrie", func() {
			trie := NewMTTrie()
			lpMock2 := new(loadedPlugin)
			lpMock2.Meta.Version = 2

			lpMock5 := new(loadedPlugin)
			lpMock5.Meta.Version = 5

			lpAnothermock2 := new(loadedPlugin)
			lpAnothermock2.Meta.Version = 2

			lpAnothermock5 := new(loadedPlugin)
			lpAnothermock5.Meta.Version = 5

			// notice, that there are two version of each of mock metrics
			mockMetrics := []*metricType{
				newMetricType(core.NewNamespace("intel", "mock", "foo"), time.Now(), lpMock2),
				newMetricType(core.NewNamespace("intel", "mock", "foo"), time.Now(), lpMock5),

				newMetricType(core.NewNamespace("intel", "mock", "bar"), time.Now(), lpMock2),
				newMetricType(core.NewNamespace("intel", "mock", "bar"), time.Now(), lpMock5),

				newMetricType(core.NewNamespace("intel", "mock").AddDynamicElement("host", "host id").AddStaticElement("baz"), time.Now(), lpMock2),
				newMetricType(core.NewNamespace("intel", "mock").AddDynamicElement("host", "host id").AddStaticElement("baz"), time.Now(), lpMock5),
			}
			//  notice, that there are two version of each of anothermock metrics
			anothermockMetrics := []*metricType{
				newMetricType(core.NewNamespace("intel", "anothermock", "foo"), time.Now(), lpAnothermock2),
				newMetricType(core.NewNamespace("intel", "anothermock", "foo"), time.Now(), lpAnothermock5),

				newMetricType(core.NewNamespace("intel", "anothermock", "bar"), time.Now(), lpAnothermock2),
				newMetricType(core.NewNamespace("intel", "anothermock", "bar"), time.Now(), lpAnothermock5),

				newMetricType(core.NewNamespace("intel", "anothermock").AddDynamicElement("host", "host id").AddStaticElement("baz"), time.Now(), lpAnothermock2),
				newMetricType(core.NewNamespace("intel", "anothermock").AddDynamicElement("host", "host id").AddStaticElement("baz"), time.Now(), lpAnothermock5),
			}

			// adding mockMetrics and anothermockMetrics to mtrie
			metrics := append(mockMetrics, anothermockMetrics...)
			for _, m := range metrics {
				trie.Add(m)
			}

			// numbers of metrics (they are represented in two version)
			numOfMockMetrics := len(mockMetrics) / 2
			numOfAnothermockMetrics := len(anothermockMetrics) / 2
			numOfAllMetrics := numOfMockMetrics + numOfAnothermockMetrics

			Convey("when the requested namespace is /intel/mock/*", func() {
				Convey("get the latest version", func() {
					mts, err := trie.GetMetrics([]string{"intel", "mock", "*"}, -1)
					So(err, ShouldBeNil)
					So(len(mts), ShouldEqual, numOfMockMetrics)
					Convey("version should be the latest", func() {
						for _, mt := range mts {
							So(mt.Version(), ShouldEqual, 5)
						}
					})
				})
				Convey("get the queried version", func() {
					mts, err := trie.GetMetrics([]string{"intel", "mock", "*"}, 2)
					So(err, ShouldBeNil)
					So(len(mts), ShouldEqual, numOfMockMetrics)
					Convey("version should be the queried version", func() {
						for _, mt := range mts {
							So(mt.Version(), ShouldEqual, 2)
						}
					})
				})
			})
			Convey("when the requested namespace is /intel/anothermock/*", func() {
				Convey("get the latest version", func() {
					mts, err := trie.GetMetrics([]string{"intel", "anothermock", "*"}, -1)
					So(err, ShouldBeNil)
					So(len(mts), ShouldEqual, numOfAnothermockMetrics)
					Convey("version should be the latest", func() {
						for _, mt := range mts {
							So(mt.Version(), ShouldEqual, 5)
						}
					})
				})
				Convey("get the queried version", func() {
					mts, err := trie.GetMetrics([]string{"intel", "anothermock", "*"}, 2)
					So(err, ShouldBeNil)
					So(len(mts), ShouldEqual, numOfAnothermockMetrics)
					Convey("version should be the queried version", func() {
						for _, mt := range mts {
							So(mt.Version(), ShouldEqual, 2)
						}
					})
				})
			})
			Convey("when the requested namespace is /intel/*", func() {
				Convey("get the latest version", func() {
					mts, err := trie.GetMetrics([]string{"intel", "*"}, -1)
					So(err, ShouldBeNil)
					So(len(mts), ShouldEqual, numOfAllMetrics)
					Convey("version should be the latest", func() {
						for _, mt := range mts {
							So(mt.Version(), ShouldEqual, 5)
						}
					})
				})
				Convey("get the queried version", func() {
					mts, err := trie.GetMetrics([]string{"intel", "*"}, 2)
					So(err, ShouldBeNil)
					So(len(mts), ShouldEqual, numOfAllMetrics)
					Convey("version should equals the queried version", func() {
						for _, mt := range mts {
							So(mt.Version(), ShouldEqual, 2)
						}
					})
				})
			})
			Convey("when the requested namespace is /intel/*/foo", func() {
				Convey("get the latest version", func() {
					mts, err := trie.GetMetrics([]string{"intel", "*", "foo"}, -1)
					So(err, ShouldBeNil)
					//expected two metrics: `/intel/mock/foo` and `/intel/anothermock/foo`
					So(len(mts), ShouldEqual, 2)
					Convey("version should be the latest", func() {
						for _, mt := range mts {
							So(mt.Version(), ShouldEqual, 5)
						}
					})
				})
				Convey("get the queried version", func() {
					mts, err := trie.GetMetrics([]string{"intel", "*", "foo"}, 2)
					So(err, ShouldBeNil)
					//expected two metrics: `/intel/mock/foo` and `/intel/anothermock/foo`
					So(len(mts), ShouldEqual, 2)
					Convey("version should be the queried version", func() {
						for _, mt := range mts {
							So(mt.Version(), ShouldEqual, 2)
						}
					})
				})
			})
			Convey("when the requested namespace is /intel/*/*/baz", func() {
				Convey("get the latest version", func() {
					mts, err := trie.GetMetrics([]string{"intel", "*", "*", "baz"}, -1)
					So(err, ShouldBeNil)
					//expected two metrics: `/intel/mock/*/baz` and `/intel/anothermock/*/baz`
					So(len(mts), ShouldEqual, 2)
					Convey("version should be the latest", func() {
						for _, mt := range mts {
							So(mt.Version(), ShouldEqual, 5)
						}
					})
				})
				Convey("get the queried version", func() {
					mts, err := trie.GetMetrics([]string{"intel", "*", "*", "baz"}, 2)
					So(err, ShouldBeNil)
					//expected two metrics: `/intel/mock/*/baz` and `/intel/anothermock/*/baz`
					So(len(mts), ShouldEqual, 2)
					Convey("version should be the queried version", func() {
						for _, mt := range mts {
							So(mt.Version(), ShouldEqual, 2)
						}
					})
				})
			})
		})
	})
}

func TestTrie_GetMetric(t *testing.T) {
	Convey("Simply get metric", t, func() {
		Convey("adding nodes to mttrie", func() {
			trie := NewMTTrie()
			lp2 := new(loadedPlugin)
			lp2.Meta.Version = 2

			lp5 := new(loadedPlugin)
			lp5.Meta.Version = 5
			mtstatic2 := newMetricType(core.NewNamespace("intel", "mock", "foo"), time.Now(), lp2)
			mtstatic5 := newMetricType(core.NewNamespace("intel", "mock", "foo"), time.Now(), lp5)

			mtdynamic2 := newMetricType(core.NewNamespace("intel", "mock").AddDynamicElement("host", "host id").AddStaticElement("baz"), time.Now(), lp2)
			mtdynamic5 := newMetricType(core.NewNamespace("intel", "mock").AddDynamicElement("host", "host id").AddStaticElement("baz"), time.Now(), lp5)

			trie.Add(mtstatic2)
			trie.Add(mtstatic5)
			trie.Add(mtdynamic2)
			trie.Add(mtdynamic5)

			Convey("for requested static metric", func() {
				Convey("get the latest version", func() {
					mt, err := trie.GetMetric([]string{"intel", "mock", "foo"}, -1)
					So(err, ShouldBeNil)
					So(mt, ShouldEqual, mtstatic5)
				})
				Convey("get the queried version", func() {
					mt, err := trie.GetMetric([]string{"intel", "mock", "foo"}, 2)
					So(err, ShouldBeNil)
					So(mt, ShouldEqual, mtstatic2)
				})
				Convey("error: the queried version of metric cannot be found", func() {
					mt, err := trie.GetMetric([]string{"intel", "mock", "foo"}, 6)
					So(err, ShouldNotBeNil)
					So(mt, ShouldBeNil)
					So(err.Error(), ShouldContainSubstring, "Metric not found: /intel/mock/foo (version: 6)")
				})
				Convey("error: the queried metric cannot be found", func() {
					mt, err := trie.GetMetric([]string{"intel", "mock", "invalid"}, -1)
					So(err, ShouldNotBeNil)
					So(mt, ShouldBeNil)
					So(err.Error(), ShouldContainSubstring, "Metric not found: /intel/mock/invalid (version: -1)")
				})
			})

			Convey("for requested dynamic metric", func() {
				Convey("get the latest version", func() {
					mt, err := trie.GetMetric([]string{"intel", "mock", "*", "baz"}, -1)
					So(err, ShouldBeNil)
					So(mt, ShouldEqual, mtdynamic5)
				})
				Convey("get the queried version", func() {
					mt, err := trie.GetMetric([]string{"intel", "mock", "*", "baz"}, 2)
					So(err, ShouldBeNil)
					So(mt, ShouldEqual, mtdynamic2)
				})
				Convey("error: the queried version of metric cannot be found", func() {
					mt, err := trie.GetMetric([]string{"intel", "mock", "*", "baz"}, 6)
					So(err, ShouldNotBeNil)
					So(mt, ShouldBeNil)
					So(err.Error(), ShouldContainSubstring, "Metric not found: /intel/mock/*/baz (version: 6)")
				})
				Convey("error: the queried metric cannot be found", func() {
					mt, err := trie.GetMetric([]string{"intel", "mock", "*", "invalid"}, -1)
					So(err, ShouldNotBeNil)
					So(mt, ShouldBeNil)
					So(err.Error(), ShouldContainSubstring, "Metric not found: /intel/mock/*/invalid (version: -1)")
				})
			})

			Convey("for incorrect format of namespace", func() {
				Convey("error: incorrect format of requested metric", func() {
					mt, err := trie.GetMetric([]string{}, 6)
					So(err, ShouldNotBeNil)
					So(mt, ShouldBeNil)
					So(err.Error(), ShouldContainSubstring, "Incorrect format of requested metric, empty list of namespace elements")
				})
			})
		})
	})
	Convey("improper usage of GetMetric - requested namespace fulfills more than one metric's namespace", t, func() {
		Convey("adding nodes to mttrie", func() {
			trie := NewMTTrie()
			lp2 := new(loadedPlugin)
			lp2.Meta.Version = 2

			mts := []*metricType{
				newMetricType(core.NewNamespace("intel", "mock", "foo"), time.Now(), lp2),
				newMetricType(core.NewNamespace("intel", "mock", "baz"), time.Now(), lp2),
			}

			for _, m := range mts {
				trie.Add(m)
			}

			Convey("error: the requested namespace contains a query", func() {
				mt, err := trie.GetMetric([]string{"intel", "mock", "*"}, -1)
				So(err, ShouldNotBeNil)
				So(mt, ShouldBeNil)
				So(err.Error(), ShouldEqual, "Incoming namespace `/intel/mock/*` is too ambiguous (version: -1)")
			})
		})
	})
}

func TestTrie_GetVersions(t *testing.T) {
	Convey("Adding nodes to mttrie", t, func() {
		trie := NewMTTrie()
		lp2 := new(loadedPlugin)
		lp2.Meta.Version = 2

		lp5 := new(loadedPlugin)
		lp5.Meta.Version = 5

		mtstatic2 := newMetricType(core.NewNamespace("intel", "mock", "foo"), time.Now(), lp2)
		mtstatic5 := newMetricType(core.NewNamespace("intel", "mock", "foo"), time.Now(), lp5)

		mtdynamic2 := newMetricType(core.NewNamespace("intel", "mock").AddDynamicElement("host", "host id").AddStaticElement("baz"), time.Now(), lp2)
		mtdynamic5 := newMetricType(core.NewNamespace("intel", "mock").AddDynamicElement("host", "host id").AddStaticElement("baz"), time.Now(), lp5)

		trie.Add(mtstatic2)
		trie.Add(mtstatic5)

		trie.Add(mtdynamic2)
		trie.Add(mtdynamic5)

		Convey("for requested static metric", func() {
			Convey("get all available versions", func() {
				mts, err := trie.GetVersions([]string{"intel", "mock", "foo"})
				So(err, ShouldBeNil)
				So(len(mts), ShouldEqual, 2)
				So(mts, ShouldContain, mtstatic2)
				So(mts, ShouldContain, mtstatic5)
			})
			Convey("error: the queried metric cannot be found", func() {
				mts, err := trie.GetVersions([]string{"intel", "mock", "invalid"})
				So(err, ShouldNotBeNil)
				So(mts, ShouldBeNil)
				So(err.Error(), ShouldContainSubstring, "Metric not found: /intel/mock/invalid")
			})
		})
		Convey("for requested dynamic metric", func() {
			Convey("get all available versions", func() {
				mts, err := trie.GetVersions([]string{"intel", "mock", "*", "baz"})
				So(err, ShouldBeNil)
				So(len(mts), ShouldEqual, 2)
				So(mts, ShouldContain, mtdynamic2)
				So(mts, ShouldContain, mtdynamic5)
			})
			Convey("error: the queried metric cannot be found", func() {
				mts, err := trie.GetVersions([]string{"intel", "mock", "*", "invalid"})
				So(err, ShouldNotBeNil)
				So(mts, ShouldBeNil)
				So(err.Error(), ShouldContainSubstring, "Metric not found: /intel/mock/*/invalid")
			})
		})
		Convey("for requested query", func() {
			Convey("get all available versions", func() {
				mts, err := trie.GetVersions([]string{"intel", "*"})
				So(err, ShouldBeNil)
				So(len(mts), ShouldEqual, 4)
				So(mts, ShouldContain, mtstatic2)
				So(mts, ShouldContain, mtstatic5)
				So(mts, ShouldContain, mtdynamic2)
				So(mts, ShouldContain, mtdynamic5)
			})
			Convey("error: the queried metric cannot be found", func() {
				mts, err := trie.GetVersions([]string{"invalid", "*"})
				So(err, ShouldNotBeNil)
				So(mts, ShouldBeNil)
				So(err.Error(), ShouldContainSubstring, "Metric not found: /invalid/*")
			})
		})

		Convey("for incorrect format of namespace", func() {
			Convey("error: incorrect format of requested metric", func() {
				mts, err := trie.GetVersions([]string{})
				So(err, ShouldNotBeNil)
				So(mts, ShouldBeNil)
				So(err.Error(), ShouldContainSubstring, "Incorrect format of requested metric, empty list of namespace elements")
			})
		})
	})
}
