package timer

import "testing"

func TestTweenSequence(t *testing.T) {
	tweenSource := [4]Tween{
		{
			StartValue: 1.0,
			EndValue:   2.0,
			Duration:   1.0,
		},
		{
			StartValue: 2.0,
			EndValue:   2.5,
			Duration:   0.5,
		},
		{
			StartValue: 0.0,
			EndValue:   5.0,
			Duration:   5.0,
		},
	}
	tweens := Tweens(tweenSource[:])
	for _, expectedValue := range []float32{
		1.5,
		2.0,
		2.5,
		0.5,
		1.0,
		1.5,
		2.0,
		2.5,
		3.0,
		3.5,
		4.0,
		4.5,
		5.0,
	} {
		value := tweens.Update(0.5)
		currentTween := tweens[0]
		if value != expectedValue {
			t.Fatalf("expected value to be %v at time %v in tween #%v, but was %v", expectedValue, currentTween.Time, len(tweenSource)-len(tweens), value)
		}
	}
}
