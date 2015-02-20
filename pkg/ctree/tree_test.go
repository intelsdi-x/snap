package ctree

import (
	"fmt"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

type dummyNode struct {
	data string
}

func (d *dummyNode) Merge(dn Node) {
	d.data = fmt.Sprintf("%s/%s", d.data, dn.(*dummyNode).data)
}

func newDummyNode() *dummyNode {
	return new(dummyNode)
}

func TestConfigTree(t *testing.T) {
	Convey("NewConfigTree()", t, func() {
		Convey("returns a pointer to a ConfigTree", func() {
			t := New()
			So(t, ShouldHaveSameTypeAs, new(ConfigTree))
		})
	})

	Convey("Add()", t, func() {
		c := New()
		c.Add([]string{"intel", "sdilabs", "joel", "dan", "nick", "justin", "sarah"}, newDummyNode())
		c.Add([]string{"intel", "sdilabs", "joel", "dan"}, newDummyNode())
		c.Add([]string{"intel", "manhole", "joel", "dan"}, newDummyNode())
	})

	Convey("Get()", t, func() {
		Convey("order preserved", func() {
			d1 := newDummyNode()
			d1.data = "a"
			d2 := newDummyNode()
			d2.data = "b"
			d3 := newDummyNode()
			d3.data = "c"
			d4 := newDummyNode()
			d4.data = "d"
			c := New()
			c.Add([]string{"intel", "foo", "sdilabs", "joel", "dan", "nick", "justin", "sarah"}, d1)
			c.Add([]string{"intel", "foo", "sdilabs", "joel", "dan"}, d2)
			c.Add([]string{"intel", "foo", "manhole", "joel", "dan"}, d3)
			c.Add([]string{"intel", "foo", "manhole", "joel", "dan", "mark"}, d4)
			c.Freeze()
			g := c.Get([]string{"intel", "foo", "sdilabs", "joel", "dan", "nick", "justin", "sarah"})
			So(g.(*dummyNode).data, ShouldResemble, "b/a")
			So(g, ShouldNotBeNil)
		})

		Convey("blank tree return nil", func() {
			c := New()
			c.Freeze()
			n := c.Get([]string{"intel", "foo", "sdilabs", "joel", "dan", "nick", "justin", "sarah"})
			So(n, ShouldBeNil)
		})

		Convey("long ns short tree", func() {
			c := New()
			d1 := newDummyNode()
			d1.data = "a"
			c.Add([]string{"intel"}, d1)
			c.Freeze()
			n := c.Get([]string{"intel", "foo", "sdilabs", "joel", "dan", "nick", "justin", "sarah"})
			So(n, ShouldBeNil)
		})

		Convey("long tree short ns", func() {
			c := New()
			d1 := newDummyNode()
			d1.data = "a"
			c.Add([]string{"intel", "foo", "sdilabs", "joel", "dan", "nick", "justin", "sarah"}, d1)
			c.Freeze()
			n := c.Get([]string{"intel"})
			So(n, ShouldBeNil)
		})

		Convey("basic get", func() {
			d1 := newDummyNode()
			d1.data = "a"
			d2 := newDummyNode()
			d2.data = "b"
			c := New()
			c.Add([]string{"intel", "foo", "sdilabs", "joel", "dan", "nick", "justin", "sarah"}, d1)
			c.Add([]string{"intel", "foo", "sdilabs", "joel", "dan"}, d2)
			c.Freeze()
			g := c.Get([]string{"intel", "foo", "sdilabs", "joel", "dan", "nick", "justin", "sarah"})
			So(g, ShouldNotBeNil)
		})

		Convey("get is inbetween two nodes in tree", func() {
			d1 := newDummyNode()
			d1.data = "a"
			d2 := newDummyNode()
			d2.data = "b"
			c := New()
			c.Add([]string{"intel", "foo", "sdilabs", "joel", "dan", "nick", "justin", "sarah"}, d1)
			c.Add([]string{"intel", "foo", "sdilabs"}, d2)
			c.Freeze()
			g := c.Get([]string{"intel", "foo", "sdilabs", "joel"})
			So(g, ShouldNotBeNil)
		})

		Convey("adding a new root panics", func() {
			d1 := newDummyNode()
			d1.data = "a"
			d2 := newDummyNode()
			d2.data = "b"
			c := New()
			c.Add([]string{"intel", "foo", "sdilabs", "joel", "dan", "nick", "justin", "sarah"}, d1)

			So(func() {
				c.Add([]string{"mashery", "foo", "sdilabs", "joel", "dan", "nick", "justin", "sarah"}, d2)
			}, ShouldPanic)
		})

		Convey("doesn't panic on empty ns", func() {
			d1 := newDummyNode()
			d1.data = "a"
			c := New()
			So(func() {
				c.Add([]string{}, d1)
				g := c.Get([]string{})
				So(g, ShouldBeNil)
			}, ShouldNotPanic)
		})
	})

	Convey("Frozen()", t, func() {
		c := New()
		c.Freeze()
		So(c.Frozen(), ShouldBeTrue)
	})

}
