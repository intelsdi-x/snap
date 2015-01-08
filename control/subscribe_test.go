package control

import (
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

func TestAdd(t *testing.T) {
	Convey("Increments the counter", t, func() {
		s := new(subscriptions)
		s.Add()
		So(s.Count(), ShouldEqual, 1)
	})
}

func TestRemove(t *testing.T) {
	Convey("when count is 1", t, func() {
		s := new(subscriptions)
		s.Add()
		Convey("then it decrements the counter", func() {
			s.Remove()
			So(s.Count(), ShouldEqual, 0)
		})
	})
	Convey("when count is 0", t, func() {
		s := new(subscriptions)
		s.Remove()
		Convey("then it returns an error", func() {
			err := s.Remove()
			So(err, ShouldNotBeNil)
		})
	})
}
