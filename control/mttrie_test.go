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
			So(err.Error(), ShouldContainSubstring, "Metrics not found below a given namespace: /not/present")
		})
		Convey("Fetch with error: depth exceeded", func() {
			lp := new(loadedPlugin)
			lp.Meta.Version = 1
			mt := newMetricType(core.NewNamespace("intel", "foo"), time.Now(), lp)
			trie.Add(mt)
			_, err := trie.Fetch([]string{"intel", "foo", "bar", "baz"})
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "Metrics not found below a given namespace: /intel/foo/bar/baz")

		})
	})
	Convey("GetMetrics()", t, func() {
		trie := NewMTTrie()
		ver := -1

		Convey("simply get", func() {
			lp2 := new(loadedPlugin)
			lp2.Meta.Version = 2

			lp5 := new(loadedPlugin)
			lp5.Meta.Version = 5

			mt2 := newMetricType(core.NewNamespace("intel", "foo"), time.Now(), lp2)
			mt5 := newMetricType(core.NewNamespace("intel", "foo"), time.Now(), lp5)

			trie.Add(mt2)
			trie.Add(mt5)

			Convey("get the latest version", func() {
				n, err := trie.GetMetrics([]string{"intel", "foo"}, -1)
				So(err, ShouldBeNil)
				So(len(n), ShouldEqual, 1)
				So(n[0], ShouldEqual, mt5)
			})
			Convey("get the queried version", func() {
				n, err := trie.GetMetrics([]string{"intel", "foo"}, 2)
				So(err, ShouldBeNil)
				So(len(n), ShouldEqual, 1)
				So(n[0], ShouldEqual, mt2)
			})
			Convey("error: the queried version of metric cannot be found", func() {
				n, err := trie.GetMetrics([]string{"intel", "foo"}, 6)
				So(err, ShouldNotBeNil)
				So(n, ShouldBeEmpty)
				So(err.Error(), ShouldContainSubstring, "Metric not found: /intel/foo (version: 6)")
			})
		})
		Convey("error: no data at node", func() {
			lp := new(loadedPlugin)
			lp.Meta.Version = 1
			mt := newMetricType(core.NewNamespace("intel", "foo"), time.Now(), lp)
			trie.Add(mt)
			n, err := trie.GetMetrics([]string{"intel"}, ver)
			So(n, ShouldBeNil)
			So(err.Error(), ShouldContainSubstring, "Metric not found:")
		})
	})
}
