package ctree

import (
	// "fmt"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestConfigTree(t *testing.T) {
	Convey("NewConfigTree()", t, func() {
		Convey("returns a pointer to a ConfigTree", func() {
			t := New()
			So(t, ShouldHaveSameTypeAs, new(ConfigTree))
		})
	})

	Convey("Add()", t, func() {
		c := New()
		c.Add([]string{"intel", "sdilabs", "joel", "dan", "nick", "justin", "sarah"}, new(DummyNode))
		c.Add([]string{"intel", "sdilabs", "joel", "dan"}, new(DummyNode))
		c.Add([]string{"intel", "manhole", "joel", "dan"}, new(DummyNode))
		c.Freeze()
		c.print()
	})

	Convey("Get()", t, func() {
		d1 := new(DummyNode)
		d2 := new(DummyNode)
		// d3 := new(DummyNode)

		c := New()
		c.Add([]string{"intel", "sdilabs", "joel", "dan", "nick", "justin", "sarah"}, d1)
		c.Add([]string{"intel", "sdilabs", "joel", "dan"}, d2)
		// c.Add([]string{"intel", "manhole", "joel", "dan"}, d3)
		c.Freeze()
		c.print()

		c.Get([]string{"intel", "sdilabs", "joel", "dan", "nick", "justin", "sarah"})
	})

}

func TestNode(t *testing.T) {
	Convey("add()", t, func() {})
}
