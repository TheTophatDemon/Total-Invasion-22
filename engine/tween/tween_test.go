package tween

import (
	"testing"
)

func TestTweenSequence(t *testing.T) {
	tweens := Sequence{Tweens: []Data{
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
	}}
	for timeStep, expected := range [](struct {
		value                   float32
		tweenDone, sequenceDone bool
	}){
		{value: 1.0, tweenDone: false, sequenceDone: false},
		{value: 1.5, tweenDone: false, sequenceDone: false},
		{value: 2.0, tweenDone: true, sequenceDone: false},
		{value: 2.0, tweenDone: false, sequenceDone: false},
		{value: 2.5, tweenDone: true, sequenceDone: false},
		{value: 0.0, tweenDone: false, sequenceDone: false},
		{value: 0.5, tweenDone: false, sequenceDone: false},
		{value: 1.0, tweenDone: false, sequenceDone: false},
		{value: 1.5, tweenDone: false, sequenceDone: false},
		{value: 2.0, tweenDone: false, sequenceDone: false},
		{value: 2.5, tweenDone: false, sequenceDone: false},
		{value: 3.0, tweenDone: false, sequenceDone: false},
		{value: 3.5, tweenDone: false, sequenceDone: false},
		{value: 4.0, tweenDone: false, sequenceDone: false},
		{value: 4.5, tweenDone: false, sequenceDone: false},
		{value: 5.0, tweenDone: true, sequenceDone: true},
		{value: 5.0, tweenDone: true, sequenceDone: true},
	} {
		value, tweenDone, sequenceDone := tweens.Update(0.5)
		tweenNumber := tweens.Index + 1
		if value != expected.value {
			t.Fatalf("expected value to be %v at timestep %v in tween #%v, but was %v", expected.value, timeStep, tweenNumber, value)
		}
		if tweenDone != expected.tweenDone {
			t.Fatalf("expected tween #%v done=%t at timestep %v", tweenNumber, expected.tweenDone, timeStep)
		}
		if sequenceDone != expected.sequenceDone {
			t.Fatalf("expected sequence done=%t at tween #%v timestep %v", expected.sequenceDone, tweenNumber, timeStep)
		}
	}

	// Test with looping
	tweens = Sequence{
		Tweens: []Data{
			{
				StartValue: 1.0,
				EndValue:   0.0,
				Duration:   1.0,
			},
			{
				StartValue: 0.0,
				EndValue:   1.0,
				Duration:   1.0,
			},
		},
		Loop: true,
	}

	for timeStep, expected := range [](struct {
		value                   float32
		tweenDone, sequenceDone bool
	}){
		{value: 1.0, tweenDone: false, sequenceDone: false},
		{value: 0.5, tweenDone: false, sequenceDone: false},
		{value: 0.0, tweenDone: true, sequenceDone: false},
		{value: 0.0, tweenDone: false, sequenceDone: false},
		{value: 0.5, tweenDone: false, sequenceDone: false},
		{value: 1.0, tweenDone: true, sequenceDone: true},
		{value: 1.0, tweenDone: false, sequenceDone: false},
		{value: 0.5, tweenDone: false, sequenceDone: false},
		{value: 0.0, tweenDone: true, sequenceDone: false},
		{value: 0.0, tweenDone: false, sequenceDone: false},
		{value: 0.5, tweenDone: false, sequenceDone: false},
		{value: 1.0, tweenDone: true, sequenceDone: true},
	} {
		value, tweenDone, sequenceDone := tweens.Update(0.5)
		tweenNumber := tweens.Index + 1
		if value != expected.value {
			t.Fatalf("expected value to be %v at timestep %v in tween #%v, but was %v", expected.value, timeStep, tweenNumber, value)
		}
		if tweenDone != expected.tweenDone {
			t.Fatalf("expected tween #%v done=%t at timestep %v", tweenNumber, expected.tweenDone, timeStep)
		}
		if sequenceDone != expected.sequenceDone {
			t.Fatalf("expected sequence done=%t at tween #%v timestep %v", expected.sequenceDone, tweenNumber, timeStep)
		}
	}
}
