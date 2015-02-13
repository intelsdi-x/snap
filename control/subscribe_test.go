package control

import (
	"errors"
	"sync"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestnewSubscriptionTable(t *testing.T) {
	Convey("it initializes the pieces of the table correctly", t, func() {
		s := newSubscriptionsTable()
		So(s.table, ShouldHaveSameTypeAs, &map[string]int{})
		So(s.mutex, ShouldHaveSameTypeAs, &sync.Mutex{})
	})
}

func TestSubscribe(t *testing.T) {
	s := newSubscriptionsTable()
	Convey("when the metric is not in the table", t, func() {
		Convey("then it gets added to the table", func() {
			s.Subscribe("test.foo")
			So(s.Count("test.foo"), ShouldEqual, 1)
		})
	})
	Convey("when the metric is in the table", t, func() {
		Convey("then it gets correctly increments the count", func() {
			s.Subscribe("test.foo")
			So(s.Count("test.foo"), ShouldEqual, 2)
		})
		Convey("then it does not add it twice to the keys array", func() {
			s.Subscribe("test.foo")
			So(len(*s.table), ShouldEqual, len(*s.keys))
		})
	})
}

func TestUnsubscribe(t *testing.T) {
	s := newSubscriptionsTable()
	Convey("when the metric is in the table", t, func() {
		s.Subscribe("test.foo")
		Convey("then its subscription count is decremented", func() {
			s.Unsubscribe("test.foo")
			So(s.Count("test.foo"), ShouldEqual, 0)
		})
	})
	Convey("when the metric is not in the table", t, func() {
		Convey("then it returns the correct error", func() {
			err := s.Unsubscribe("test.bar")
			So(err, ShouldResemble, errors.New("subscription count cannot be less than 0 for key test.bar"))
		})
	})
	Convey("when the metric's count is already 0", t, func() {
		s.Subscribe("test.bar")
		s.Unsubscribe("test.bar")
		Convey("then it returns the correct error", func() {
			err := s.Unsubscribe("test.bar")
			So(err, ShouldResemble, errors.New("subscription count cannot be less than 0 for key test.bar"))
		})
	})
}

func TestValue(t *testing.T) {
	Convey("when there are items in the table", t, func() {
		s := newSubscriptionsTable()
		s.Subscribe("test.foo")
		Convey("then it can retrieve them by the index (.curentIter)", func() {
			s.currentIter = 1
			key, val := s.Item()
			So(key, ShouldEqual, "test.foo")
			So(val, ShouldEqual, 1)
		})
	})
}

func TestNext(t *testing.T) {
	Convey("when there are items in the table", t, func() {
		s := newSubscriptionsTable()
		s.Subscribe("test.foo")
		s.Subscribe("test.bar")
		s.Unsubscribe("test.bar")
		s.Subscribe("test.baz")
		s.Subscribe("test.baz")
		Convey("then it reports accurately whether or not there are additional items to iterate", func() {
			s.currentIter = 0
			So(s.Next(), ShouldEqual, true)
			s.currentIter = 1
			So(s.Next(), ShouldEqual, true)
			s.currentIter = 2
			So(s.Next(), ShouldEqual, true)
			s.currentIter = 3
			So(s.Next(), ShouldEqual, false)
		})
		Convey("then it can be used to iterate through the items", func() {
			s.currentIter = 0
			testmap := make(map[string]int)
			iters := 0
			for s.Next() {
				key, val := s.Item()
				testmap[key] = val
				iters++
			}
			So(testmap["test.foo"], ShouldEqual, 1)
			So(testmap["test.bar"], ShouldEqual, 0)
			So(testmap["test.baz"], ShouldEqual, 2)
			So(iters, ShouldEqual, 3)
		})
	})
}

func TestCount(t *testing.T) {
	s := newSubscriptionsTable()
	s.Subscribe("test.foo")
	s.Subscribe("test.bar")
	s.Unsubscribe("test.bar")
	s.Subscribe("test.baz")
	s.Subscribe("test.baz")
	Convey("it returns the correct count for a metric", t, func() {
		So(s.Count("test.foo"), ShouldEqual, 1)
		So(s.Count("test.bar"), ShouldEqual, 0)
		So(s.Count("test.baz"), ShouldEqual, 2)
		So(s.Count("test.qux"), ShouldEqual, 0)
	})
}
