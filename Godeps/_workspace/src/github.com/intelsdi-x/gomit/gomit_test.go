package gomit

import (
	. "github.com/smartystreets/goconvey/convey"
	"testing"
	"time"
)

type MockEventBody struct {
}

type MockThing struct {
	LastNamespace string
}

func (m *MockEventBody) Namespace() string {
	return "Mock.Event"
}

func (m *MockThing) HandleGomitEvent(e Event) {
	m.LastNamespace = e.Namespace()
}

func TestEmitter(t *testing.T) {
	Convey("gomit.Emitter", t, func() {

		Convey(".Emit", func() {
			// Ensure that we can silently emit an event when no one is handling.
			// and that the Emit() returns no error and 0 handlers.
			Convey("Emits with no Handlers", func() {
				event_controller := new(EventController)
				eb := new(MockEventBody)
				i, e := event_controller.Emit(eb)

				So(i, ShouldBeZeroValue)
				So(event_controller.HandlerCount(), ShouldEqual, 0)
				So(e, ShouldBeNil)
			})
			Convey("Emits with one Handlers", func() {
				event_controller := new(EventController)
				mt := new(MockThing)

				event_controller.RegisterHandler("m1", mt)
				eb := new(MockEventBody)
				i, e := event_controller.Emit(eb)

				So(i, ShouldEqual, 1)
				So(event_controller.HandlerCount(), ShouldEqual, 1)
				So(e, ShouldBeNil)
			})
		})

		Convey(".RegisterHandler", func() {
			Convey("Allows registration of a single Handler", func() {
				event_controller := new(EventController)
				mt := new(MockThing)
				e := event_controller.RegisterHandler("m1", mt)

				So(event_controller.HandlerCount(), ShouldEqual, 1)
				So(e, ShouldBeNil)
			})

			Convey("Does not allow a Handler to have more than one registration", func() {
				event_controller := new(EventController)
				mt1 := new(MockThing)
				mt2 := new(MockThing)

				event_controller.RegisterHandler("m1", mt1)
				// Should return error signifying it was already registered.
				e := event_controller.RegisterHandler("m1", mt2)

				So(event_controller.HandlerCount(), ShouldEqual, 1)
				So(e, ShouldNotBeNil)
			})
		})

		Convey(".HandlerCount", func() {
			// Some simple count testing
			Convey("Returns correct count", func() {
				event_controller := new(EventController)
				mt1 := new(MockThing)
				mt2 := MockThing{}
				mt3 := new(MockThing)
				mt4 := new(MockThing)
				mt5 := new(MockThing)

				e := event_controller.RegisterHandler("m1", mt1)

				So(e, ShouldBeNil)
				So(event_controller.HandlerCount(), ShouldEqual, 1)

				e = event_controller.RegisterHandler("m2", &mt2)
				So(e, ShouldBeNil)
				So(event_controller.HandlerCount(), ShouldEqual, 2)

				e = event_controller.RegisterHandler("m3", mt3)
				So(e, ShouldBeNil)

				e = event_controller.RegisterHandler("m4", mt4)
				So(e, ShouldBeNil)

				e = event_controller.RegisterHandler("m5", mt5)
				So(e, ShouldBeNil)
				So(event_controller.HandlerCount(), ShouldEqual, 5)

			})
		})

		Convey(".IsHandlerRegistered", func() {
			Convey("Returns false for Handler never registered", func() {
				event_controller := new(EventController)
				b := event_controller.IsHandlerRegistered("MyMock1")

				So(b, ShouldBeFalse)
			})
			Convey("Returns true for a registered Handler", func() {
				event_controller := new(EventController)
				mt1 := new(MockThing)

				event_controller.RegisterHandler("MyMock1", mt1)
				b := event_controller.IsHandlerRegistered("MyMock1")

				So(b, ShouldBeTrue)
			})
			Convey("Returns false for a registered Handler that was unregistered", func() {
				event_controller := new(EventController)
				mt1 := new(MockThing)
				event_controller.RegisterHandler("M1", mt1)
				event_controller.UnregisterHandler("M1")
				b := event_controller.IsHandlerRegistered("M1")

				So(b, ShouldBeFalse)
			})
		})

		Convey(".UnregsiterHandler", func() {
			Convey("Unregisters the Handler", func() {
				event_controller := new(EventController)
				mt1 := new(MockThing)
				event_controller.RegisterHandler("m1", mt1)
				event_controller.UnregisterHandler("m1")
				b := event_controller.IsHandlerRegistered("m1")

				So(b, ShouldBeFalse)
			})
		})

	})
}

func TestHandler(t *testing.T) {
	Convey("gomit.Handler", t, func() {
		Convey("Handler is called with correct event", func() {
			event_controller := new(EventController)
			mt := new(MockThing)

			event_controller.RegisterHandler("m1", mt)
			eb := new(MockEventBody)

			i, e := event_controller.Emit(eb)
			// We have to pause to let Handlers run.
			time.Sleep(time.Millisecond * 100)

			// One handler called
			So(i, ShouldEqual, 1)
			// One handler registered
			So(event_controller.HandlerCount(), ShouldEqual, 1)
			// MockThing should have Event namespace (handler was called)
			So(mt.LastNamespace, ShouldEqual, eb.Namespace())
			So(e, ShouldBeNil)
		})
		Convey("Should only emit to the first registered Handler", func() {
			event_controller := new(EventController)
			mt1 := new(MockThing)
			mt2 := new(MockThing)
			eb := new(MockEventBody)

			event_controller.RegisterHandler("m1", mt1)
			// Should return error signifying it was already registered.
			e := event_controller.RegisterHandler("m1", mt2)
			So(e, ShouldNotBeNil)

			i, e := event_controller.Emit(eb)
			// We have to pause to let Handlers run.
			time.Sleep(time.Millisecond * 100)

			// One handler called
			So(i, ShouldEqual, 1)
			So(event_controller.HandlerCount(), ShouldEqual, 1)
			// This was the first one registered and should match
			So(mt1.LastNamespace, ShouldEqual, eb.Namespace())
			// This was the second one (attempted) and should not match
			So(mt2.LastNamespace, ShouldNotEqual, eb.Namespace())
			So(e, ShouldBeNil)
		})
	})
}

func TestNewEventController(t *testing.T) {
	eh := NewEventController()
	Convey("returns a pointer", t, func() {
		So(eh, ShouldNotBeNil)
	})
	Convey("that pointer should point to a type EventController", t, func() {
		So(eh, ShouldHaveSameTypeAs, new(EventController))
	})
	Convey(".Handlers should not be nil", t, func() {
		So(eh.Handlers, ShouldNotBeNil)
	})
}
