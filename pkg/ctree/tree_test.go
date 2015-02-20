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
		fmt.Println("")
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
		c.print()
		g := c.Get([]string{"intel", "foo", "sdilabs", "joel", "dan", "nick", "justin", "sarah"})
		fmt.Println(g.(*dummyNode).data)
	})

}
