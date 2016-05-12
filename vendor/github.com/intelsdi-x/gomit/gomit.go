/*
http://www.apache.org/licenses/LICENSE-2.0.txt


Copyright 2015 Intel Corporation

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

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
