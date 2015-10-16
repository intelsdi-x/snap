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

package tribe

import "sync/atomic"

type LClock struct {
	Inc uint64
}

type LTime uint64

//Update
func (l *LClock) Update(lt LTime) {
	for {
		cur := LTime(atomic.LoadUint64(&l.Inc))
		// If we are newer return now
		if lt < cur {
			return
		}
		// If we CAS successfully return else our counter changed
		// and we need to try again
		if atomic.CompareAndSwapUint64(&l.Inc, uint64(cur), uint64(lt)+1) {
			return
		}
	}
}

// Increment
func (l *LClock) Increment() LTime {
	return LTime(atomic.AddUint64(&l.Inc, 1))
}

// Time returns the current value of the clock
func (l *LClock) Time() LTime {
	return LTime(atomic.LoadUint64(&l.Inc))
}
