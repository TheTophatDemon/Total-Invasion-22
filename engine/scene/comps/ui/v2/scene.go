package ui

import (
	"github.com/go-gl/gl/v3.3-core/gl"
	"tophatdemon.com/total-invasion-ii/engine/color"
	"tophatdemon.com/total-invasion-ii/engine/failure"
	"tophatdemon.com/total-invasion-ii/engine/math2"
	"tophatdemon.com/total-invasion-ii/engine/render"
)

type RenderQueue struct {
	scissor  math2.Rect
	elements []*Element
}

var textShadowColor color.Color

func (renderQ *RenderQueue) Add(elems ...*Element) {
	for _, elem := range elems {
		if elem == nil {
			continue
		}
		if renderQ.scissor.Width > 0 && renderQ.scissor.Height > 0 {
			elem.Scissor = renderQ.scissor
		}
		newQ := append(renderQ.elements, elem)
		var displacedElem *Element

		for i := range newQ {
			if newQ[i].Depth() > elem.Depth() || i == len(newQ)-1 {
				if displacedElem == nil {
					displacedElem = newQ[i]
					newQ[i] = elem
				} else {
					newQ[i], displacedElem = displacedElem, newQ[i]
				}
			}
		}
		renderQ.elements = newQ
	}
}

func (renderQ *RenderQueue) Render(context *render.Context) {
	failure.CheckOpenGLError()
	gl.CullFace(gl.FRONT)
	gl.Disable(gl.DEPTH_TEST)
	gl.Disable(gl.SCISSOR_TEST)
	defer gl.Enable(gl.DEPTH_TEST)
	defer failure.CheckOpenGLError()

	for _, elem := range renderQ.elements {
		elem.Render(context)
	}
	renderQ.Clear()
}

func (renderQ *RenderQueue) Clear() {
	renderQ.elements = renderQ.elements[0:0]
}

func (renderQ *RenderQueue) SetScissor(x, y, width, height float32) {
	renderQ.scissor = math2.Rect{X: x, Y: y, Width: width, Height: height}
}

func (renderQ *RenderQueue) ClearScissor() {
	renderQ.scissor = math2.Rect{}
}

func SetTextShadowColor(clr color.Color) {
	textShadowColor = clr
}
