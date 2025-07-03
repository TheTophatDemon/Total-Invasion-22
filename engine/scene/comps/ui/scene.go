package ui

import (
	"github.com/go-gl/gl/v3.3-core/gl"
	"tophatdemon.com/total-invasion-ii/engine/failure"
	"tophatdemon.com/total-invasion-ii/engine/math2"
	"tophatdemon.com/total-invasion-ii/engine/render"
)

type (
	Transform struct {
		Dest         math2.Rect
		Depth, Shear float32
		Scale        float32 // A multiplier for the item's size around its center. If it is 0, then a default value of 1 will be used.
		Rotation     float32 // Rotation angle in radians
	}

	RenderQueue []Element

	// Represents a UI element
	Element interface {
		DrawDepth() float32 // Used for sorting
		Render(*render.Context)
	}
)

func (renderQ *RenderQueue) Add(elems ...Element) {
	for _, elem := range elems {
		newQ := append(*renderQ, elem)
		var displacedElem Element

		for i := range newQ {
			if newQ[i].DrawDepth() > elem.DrawDepth() || i == len(newQ)-1 {
				if displacedElem == nil {
					displacedElem = newQ[i]
					newQ[i] = elem
				} else {
					temp := newQ[i]
					newQ[i] = displacedElem
					displacedElem = temp
				}
			}
		}
		*renderQ = newQ
	}
}

func (renderQ *RenderQueue) Render(context *render.Context) {
	failure.CheckOpenGLError()
	gl.CullFace(gl.FRONT)
	gl.Disable(gl.DEPTH_TEST)
	defer gl.Enable(gl.DEPTH_TEST)
	defer failure.CheckOpenGLError()

	for _, elem := range *renderQ {
		elem.Render(context)
	}
	// Clear the queue for the next render
	*renderQ = (*renderQ)[0:0]
}
