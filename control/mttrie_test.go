package control

import (
	"testing"
	"time"

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
			mt := newMetricType([]string{"intel", "foo"}, time.Now(), new(loadedPlugin))
			mt2 := newMetricType([]string{"intel", "baz", "qux"}, time.Now(), new(loadedPlugin))
			trie.Add(mt)
			trie.Add(mt2)

			in, err := trie.Fetch([]string{"intel"})
			So(err, ShouldBeNil)
			So(len(in), ShouldEqual, 2)
			for _, mt := range in {
				So(mt, ShouldNotBeNil)
				So(mt.Namespace(), ShouldHaveSameTypeAs, []string{""})
			}
		})
		Convey("Add and collect with nodes with children", func() {
			mt := newMetricType([]string{"intel", "foo", "bar"}, time.Now(), new(loadedPlugin))
			mt2 := newMetricType([]string{"intel", "foo"}, time.Now(), new(loadedPlugin))
			mt3 := newMetricType([]string{"intel", "foo", "qux"}, time.Now(), new(loadedPlugin))
			trie.Add(mt)
			trie.Add(mt2)
			trie.Add(mt3)

			in, err := trie.Fetch([]string{"intel"})
			So(err, ShouldBeNil)
			So(len(in), ShouldEqual, 3)
		})
		Convey("Add and collect at node with mt and children", func() {
			mt := newMetricType([]string{"intel", "foo", "bar"}, time.Now(), new(loadedPlugin))
			mt2 := newMetricType([]string{"intel", "foo"}, time.Now(), new(loadedPlugin))
			trie.Add(mt)
			trie.Add(mt2)

			in, err := trie.Fetch([]string{"intel", "foo"})
			So(err, ShouldBeNil)
			So(len(in), ShouldEqual, 2)
		})
		Convey("add and collect single depth namespace", func() {
			mt := newMetricType([]string{"test"}, time.Now(), new(loadedPlugin))
			trie.Add(mt)
			t, err := trie.Fetch([]string{"test"})
			So(err, ShouldBeNil)
			So(t[0].Namespace(), ShouldResemble, []string{"test"})
		})
		Convey("add and longer length with single child", func() {
			mt := newMetricType([]string{"d", "a", "n", "b", "a", "r"}, time.Now(), new(loadedPlugin))
			trie.Add(mt)
			d, err := trie.Fetch([]string{"d", "a", "n", "b", "a", "r"})
			So(err, ShouldBeNil)
			So(d[0].Namespace(), ShouldResemble, []string{"d", "a", "n", "b", "a", "r"})
			dd, err := trie.Fetch([]string{"d", "a", "n"})
			So(err, ShouldBeNil)
			So(dd[0].Namespace(), ShouldResemble, []string{"d", "a", "n", "b", "a", "r"})
		})
		Convey("Mulitple versions", func() {
			lp := new(loadedPlugin)
			lp.Meta.Version = 1
			mt := newMetricType([]string{"intel", "foo"}, time.Now(), lp)
			lp2 := new(loadedPlugin)
			lp2.Meta.Version = 2
			mt2 := newMetricType([]string{"intel", "foo"}, time.Now(), lp2)
			trie.Add(mt)
			trie.Add(mt2)
			n, err := trie.Fetch([]string{"intel"})
			So(err, ShouldBeNil)
			So(len(n), ShouldEqual, 2)
		})
		Convey("Fetch with error: not found", func() {
			_, err := trie.Fetch([]string{"not", "present"})
			So(err, ShouldNotBeNil)
			So(err, ShouldResemble, errorMetricNotFound([]string{"not", "present"}))
		})
		Convey("Fetch with error: depth exceeded", func() {
			lp := new(loadedPlugin)
			lp.Meta.Version = 1
			mt := newMetricType([]string{"intel", "foo"}, time.Now(), lp)
			trie.Add(mt)
			_, err := trie.Fetch([]string{"intel", "foo", "bar", "baz"})
			So(err, ShouldNotBeNil)
			So(err, ShouldResemble, errorMetricNotFound([]string{"intel", "foo", "bar", "baz"}))
		})
	})
	Convey("Get", t, func() {
		trie := NewMTTrie()
		Convey("simple get", func() {
			lp := new(loadedPlugin)
			lp.Meta.Version = 1
			mt := newMetricType([]string{"intel", "foo"}, time.Now(), lp)
			trie.Add(mt)
			n, err := trie.Get([]string{"intel", "foo"})
			So(err, ShouldBeNil)
			So(len(n), ShouldEqual, 1)
			So(n[0].Namespace(), ShouldResemble, []string{"intel", "foo"})
		})
		Convey("error: no data at node", func() {
			lp := new(loadedPlugin)
			lp.Meta.Version = 1
			mt := newMetricType([]string{"intel", "foo"}, time.Now(), lp)
			trie.Add(mt)
			n, err := trie.Get([]string{"intel"})
			So(n, ShouldBeNil)
			So(err, ShouldResemble, errorMetricNotFound([]string{"intel"}))
		})
	})
}
