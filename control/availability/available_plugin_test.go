package availability

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestSubscribe(t *testing.T) {
	ap := &AvailablePlugin{
		sub: new(Subscriptions),
	}
	ap.Subscribe()
	Convey("adds a subscription", t, func() {
		So(ap.Subscriptions(), ShouldEqual, 1)
	})
}

func TestUnsubscribe(t *testing.T) {
	ap := &AvailablePlugin{
		sub: new(Subscriptions),
	}
	Convey("when there are subscriptions to unsubscribe", t, func() {
		ap.Subscribe()
		Convey("It removes a subscription", func() {
			ap.Unsubscribe()
			So(ap.Subscriptions(), ShouldEqual, 0)
		})
	})
	Convey("when there are not subscriptions to unsubscribe", t, func() {
		Convey("It returns an error", func() {
			err := ap.Unsubscribe()
			So(err, ShouldNotBeNil)
		})
	})
}

func TestSubscriptions(t *testing.T) {
	ap := &AvailablePlugin{
		sub: new(Subscriptions),
	}
	Convey("should accurately represent the subscription count", t, func() {
		for i := 0; i < 6; i++ {
			ap.Subscribe()
		}
		for i := 0; i < 3; i++ {
			ap.Unsubscribe()
		}
		So(ap.Subscriptions(), ShouldEqual, 3)
	})
}
