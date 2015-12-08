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

package chrono

import (
	"time"
)

// The chrono structure provides an artificial notion of time that can be
// stopped, forwarded to make testing of timed code deterministic.
type chrono struct {
	skew     time.Duration
	paused   bool
	pausedAt time.Time
}

// Now() should replace usage of time.Now(). If chrono has been paused, Now()
// returns the time when it was paused. If there is any skew (due to forwarding
// or reversing), this is always added to the end time.
func (c *chrono) Now() time.Time {
	var now time.Time
	if c.paused {
		now = c.pausedAt
	} else {
		now = time.Now()
	}
	return now.Add(c.skew)
}

// Forwards time of chrono with skew time. This can be used in both running and
// paused mode.
func (c *chrono) Forward(skew time.Duration) {
	c.skew = skew
}

// Resets any previous set clock skew.
func (c *chrono) Reset() {
	c.skew = 0
}

// Pause "Stops" time by recording current time and shortcircuit Now() to return this
// time instead of the actual time (plus skew).
func (c *chrono) Pause() {
	c.pausedAt = c.Now()
	c.paused = true
}

// Continues time after having been paused. This has no effect if clock is
// already running.
func (c *chrono) Continue() {
	c.paused = false
}

// Chrono variable
var Chrono chrono
