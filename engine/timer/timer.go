package timer

import "tophatdemon.com/total-invasion-ii/engine/math2"

type Timer struct {
	Elapsed  float32     // Number of seconds elapsed since last tick. Can be modified directly to make intervals temporarily longer or shorter.
	Interval float32     // Number of seconds between each tick
	NumTicks int         // How many ticks have transpired since this timer was created / last reset
	MaxTicks int         // Maximum tick count. Once this number is exceeded and it is non-zero, no more ticks occur.
	Callback func(Timer) // Optional function to call each time a tick occurs.
}

// Updates the timer. Returns true if the timer ticked this frame.
func (tmr *Timer) Update(deltaTime float32) bool {
	if tmr == nil || tmr.Interval == 0.0 {
		return false
	}
	tmr.Elapsed += deltaTime
	if tmr.Elapsed > tmr.Interval {
		tmr.Elapsed = math2.Mod(tmr.Elapsed, tmr.Interval)
		tmr.NumTicks++
		if tmr.MaxTicks > 0 && tmr.NumTicks > tmr.MaxTicks {
			tmr.NumTicks = tmr.MaxTicks
			return false
		} else if tmr.Callback != nil {
			tmr.Callback(*tmr)
		}
		return true
	}
	return false
}

func (tmr *Timer) Reset() {
	tmr.Elapsed = 0.0
	tmr.NumTicks = 0
}
