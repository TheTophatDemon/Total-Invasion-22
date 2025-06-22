package ui

import "testing"

func TestElementSorting(t *testing.T) {
	var q RenderQueue

	q.Add(&Box{Transform: Transform{Depth: 4.0}},
		&Box{Transform: Transform{Depth: 3.0}},
		&Box{Transform: Transform{Depth: 1.0}},
		&Box{Transform: Transform{Depth: 2.0}},
		&Box{Transform: Transform{Depth: 3.0}})

	for i, expectedDepth := range []float32{1.0, 2.0, 3.0, 3.0, 4.0} {
		if q[i].DrawDepth() != expectedDepth {
			t.Errorf("expected depth %v at slot %v but got %v", expectedDepth, i, q[i].DrawDepth())
		}
	}
}
