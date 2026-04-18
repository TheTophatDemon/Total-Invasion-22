package tween

import (
	"math"

	"tophatdemon.com/total-invasion-ii/engine/math2"
)

type Func func(time, start, diff, duration float32) float32

func Linear(time, start, diff, duration float32) float32 {
	return start + (diff * (time / duration))
}

func CubicIn(time, start, diff, duration float32) float32 {
	return (diff * math2.Pow(time/duration, 3.0)) + start
}

func CubicOut(time, start, diff, duration float32) float32 {
	return (diff * (math2.Pow((time/duration)-1.0, 3.0) + 1)) + start
}

type (
	// Represents how a value is animated from one moment in time to the next.
	Data struct {
		// The starting value. If this is set to Infer, then it will use the previous value in a Sequence if available.
		StartValue float32
		// The destination value. If this is set to Infer, then it will use the starting value instead.
		EndValue float32
		// Duration of the tween in seconds
		Duration float32
		// Amount of time that has elapsed in the currently playing tween.
		Elapsed float32
		// The function used to determine the value based off of the time. Will use Linear by default.
		Interpolation Func
	}
	Result struct {
		Value float32
		Done  bool
	}
	SeqResult struct {
		Value                   float32
		TweenDone, SequenceDone bool
		SequenceIndex           int
	}
)

// This value marks a start or end value with NaN so it will be assigned a previous value.
var Infer = float32(math.NaN())

func (tw *Data) Update(deltaTime float32) Result {
	if math2.IsNan(tw.StartValue) {
		if math2.IsNan(tw.EndValue) {
			// At this point there is nothing to be inferred.
			tw.StartValue = 0
		} else {
			tw.StartValue = tw.EndValue
		}
	}
	if math2.IsNan(tw.EndValue) {
		tw.EndValue = tw.StartValue
	}
	var fn Func = tw.Interpolation
	if tw.Interpolation == nil {
		fn = Linear
	}
	if tw.Elapsed >= tw.Duration {
		tw.Elapsed = tw.Duration
		return Result{Value: tw.EndValue, Done: true}
	}
	value := fn(tw.Elapsed, tw.StartValue, tw.EndValue-tw.StartValue, tw.Duration)
	tw.Elapsed += deltaTime
	return Result{
		Value: value,
		Done:  false,
	}
}

type Sequence struct {
	Tweens []Data
	Index  int
	Loop   bool
}

func (seq *Sequence) Update(deltaTime float32) SeqResult {
	if seq == nil || len(seq.Tweens) == 0 {
		return SeqResult{}
	}
	tween := &seq.Tweens[min(len(seq.Tweens)-1, seq.Index)]
	if math2.IsNan(tween.StartValue) && seq.Index > 0 {
		// Set the start value to the previous tween's end value
		tween.StartValue = seq.Tweens[seq.Index-1].EndValue
	}
	tweenResult := tween.Update(deltaTime)
	if tweenResult.Done {
		res := SeqResult{
			Value:         tweenResult.Value,
			TweenDone:     true,
			SequenceDone:  seq.Index == len(seq.Tweens)-1,
			SequenceIndex: seq.Index,
		}
		seq.Index++
		if seq.Index >= len(seq.Tweens) {
			if seq.Loop {
				seq.Index = 0
				for i := range seq.Tweens {
					seq.Tweens[i].Elapsed = 0.0
				}
			} else {
				seq.Index = len(seq.Tweens) - 1
			}
		}
		return res
	}
	return SeqResult{
		Value:         tweenResult.Value,
		TweenDone:     false,
		SequenceDone:  false,
		SequenceIndex: seq.Index,
	}
}
