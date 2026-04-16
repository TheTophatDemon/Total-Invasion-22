package timer

type TweenFunc func(time, start, diff, duration float32) float32

func Linear(time, start, diff, duration float32) float32 {
	return start + (diff * (time / duration))
}

type Tween struct {
	StartValue, EndValue, Duration float32
	Time                           float32
	Interpolation                  TweenFunc
}

func (tw *Tween) Update(deltaTime float32) float32 {
	tw.Time += deltaTime
	if tw.Time > tw.Duration {
		tw.Time = tw.Duration
		return tw.EndValue
	}
	var fn TweenFunc
	if tw.Interpolation == nil {
		fn = Linear
	}
	return fn(tw.Time, tw.StartValue, tw.EndValue-tw.StartValue, tw.Duration)
}

type Tweens []Tween

func (tws *Tweens) Update(deltaTime float32) float32 {
	if tws == nil || len(*tws) == 0 {
		return 0
	}
	pointerToFirst := &(*tws)[0]
	value := pointerToFirst.Update(deltaTime)
	if pointerToFirst.Time == pointerToFirst.Duration && len(*tws) > 1 {
		// Move to next tween in the sequence
		*tws = (*tws)[1:]
	}
	return value
}
