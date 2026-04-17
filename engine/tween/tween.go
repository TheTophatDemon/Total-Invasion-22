package tween

type Func func(time, start, diff, duration float32) float32

func Linear(time, start, diff, duration float32) float32 {
	return start + (diff * (time / duration))
}

type Data struct {
	StartValue, EndValue, Duration float32
	Time                           float32
	Interpolation                  Func
}

func (tw *Data) Update(deltaTime float32) (value float32, done bool) {
	var fn Func
	if tw.Interpolation == nil {
		fn = Linear
	}
	if tw.Time >= tw.Duration {
		tw.Time = tw.Duration
		return tw.EndValue, true
	}
	value = fn(tw.Time, tw.StartValue, tw.EndValue-tw.StartValue, tw.Duration)
	tw.Time += deltaTime
	return
}

type Sequence struct {
	Tweens []Data
	Index  int
	Loop   bool
}

func (seq *Sequence) Update(deltaTime float32) (value float32, tweenDone, sequenceDone bool) {
	if seq == nil || len(seq.Tweens) == 0 {
		return 0, true, true
	}
	tween := &seq.Tweens[min(len(seq.Tweens)-1, seq.Index)]
	value, done := tween.Update(deltaTime)
	if done {
		sequenceDone = seq.Index == len(seq.Tweens)-1
		seq.Index++
		if seq.Index >= len(seq.Tweens) {
			if seq.Loop {
				seq.Index = 0
				for i := range seq.Tweens {
					seq.Tweens[i].Time = 0.0
				}
			} else {
				seq.Index = len(seq.Tweens) - 1
			}
		}
		return value, true, sequenceDone
	}
	return value, false, false
}
