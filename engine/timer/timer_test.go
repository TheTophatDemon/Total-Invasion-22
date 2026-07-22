package timer

import "testing"

// When the timer is zero-initialized, it should never tick when it is updated.
func TestTimerZeroValue(t *testing.T) {
	timer := Timer{}
	for range 10 {
		if timer.Update(0.1) {
			t.Errorf("timer update should not return true")
		}
	}
	if timer.Elapsed != 0 || timer.Interval != 0 || timer.MaxTicks != 0 || timer.NumTicks != 0 {
		t.Errorf("timer state was modified during update: %v", timer)
	}
}
