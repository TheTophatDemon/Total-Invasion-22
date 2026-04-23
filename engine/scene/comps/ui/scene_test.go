package ui

import (
	"testing"

	"tophatdemon.com/total-invasion-ii/engine/math2"
)

func TestElementSorting(t *testing.T) {
	var q RenderQueue

	boxes := [...]Element{
		NewBox(Transform{Depth: 4.0}, nil),
		NewBox(Transform{Depth: 3.0}, nil),
		NewBox(Transform{Depth: 1.0}, nil),
		NewBox(Transform{Depth: 2.0}, nil),
		NewBox(Transform{Depth: 3.0}, nil),
	}
	for i := range boxes {
		q.Add(&boxes[i])
	}

	for i, expectedDepth := range []float32{1.0, 2.0, 3.0, 3.0, 4.0} {
		if actualDepth := q.elements[i].Depth(); actualDepth != expectedDepth {
			t.Errorf("expected depth %v at slot %v but got %v", expectedDepth, i, actualDepth)
		}
	}
}

func TestElementScissor(t *testing.T) {
	var q RenderQueue

	for i := range 6 {
		switch i {
		case 2:
			q.SetScissor(4, 4, 50, 50)
		case 4:
			q.ClearScissor()
		}
		q.Add(new(NewBox(Transform{}, nil)))
	}

	for i, elem := range q.elements {
		if i >= 2 && i < 4 {
			if elem.Scissor != (math2.Rect{X: 4.0, Y: 4.0, Width: 50.0, Height: 50.0}) {
				t.Fatalf("element %v did not have matching scissor, it was: %v", i, elem.Scissor)
			}
		} else {
			if elem.Scissor != (math2.Rect{}) {
				t.Fatalf("element %v has a scissor (%v) and should not", i, elem.Scissor)
			}
		}
	}
}
