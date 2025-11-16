package ui

import (
	"github.com/go-gl/gl/v3.3-core/gl"
	"tophatdemon.com/total-invasion-ii/engine/failure"
	"tophatdemon.com/total-invasion-ii/engine/render"
)

type RenderQueue []*Element

func (renderQ *RenderQueue) Add(elems ...*Element) {
	for _, elem := range elems {
		newQ := append(*renderQ, elem)
		var displacedElem *Element

		for i := range newQ {
			if newQ[i].Depth > elem.Depth || i == len(newQ)-1 {
				if displacedElem == nil {
					displacedElem = newQ[i]
					newQ[i] = elem
				} else {
					newQ[i], displacedElem = displacedElem, newQ[i]
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
