// +build legacy

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

package ctree

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

type mockNode struct {
	data string
}

func (d mockNode) Merge(dn Node) Node {
	d.data = fmt.Sprintf("%s/%s", d.data, dn.(*mockNode).data)
	return d
}

func (d *mockNode) GobEncode() ([]byte, error) {
	w := new(bytes.Buffer)
	encoder := gob.NewEncoder(w)

	if err := encoder.Encode(d.data); err != nil {
		return nil, err
	}

	return w.Bytes(), nil
}

func (d *mockNode) GobDecode(buf []byte) error {
	if len(buf) == 0 {
		//there is nothing to do
		return nil
	}
	r := bytes.NewBuffer(buf)
	decoder := gob.NewDecoder(r)
	return decoder.Decode(&d.data)
}

func newMockNode() *mockNode {
	return new(mockNode)
}

func TestConfigTree(t *testing.T) {
	Convey("New()", t, func() {
		Convey("returns a pointer to a ConfigTree", func() {
			t := New()
			So(t, ShouldHaveSameTypeAs, new(ConfigTree))
		})
	})

	Convey("Add()", t, func() {
		c := New()
		c.Add([]string{"intel", "sdilabs", "joel", "dan", "nick", "justin", "sarah"}, newMockNode())
		c.Add([]string{"intel", "sdilabs", "joel", "dan"}, newMockNode())
		c.Add([]string{"intel", "manhole", "joel", "dan"}, newMockNode())
	})

	Convey("GetAll()", t, func() {
		c := New()
		c.Add([]string{"intel", "sdilabs", "joel", "dan", "nick", "justin", "sarah"}, newMockNode())
		c.Add([]string{"intel", "sdilabs", "joel", "dan"}, newMockNode())
		c.Add([]string{"intel", "manhole", "joel", "dan"}, newMockNode())
		results := c.GetAll()
		So(results, ShouldNotBeNil)
		So(len(results), ShouldEqual, 3)
	})

	Convey("Get()", t, func() {
		Convey("order preserved", func() {
			d1 := newMockNode()
			d1.data = "a"
			d2 := newMockNode()
			d2.data = "b"
			d3 := newMockNode()
			d3.data = "c"
			d4 := newMockNode()
			d4.data = "d"
			c := New()
			c.Add([]string{"intel", "foo", "sdilabs", "joel", "dan", "nick", "justin", "sarah"}, d1)
			c.Add([]string{"intel", "foo", "sdilabs", "joel", "dan"}, d2)
			c.Add([]string{"intel", "foo", "manhole", "joel", "dan"}, d3)
			c.Add([]string{"intel", "foo", "manhole", "joel", "dan", "mark"}, d4)
			c.Print()
			g := c.Get([]string{"intel", "foo", "sdilabs", "joel", "dan", "nick", "justin", "sarah"})
			So(g, ShouldNotBeNil)
			So(g.(mockNode).data, ShouldResemble, "b/a")

			Convey("GobEncode/GobDecode", func() {
				gob.Register(&mockNode{})
				buf, err := c.GobEncode()
				So(err, ShouldBeNil)
				So(buf, ShouldNotBeNil)
				c2 := New()
				err = c2.GobDecode(buf)
				So(err, ShouldBeNil)
				So(c2.root, ShouldNotBeNil)
				So(c2.root.keys, ShouldNotBeEmpty)
				g2 := c2.Get([]string{"intel", "foo", "sdilabs", "joel", "dan", "nick", "justin", "sarah"})
				So(g2, ShouldNotBeNil)
				So(g2.(mockNode).data, ShouldResemble, "b/a")
			})
		})

		Convey("single item ns", func() {
			d1 := newMockNode()
			d1.data = "a"
			c := New()
			c.Add([]string{"1"}, d1)
			g := c.Get([]string{"1"})
			So(g, ShouldNotBeNil)
			So(g.(*mockNode).data, ShouldResemble, "a")
		})

		Convey("add 2 nodes that will not change on compression", func() {
			d1 := newMockNode()
			d1.data = "a"
			d2 := newMockNode()
			d2.data = "b"
			c := New()
			c.Debug = true
			c.Add([]string{"1"}, d1)
			c.Add([]string{"1", "2"}, d2)
			g := c.Get([]string{"1", "2"})
			So(g, ShouldNotBeNil)
			So(g.(mockNode).data, ShouldResemble, "a/b")
		})

		Convey("blank tree return nil", func() {
			c := New()
			n := c.Get([]string{"intel", "foo", "sdilabs", "joel", "dan", "nick", "justin", "sarah"})
			So(n, ShouldBeNil)
		})

		Convey("wrong root", func() {
			d1 := newMockNode()
			d1.data = "a"
			c := New()
			c.Add([]string{"1"}, d1)
			n := c.Get([]string{"2"})
			So(n, ShouldBeNil)
		})

		Convey("long ns short tree", func() {
			c := New()
			c.Debug = true
			d1 := newMockNode()
			d1.data = "a"
			c.Add([]string{"intel"}, d1)
			n := c.Get([]string{"intel", "foo", "sdilabs", "joel", "dan", "nick", "justin", "sarah"})
			So(n, ShouldNotBeNil)
		})

		Convey("long tree short ns", func() {
			c := New()
			d1 := newMockNode()
			d1.data = "a"
			c.Add([]string{"intel", "foo", "sdilabs", "joel", "dan", "nick", "justin", "sarah"}, d1)
			n := c.Get([]string{"intel"})
			So(n, ShouldBeNil)
		})

		Convey("basic get", func() {
			d1 := newMockNode()
			d1.data = "a"
			d2 := newMockNode()
			d2.data = "b"
			c := New()
			c.Add([]string{"intel", "foo", "sdilabs", "joel", "dan", "nick", "justin", "sarah"}, d1)
			c.Add([]string{"intel", "foo", "sdilabs", "joel", "dan"}, d2)
			g := c.Get([]string{"intel", "foo", "sdilabs", "joel", "dan", "nick", "justin", "sarah"})
			So(g, ShouldNotBeNil)
		})

		Convey("get is in between two nodes in tree", func() {
			d1 := newMockNode()
			d1.data = "a"
			d2 := newMockNode()
			d2.data = "b"
			c := New()
			c.Add([]string{"intel", "foo", "sdilabs", "joel", "dan", "nick", "justin", "sarah"}, d1)
			c.Add([]string{"intel", "foo", "sdilabs"}, d2)
			g := c.Get([]string{"intel", "foo", "sdilabs", "joel"})
			So(g, ShouldNotBeNil)
		})

		Convey("adding a new root panics", func() {
			d1 := newMockNode()
			d1.data = "a"
			d2 := newMockNode()
			d2.data = "b"
			c := New()
			c.Add([]string{"intel", "foo", "sdilabs", "joel", "dan", "nick", "justin", "sarah"}, d1)

			So(func() {
				c.Add([]string{"mashery", "foo", "sdilabs", "joel", "dan", "nick", "justin", "sarah"}, d2)
			}, ShouldPanic)
		})

		Convey("doesn't panic on ns where the root and ns don't have a policy", func() {
			d1 := newMockNode()
			d1.data = "a"
			d2 := newMockNode()
			d2.data = "b"
			c := New()
			c.Add([]string{"intel", "foo", "sdi-x", "cody"}, d1)
			c.Add([]string{"intel", "foo", "sdi-x", "nan"}, d2)
			So(func() {
				g := c.Get([]string{"intel", "foo", "sdi-x", "emily", "tiffany", "matt"})
				So(g, ShouldBeNil)
			}, ShouldNotPanic)
		})

		Convey("doesn't panic on empty ns", func() {
			d1 := newMockNode()
			d1.data = "a"
			c := New()
			So(func() {
				c.Add([]string{}, d1)
				g := c.Get([]string{})
				So(g, ShouldBeNil)
			}, ShouldNotPanic)
		})
	})
}
