package ui

import (
	"image/color"

	"github.com/go-gl/mathgl/mgl32"
	"tophatdemon.com/total-invasion-ii/engine/assets"
	"tophatdemon.com/total-invasion-ii/engine/math2"
	"tophatdemon.com/total-invasion-ii/engine/world/comps"
)

type Box struct {
	Color      color.Color
	Texture    *assets.Texture
	AnimPlayer comps.AnimationPlayer

	src, dest      math2.Rect
	transformDirty bool
	transform      mgl32.Mat4
}

func NewBox(src, dest math2.Rect, texture *assets.Texture, color color.Color) Box {
	return Box{
		Color:          color,
		Texture:        texture,
		src:            src,
		dest:           dest,
		transformDirty: true,
	}
}

func NewBoxFull(dest math2.Rect, texture *assets.Texture, color color.Color) Box {
	src := math2.Rect{}
	if texture != nil {
		src = math2.Rect{
			X:      0.0,
			Y:      0.0,
			Width:  float32(texture.Width()),
			Height: float32(texture.Height()),
		}
	}
	return Box{
		Color:          color,
		Texture:        texture,
		src:            src,
		dest:           dest,
		transformDirty: true,
	}
}

func (box *Box) Update(deltaTime float32) {
	box.AnimPlayer.Update(deltaTime)
}

func (box *Box) Dest() math2.Rect {
	return box.dest
}

func (box *Box) SetDest(dest math2.Rect) {
	box.dest = dest
	box.transformDirty = true
}

func (box *Box) Src() math2.Rect {
	return box.src
}

func (box *Box) SetSrc(src math2.Rect) {
	box.src = src
	box.transformDirty = true
}

func (box *Box) Transform() mgl32.Mat4 {
	if box.transformDirty {
		box.transformDirty = false
		bx, by := box.dest.Center()
		scx := box.dest.Width / 2.0
		scy := box.dest.Height / 2.0
		box.transform = mgl32.Translate3D(bx, by, 0.0).Mul4(mgl32.Scale3D(scx, scy, 1.0))
	}
	return box.transform
}
