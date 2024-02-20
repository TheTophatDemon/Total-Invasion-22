package ui

import (
	"github.com/go-gl/mathgl/mgl32"
	"tophatdemon.com/total-invasion-ii/engine/assets/textures"
	"tophatdemon.com/total-invasion-ii/engine/color"
	"tophatdemon.com/total-invasion-ii/engine/math2"
	"tophatdemon.com/total-invasion-ii/engine/scene/comps"
)

type Box struct {
	Color       color.Color
	Texture     *textures.Texture
	AnimPlayer  comps.AnimationPlayer
	FlippedHorz bool

	src, dest      math2.Rect
	transformDirty bool
	transform      mgl32.Mat4
}

func NewBox(src, dest math2.Rect, texture *textures.Texture, color color.Color) Box {
	return Box{
		Color:          color,
		Texture:        texture,
		src:            src,
		dest:           dest,
		transformDirty: true,
	}
}

func NewBoxFull(dest math2.Rect, texture *textures.Texture, color color.Color) Box {
	src := math2.Rect{}
	if texture != nil {
		src = math2.Rect{
			X:      0.0,
			Y:      1.0,
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
	if frame := box.AnimPlayer.Frame(); frame.Duration > 0.0 {
		box.src = frame.Rect
	}
}

func (box *Box) Dest() math2.Rect {
	return box.dest
}

func (box *Box) SetDest(dest math2.Rect) *Box {
	box.dest = dest
	box.transformDirty = true
	return box
}

func (box *Box) Src() math2.Rect {
	return box.src
}

func (box *Box) SetSrc(src math2.Rect) *Box {
	box.src = src
	box.transformDirty = true
	return box
}

func (box *Box) SetColor(clr color.Color) *Box {
	box.Color = clr
	return box
}

func (box *Box) SetTexture(t *textures.Texture) *Box {
	box.Texture = t
	return box
}

func (box *Box) SetFlipHorz(f bool) *Box {
	box.FlippedHorz = f
	return box
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
