package ctree

import (
	"fmt"
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
		c.Add([]string{"intel", "sdilabs", "joel", "dan", "nick", "justin", "sarah"})

		fmt.Printf("%v\n", *c.root)
	})

	Convey("Get()", t, func() {})

}

func TestNode(t *testing.T) {
	Convey("add()", t, func() {})
}
