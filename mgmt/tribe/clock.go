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
