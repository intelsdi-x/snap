package ctree

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

type dummyNode struct {
	data map[string]string
}

func (d *dummyNode) Data() interface{} {
	return d.data
}

func (d *dummyNode) Merge(dn Node) {
	m := dn.Data().(map[string]string)
	for k, v := range m {
		d.data[k] = v
	}
}

func newDummyNode() *dummyNode {
	return &dummyNode{
		data: make(map[string]string),
	}
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
		d1 := newDummyNode()
		d1.data["foo"] = "bar"
		d2 := newDummyNode()
		d2.data["foo"] = "baz"
		c := New()
		c.Add([]string{"intel", "sdilabs", "joel", "dan", "nick", "justin", "sarah"}, d1)
		c.Add([]string{"intel", "sdilabs", "joel", "dan"}, d2)
		c.Freeze()
		c.Get([]string{"intel", "sdilabs", "joel", "dan", "nick", "justin", "sarah"})
	})

}
