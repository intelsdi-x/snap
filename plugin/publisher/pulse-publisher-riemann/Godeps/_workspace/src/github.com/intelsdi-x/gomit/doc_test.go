package gomit

import (
	"fmt"
	"time"
)

type Widget struct {
	EventCount int
}

func (w *Widget) HandleGomitEvent(e Event) {
	w.EventCount++
}

type RandomEventBody struct {
}

func (r *RandomEventBody) Namespace() string {
	return "random.event"
}

func Example() {
	event_controller := new(EventController)
	/*
		type Widget struct {
			EventCount int
		}

		func (w *Widget) HandleGomitEvent(e Event) {
			w.EventCount++
		}
	*/
	widget := new(Widget)

	event_controller.RegisterHandler("widget1", widget)

	event_controller.Emit(new(RandomEventBody))
	event_controller.Emit(new(RandomEventBody))
	event_controller.Emit(new(RandomEventBody))

	time.Sleep(time.Millisecond * 100)
	fmt.Println(widget.EventCount)
	// Output: 3
}

// Empty but makes the example not print the whole file
func ExampleFoo() {
}
