package mttrie

import (
	"testing"
	"time"

	"github.com/intelsdilabs/pulse/core/cdata"

	. "github.com/smartystreets/goconvey/convey"
)

type mockMetricType struct {
	namespace []string
	version   int
}

func (m mockMetricType) Namespace() []string           { return m.namespace }
func (m mockMetricType) LastAdvertisedTime() time.Time { return time.Now() }
func (m mockMetricType) Version() int                  { return m.version }
func (m mockMetricType) Config() *cdata.ConfigDataNode { return nil }

func TestTrie(t *testing.T) {
	Convey("Make a new trie", t, func() {
		trie := New()
		So(trie, ShouldNotBeNil)
		So(trie, ShouldHaveSameTypeAs, &MTTrie{})
	})
	Convey("Fetch", t, func() {
		trie := New()
		Convey("Add and collect split namespace", func() {
			trie.Add([]string{"intel", "foo"}, mockMetricType{namespace: []string{"intel", "foo"}})
			trie.Add([]string{"intel", "baz", "qux"}, mockMetricType{namespace: []string{"intel", "baz", "qux"}})
			in, err := trie.Fetch([]string{"intel"})
			So(err, ShouldBeNil)
			So(len(in), ShouldEqual, 2)
			for _, mt := range in {
				So(mt, ShouldNotBeNil)
				So(mt.Namespace(), ShouldHaveSameTypeAs, []string{""})
			}
		})
		Convey("Add and collect with nodes with children", func() {
			trie.Add([]string{"intel", "foo", "bar"}, mockMetricType{namespace: []string{"intel", "foo", "bar"}})
			trie.Add([]string{"intel", "foo"}, mockMetricType{namespace: []string{"intel", "foo"}})
			trie.Add([]string{"intel", "baz", "qux"}, mockMetricType{namespace: []string{"intel", "baz", "qux"}})
			in, err := trie.Fetch([]string{"intel"})
			So(err, ShouldBeNil)
			So(len(in), ShouldEqual, 3)
		})
		Convey("Add and collect at node with mt and children", func() {
			trie.Add([]string{"intel", "foo", "bar"}, mockMetricType{namespace: []string{"intel", "foo", "bar"}})
			trie.Add([]string{"intel", "foo"}, mockMetricType{namespace: []string{"intel", "foo"}})
			in, err := trie.Fetch([]string{"intel", "foo"})
			So(err, ShouldBeNil)
			So(len(in), ShouldEqual, 2)
		})
		Convey("add and collect single depth namespace", func() {
			trie.Add([]string{"test"}, mockMetricType{namespace: []string{"test"}})
			t, err := trie.Fetch([]string{"test"})
			So(err, ShouldBeNil)
			So(t[0].Namespace(), ShouldResemble, []string{"test"})
		})
		Convey("add and longer length with single child", func() {
			trie.Add([]string{"d", "a", "n", "b", "a", "r"}, mockMetricType{namespace: []string{"d", "a", "n", "b", "a", "r"}})
			d, err := trie.Fetch([]string{"d", "a", "n", "b", "a", "r"})
			So(err, ShouldBeNil)
			So(d[0].Namespace(), ShouldResemble, []string{"d", "a", "n", "b", "a", "r"})
			dd, err := trie.Fetch([]string{"d", "a", "n"})
			So(err, ShouldBeNil)
			So(dd[0].Namespace(), ShouldResemble, []string{"d", "a", "n", "b", "a", "r"})
		})
		Convey("Mulitple versions", func() {
			trie.Add([]string{"intel", "foo"}, mockMetricType{
				namespace: []string{"intel", "foo"},
				version:   1,
			})
			trie.Add([]string{"intel", "foo"}, mockMetricType{
				namespace: []string{"intel", "foo"},
				version:   2,
			})
			n, err := trie.Fetch([]string{"intel"})
			So(err, ShouldBeNil)
			So(len(n), ShouldEqual, 2)
		})
		Convey("Fetch with error: not found", func() {
			_, err := trie.Fetch([]string{"not", "present"})
			So(err, ShouldNotBeNil)
			So(err, ShouldResemble, ErrNotFound)
		})
		Convey("Fetch with error: depth exceeded", func() {
			trie.Add([]string{"intel", "baz", "qux"}, mockMetricType{namespace: []string{"intel", "baz", "qux"}})
			_, err := trie.Fetch([]string{"intel", "baz", "qux", "foo"})
			So(err, ShouldNotBeNil)
			So(err, ShouldResemble, ErrNotFound)
		})
	})
	Convey("Get", t, func() {
		trie := New()
		Convey("simple get", func() {
			trie.Add([]string{"intel", "foo"}, mockMetricType{
				namespace: []string{"intel", "foo"},
			})
			n, err := trie.Get([]string{"intel", "foo"})
			So(err, ShouldBeNil)
			So(len(n), ShouldEqual, 1)
			So(n[0].Namespace(), ShouldResemble, []string{"intel", "foo"})
		})
		Convey("get with multiple at node", func() {
			trie.Add([]string{"intel", "foo"}, mockMetricType{
				namespace: []string{"intel", "foo"},
			})
			trie.Add([]string{"intel", "foo"}, mockMetricType{
				namespace: []string{"intel", "foo"},
				version:   1,
			})
			n, err := trie.Get([]string{"intel", "foo"})
			So(err, ShouldBeNil)
			So(len(n), ShouldEqual, 2)
		})
		Convey("error: node doesn't exist", func() {
			n, err := trie.Get([]string{"intel", "foo"})
			So(n, ShouldBeNil)
			So(err, ShouldEqual, ErrNotFound)
		})
		Convey("error: no data at node", func() {
			trie.Add([]string{"intel", "foo", "bar"}, mockMetricType{
				namespace: []string{"intel", "foo", "bar"},
			})
			n, err := trie.Get([]string{"intel", "foo"})
			So(n, ShouldBeNil)
			So(err, ShouldEqual, ErrNotFound)
		})
	})
}
