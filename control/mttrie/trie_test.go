package mttrie

import (
	"testing"
	"time"

	"github.com/intelsdilabs/pulse/core/cdata"

	. "github.com/smartystreets/goconvey/convey"
)

type mockMetricType struct {
	namespace []string
}

func (m mockMetricType) Namespace() []string           { return m.namespace }
func (m mockMetricType) LastAdvertisedTime() time.Time { return time.Now() }
func (m mockMetricType) Version() int                  { return 1 }
func (m mockMetricType) Config() *cdata.ConfigDataNode { return nil }

func TestTrie(t *testing.T) {
	trie := New()
	Convey("Make a new trie", t, func() {
		So(trie, ShouldNotBeNil)
		So(trie, ShouldHaveSameTypeAs, &MTTrie{})
	})
	Convey("Add and get split namespace", t, func() {
		trie.Add([]string{"intel", "foo"}, mockMetricType{namespace: []string{"intel", "foo"}})
		trie.Add([]string{"intel", "baz", "qux"}, mockMetricType{namespace: []string{"intel", "baz", "qux"}})
		in, err := trie.Get([]string{"intel"})
		So(err, ShouldBeNil)
		So(len(in), ShouldEqual, 2)
		for _, mt := range in {
			So(mt, ShouldNotBeNil)
			So(mt.Namespace(), ShouldHaveSameTypeAs, []string{""})
		}
	})
	Convey("add and get single depth namespace", t, func() {
		trie.Add([]string{"test"}, mockMetricType{namespace: []string{"test"}})
		t, err := trie.Get([]string{"test"})
		So(err, ShouldBeNil)
		So(t[0].Namespace(), ShouldResemble, []string{"test"})
	})
	Convey("add and get longer length with single child", t, func() {
		trie.Add([]string{"d", "a", "n", "b", "a", "r"}, mockMetricType{namespace: []string{"d", "a", "n", "b", "a", "r"}})
		d, err := trie.Get([]string{"d", "a", "n", "b", "a", "r"})
		So(err, ShouldBeNil)
		So(d[0].Namespace(), ShouldResemble, []string{"d", "a", "n", "b", "a", "r"})
		dd, err := trie.Get([]string{"d", "a", "n"})
		So(err, ShouldBeNil)
		So(dd[0].Namespace(), ShouldResemble, []string{"d", "a", "n", "b", "a", "r"})
	})
	Convey("Overwrite existing node", t, func() {
		trie.Add([]string{"intel", "foo"}, mockMetricType{namespace: []string{"intel", "foo", "overwrite"}})
		n, err := trie.Get([]string{"intel", "foo"})
		So(err, ShouldBeNil)
		So(n[0].Namespace(), ShouldResemble, []string{"intel", "foo", "overwrite"})
	})
	Convey("Get with error: not found", t, func() {
		_, err := trie.Get([]string{"not", "present"})
		So(err, ShouldNotBeNil)
		So(err, ShouldResemble, ErrNotFound)
	})
	Convey("Get with error: depth exceeded", t, func() {
		trie.Add([]string{"intel", "baz", "qux"}, mockMetricType{namespace: []string{"intel", "baz", "qux"}})
		_, err := trie.Get([]string{"intel", "baz", "qux", "foo"})
		So(err, ShouldNotBeNil)
		So(err, ShouldResemble, ErrNotFound)
	})
}
