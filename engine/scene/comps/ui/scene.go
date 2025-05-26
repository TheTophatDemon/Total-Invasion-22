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
		sortedElements []sortedElement
	}

	Transform struct {
		Dest         math2.Rect
		Depth, Shear float32
		Scale        float32 // A multiplier for the item's size around its center. If it is 0, then a default value of 1 will be used.
		Rotation     float32 // Rotation angle in radians
	}

	sortedElement struct {
		scene.Handle
		depth float32
	}
)

func NewUIScene(maxBoxes, maxTexts uint) *Scene {
	return &Scene{
		Boxes:          scene.NewStorageWithFuncs(maxBoxes, (*Box).Update, nil),
		Texts:          scene.NewStorage[Text](maxTexts),
		sortedElements: make([]sortedElement, 0, maxBoxes+maxTexts),
	}
}

func (scn *Scene) Update(deltaTime float32) {
	scn.Boxes.Update(deltaTime)

	// Gather the UI elements and sort them by depth.
	// This is done every frame to ensure that transparent objects display correctly.
	scn.sortedElements = scn.sortedElements[0:0]

	boxIter := scn.Boxes.Iter()
	for box, boxHandle := boxIter.Next(); box != nil; box, boxHandle = boxIter.Next() {
		scn.sortedElements = append(scn.sortedElements, sortedElement{Handle: boxHandle, depth: box.Depth})
	}

	textIter := scn.Texts.Iter()
	for text, textHandle := textIter.Next(); text != nil; text, textHandle = textIter.Next() {
		scn.sortedElements = append(scn.sortedElements, sortedElement{Handle: textHandle, depth: text.Depth})
	}

	slices.SortFunc(scn.sortedElements, func(a, b sortedElement) int {
		if !a.Exists() {
			return -1
		}
		if !b.Exists() {
			return 1
		}
		return int(math2.Signum(a.depth - b.depth))
	})
}

func (scn *Scene) Render(context *render.Context) {
	failure.CheckOpenGLError()
	gl.CullFace(gl.FRONT)
	gl.Disable(gl.DEPTH_TEST)

	type renderable interface {
		Render(*render.Context)
	}
	for _, elem := range scn.sortedElements {
		item, ok := scene.Get[renderable](elem.Handle)
		if !ok {
			continue
		}
		item.Render(context)
	}

	gl.Enable(gl.DEPTH_TEST)
	failure.CheckOpenGLError()
}
