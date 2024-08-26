package ui

import (
	"slices"

	"github.com/go-gl/gl/v3.3-core/gl"
	"tophatdemon.com/total-invasion-ii/engine/failure"
	"tophatdemon.com/total-invasion-ii/engine/math2"
	"tophatdemon.com/total-invasion-ii/engine/render"
	"tophatdemon.com/total-invasion-ii/engine/scene"
)

type (
	Scene struct {
		Boxes          scene.Storage[Box]
		Texts          scene.Storage[Text]
		sortedElements []scene.Handle
	}

	Element interface {
		Depth() float32
		Render(*render.Context)
	}
)

func NewUIScene(maxBoxes, maxTexts uint) *Scene {
	return &Scene{
		Boxes:          scene.NewStorageWithFuncs(maxBoxes, (*Box).Update, nil),
		Texts:          scene.NewStorage[Text](maxTexts),
		sortedElements: make([]scene.Handle, 0, maxBoxes+maxTexts),
	}
}

func (scn *Scene) Update(deltaTime float32) {
	scn.Boxes.Update(deltaTime)

	// Gather the UI elements and sort them by depth.
	// This is done every frame to ensure that transparent objects display correctly.
	scn.sortedElements = scn.sortedElements[0:0]

	for boxHandle := range scn.Boxes.All() {
		scn.sortedElements = append(scn.sortedElements, boxHandle)
	}

	for textHandle := range scn.Texts.All() {
		scn.sortedElements = append(scn.sortedElements, textHandle)
	}

	slices.SortFunc(scn.sortedElements, func(a, b scene.Handle) int {
		elemA, okA := scene.Get[Element](a)
		if !okA {
			return -1
		}
		elemB, okB := scene.Get[Element](b)
		if !okB {
			return 1
		}
		return int(math2.Signum(elemA.Depth() - elemB.Depth()))
	})
}

func (scn *Scene) Render(context *render.Context) {
	failure.CheckOpenGLError()
	gl.CullFace(gl.FRONT)
	gl.Disable(gl.DEPTH_TEST)

	for _, handle := range scn.sortedElements {
		elem, ok := scene.Get[Element](handle)
		if !ok {
			continue
		}
		elem.Render(context)
	}

	gl.Enable(gl.DEPTH_TEST)
	failure.CheckOpenGLError()
}
