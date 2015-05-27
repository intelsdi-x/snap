package gomit

import (
	"time"
)

// Represents an event emitted by an Emitter and handled by a Handler
type Event struct {
	Header EventHeader
	Body   EventBody
}

// Represents the compatible event body provided by the producer.
type EventBody interface {
	Namespace() string
}

// Contains common data across all Event instances.
type EventHeader struct {
	Time time.Time
}

// Provides a string Namespace for the Event. Namespace is open to implementation
// with the producers/consumers for routing.
func (e *Event) Namespace() string {
	return e.Body.Namespace()
}

// Generates an EventHeader.
func generateHeader() EventHeader {
	return EventHeader{Time: time.Now()}
}
