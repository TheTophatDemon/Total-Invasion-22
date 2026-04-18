package tween

import (
	"testing"
)

func TestSequenceNonLooping(t *testing.T) {
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
	for timeStep, expected := range []SeqResult{
		{Value: 1.0, TweenDone: false, SequenceDone: false, SequenceIndex: 0},
		{Value: 1.5, TweenDone: false, SequenceDone: false, SequenceIndex: 0},
		{Value: 2.0, TweenDone: true, SequenceDone: false, SequenceIndex: 0},
		{Value: 2.0, TweenDone: false, SequenceDone: false, SequenceIndex: 1},
		{Value: 2.5, TweenDone: true, SequenceDone: false, SequenceIndex: 1},
		{Value: 0.0, TweenDone: false, SequenceDone: false, SequenceIndex: 2},
		{Value: 0.5, TweenDone: false, SequenceDone: false, SequenceIndex: 2},
		{Value: 1.0, TweenDone: false, SequenceDone: false, SequenceIndex: 2},
		{Value: 1.5, TweenDone: false, SequenceDone: false, SequenceIndex: 2},
		{Value: 2.0, TweenDone: false, SequenceDone: false, SequenceIndex: 2},
		{Value: 2.5, TweenDone: false, SequenceDone: false, SequenceIndex: 2},
		{Value: 3.0, TweenDone: false, SequenceDone: false, SequenceIndex: 2},
		{Value: 3.5, TweenDone: false, SequenceDone: false, SequenceIndex: 2},
		{Value: 4.0, TweenDone: false, SequenceDone: false, SequenceIndex: 2},
		{Value: 4.5, TweenDone: false, SequenceDone: false, SequenceIndex: 2},
		{Value: 5.0, TweenDone: true, SequenceDone: true, SequenceIndex: 2},
		{Value: 5.0, TweenDone: true, SequenceDone: true, SequenceIndex: 2},
	} {
		res := tweens.Update(0.5)
		tweenNumber := tweens.Index + 1
		if res != expected {
			t.Fatalf("expected value to be %+v at timestep %v in tween #%v, but was %+v", expected, timeStep, tweenNumber, res)
		}
	}
}

func TestSequenceLooping(t *testing.T) {
	tweens := Sequence{
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

	t.Log("Test with looping")
	for timeStep, expected := range []SeqResult{
		{Value: 1.0, TweenDone: false, SequenceDone: false, SequenceIndex: 0},
		{Value: 0.5, TweenDone: false, SequenceDone: false, SequenceIndex: 0},
		{Value: 0.0, TweenDone: true, SequenceDone: false, SequenceIndex: 0},
		{Value: 0.0, TweenDone: false, SequenceDone: false, SequenceIndex: 1},
		{Value: 0.5, TweenDone: false, SequenceDone: false, SequenceIndex: 1},
		{Value: 1.0, TweenDone: true, SequenceDone: true, SequenceIndex: 1},
		{Value: 1.0, TweenDone: false, SequenceDone: false, SequenceIndex: 0},
		{Value: 0.5, TweenDone: false, SequenceDone: false, SequenceIndex: 0},
		{Value: 0.0, TweenDone: true, SequenceDone: false, SequenceIndex: 0},
		{Value: 0.0, TweenDone: false, SequenceDone: false, SequenceIndex: 1},
		{Value: 0.5, TweenDone: false, SequenceDone: false, SequenceIndex: 1},
		{Value: 1.0, TweenDone: true, SequenceDone: true, SequenceIndex: 1},
	} {
		res := tweens.Update(0.5)
		tweenNumber := tweens.Index + 1
		if res != expected {
			t.Fatalf("expected value to be %+v at timestep %v in tween #%v, but was %+v", expected, timeStep, tweenNumber, res)
		}
	}
}

func TestTweenInfer(t *testing.T) {
	t.Log("infer start")
	tween := Data{
		StartValue: Infer,
		EndValue:   1.0,
		Duration:   1.0,
	}
	res := tween.Update(0.5)
	expected := Result{Value: 1.0, Done: false}
	if res != expected {
		t.Fatalf("expected result %+v but got %+v", expected, res)
	}
	if tween.StartValue != 1.0 {
		t.Fatalf("tween start should be 1.0")
	}

	t.Log("infer end")
	tween = Data{
		StartValue: 2.0,
		EndValue:   Infer,
		Duration:   1.0,
	}
	res = tween.Update(0.5)
	expected = Result{Value: 2.0, Done: false}
	if res != expected {
		t.Fatalf("expected result %+v but got %+v", expected, res)
	}
	if tween.EndValue != 2.0 {
		t.Fatalf("tween end should be 2.0")
	}

	t.Log("infer both")
	tween = Data{
		StartValue: Infer,
		EndValue:   Infer,
		Duration:   1.0,
	}
	res = tween.Update(0.5)
	expected = Result{Value: 0.0, Done: false}
	if res != expected {
		t.Fatalf("expected result %+v but got %+v", expected, res)
	}
	if tween.StartValue != 0.0 {
		t.Fatalf("tween start should be 0.0")
	}
	if tween.EndValue != 0.0 {
		t.Fatalf("tween end should be 0.0")
	}
}

func TestSequenceInfer(t *testing.T) {
	tweens := Sequence{Tweens: []Data{
		{
			StartValue: Infer,
			EndValue:   Infer,
			Duration:   1.0,
		},
		{
			StartValue: Infer,
			EndValue:   2.5,
			Duration:   0.5,
		},
	}}
	for timeStep, expected := range []SeqResult{
		{Value: 0.0, TweenDone: false, SequenceDone: false, SequenceIndex: 0},
		{Value: 0.0, TweenDone: false, SequenceDone: false, SequenceIndex: 0},
		{Value: 0.0, TweenDone: true, SequenceDone: false, SequenceIndex: 0},
		{Value: 0.0, TweenDone: false, SequenceDone: false, SequenceIndex: 1},
		{Value: 2.5, TweenDone: true, SequenceDone: true, SequenceIndex: 1},
	} {
		res := tweens.Update(0.5)
		tweenNumber := tweens.Index + 1
		switch res.SequenceIndex {
		case 0:
			if tweens.Tweens[0].StartValue != 0.0 {
				t.Fatalf("tween #1's start value should be 0, but it's %v", tweens.Tweens[0].StartValue)
			}
		case 1:
			if tweens.Tweens[1].StartValue != 0.0 {
				t.Fatalf("tween #2's start value should be 0.0, but it's %v", tweens.Tweens[1].StartValue)
			}
		}
		if res != expected {
			t.Fatalf("expected value to be %+v at timestep %v in tween #%v, but was %+v", expected, timeStep, tweenNumber, res)
		}
	}
}
