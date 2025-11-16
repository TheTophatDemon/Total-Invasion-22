package ui

import "testing"

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
		if q[i].Depth != expectedDepth {
			t.Errorf("expected depth %v at slot %v but got %v", expectedDepth, i, q[i].Depth)
		}
	}
}
