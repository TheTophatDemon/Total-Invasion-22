package ui

import (
	"github.com/go-gl/mathgl/mgl32"
	"tophatdemon.com/total-invasion-ii/engine/assets/cache"
	"tophatdemon.com/total-invasion-ii/engine/assets/shaders"
	"tophatdemon.com/total-invasion-ii/engine/assets/textures"
	"tophatdemon.com/total-invasion-ii/engine/color"
	"tophatdemon.com/total-invasion-ii/engine/failure"
	"tophatdemon.com/total-invasion-ii/engine/math2"
	"tophatdemon.com/total-invasion-ii/engine/render"
	"tophatdemon.com/total-invasion-ii/engine/scene/comps"
)

type Box struct {
	Color       color.Color
	Texture     *textures.Texture
	AnimPlayer  comps.AnimationPlayer
	FlippedHorz bool
	Hidden      bool

	src, dest      math2.Rect
	depth          float32
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
		depth:          0.0,
		transformDirty: true,
	}
}

func (box *Box) InitDefault() {
	*box = Box{}
	box.Color = color.White
	box.Texture = nil
	box.AnimPlayer = comps.AnimationPlayer{}
	box.FlippedHorz = false
	box.src = math2.Rect{X: 0.0, Y: 0.0, Width: 1.0, Height: 1.0}
	box.dest = math2.Rect{}
	box.transformDirty = false
	box.transform = mgl32.Ident4()
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

func (box *Box) Depth() float32 {
	return box.depth
}

func (box *Box) SetDepth(value float32) *Box {
	box.depth = value
	box.transformDirty = true
	return box
}

func (box *Box) SetDestPosition(position mgl32.Vec2) *Box {
	box.dest = math2.Rect{
		X:      position.X(),
		Y:      position.Y(),
		Width:  box.dest.Width,
		Height: box.dest.Height,
	}
	box.transformDirty = true
	return box
}

func (box *Box) DestPosition() mgl32.Vec2 {
	return mgl32.Vec2{box.dest.X, box.dest.Y}
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
		box.transform = mgl32.Translate3D(bx, by, box.depth).Mul4(mgl32.Scale3D(scx, scy, 1.0))
	}
	return box.transform
}

func (box *Box) Render(context *render.Context) {
	if box.Hidden {
		return
	}
	failure.CheckOpenGLError()
	cache.QuadMesh.Bind()
	shaders.UIShader.Use()

	_ = context.SetUniforms(shaders.UIShader)
	_ = shaders.UIShader.SetUniformInt(shaders.UniformTex, 0)

	// Set color.
	_ = shaders.UIShader.SetUniformVec4(shaders.UniformDiffuseColor, box.Color.Vector())

	// Set texture.
	if box.Texture != nil {
		_ = shaders.UIShader.SetUniformBool(shaders.UniformNoTexture, false)
		box.Texture.Bind()
		texW, texH := float32(box.Texture.Width()), float32(box.Texture.Height())
		srcVec := mgl32.Vec4{
			box.src.X / texW,
			1.0 - (box.src.Y / texH),
			box.src.Width / texW,
			box.src.Height / texH,
		}
		_ = shaders.UIShader.SetUniformVec4(shaders.UniformSrcRect, srcVec)
	} else {
		_ = shaders.UIShader.SetUniformBool(shaders.UniformNoTexture, true)
	}

	_ = shaders.UIShader.SetUniformBool(shaders.UniformFlipHorz, box.FlippedHorz)

	// Set uniforms
	_ = shaders.UIShader.SetUniformMatrix(shaders.UniformModelMatrix, box.Transform())
	cache.QuadMesh.DrawAll()
	failure.CheckOpenGLError()
}
