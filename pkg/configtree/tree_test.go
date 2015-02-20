package ctree

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestConfigTree(t *testing.T) {
	Convey("NewConfigTree()", t, func() {
		Convey("returns a pointer to a ConfigTree", func() {
			t := NewConfigTree()
			So(t, ShouldHaveSameTypeAs, new(ConfigTree))
		})
	})

	Convey("Add()", t, func() {})

	Convey("Get()", t, func() {})

}

func TestNode(t *testing.T) {
	Convey("add()", t, func() {})
}
