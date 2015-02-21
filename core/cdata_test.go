package core

import (
	"fmt"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestConfigDataTree(t *testing.T) {

	Convey("ConfigDataTree()", t, func() {

		Convey("New()", func() {
			cd := NewConfigDataTree()
			So(cd, ShouldNotBeNil)
			So(cd.cTree, ShouldNotBeNil)

			cd.Add([]string{"1"}, &ConfigDataNode{Value: "a"})
			cd.Add([]string{"1", "2"}, &ConfigDataNode{Value: "b"})

			cd.cTree.Print()

			a := cd.Get([]string{"1", "2"})
			fmt.Println(a)
		})

	})
}
