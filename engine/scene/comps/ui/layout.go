package ui

import (
	"github.com/go-gl/mathgl/mgl32"
)

type (
	StackParams struct {
		Vertical bool
		Gap      float32
	}
)

// Places elements consecutively in a line relative to a given parent element.
func LayoutStack(parent *Element, params StackParams, elements ...*Element) {
	var origin mgl32.Vec2
	if parent != nil {
		origin = parent.Position()
	}
	axis := 0
	otherAxis := 1
	if params.Vertical {
		axis = 1
		otherAxis = 0
	}
	nextPos := origin
	for _, elem := range elements {
		if elem == nil {
			continue
		}
		elem.SetPosition(nextPos)
		sizeOnAxis := elem.Size()[axis] + params.Gap
		nextPos[axis] += sizeOnAxis
	}
	if parent != nil {
		// Resize parent to fit children
		containerSize := mgl32.Vec2{}
		containerSize[axis] = nextPos[axis] - origin[axis] - params.Gap
		containerSize[otherAxis] = parent.Size()[otherAxis]
		parent.SetSize(containerSize)
	}
}
