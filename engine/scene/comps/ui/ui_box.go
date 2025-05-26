package ui

import (
	"github.com/go-gl/mathgl/mgl32"
	"tophatdemon.com/total-invasion-ii/engine/assets/cache"
	"tophatdemon.com/total-invasion-ii/engine/assets/geom"
	"tophatdemon.com/total-invasion-ii/engine/assets/shaders"
	"tophatdemon.com/total-invasion-ii/engine/assets/textures"
	"tophatdemon.com/total-invasion-ii/engine/color"
	"tophatdemon.com/total-invasion-ii/engine/failure"
	"tophatdemon.com/total-invasion-ii/engine/math2"
	"tophatdemon.com/total-invasion-ii/engine/render"
	"tophatdemon.com/total-invasion-ii/engine/scene/comps"
)

type Box struct {
	Transform
	Src        math2.Rect
	Color      color.Color
	Texture    *textures.Texture
	AnimPlayer comps.AnimationPlayer
	Mesh       *geom.Mesh // The mesh to use. If nil, will use the default quad mesh.

	FlippedHorz  bool
	Hidden       bool
	oldTransform Transform
	matrix       mgl32.Mat4
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
		Color:   color,
		Texture: texture,
		Src:     src,
		Transform: Transform{
			Dest: dest,
		},
	}
}

func (box *Box) InitDefault() {
	*box = Box{}
	box.Color = color.White
	box.Texture = nil
	box.AnimPlayer = comps.AnimationPlayer{}
	box.FlippedHorz = false
	box.Src = math2.Rect{X: 0.0, Y: 0.0, Width: 1.0, Height: 1.0}
	box.matrix = mgl32.Ident4()
}

func (box *Box) Update(deltaTime float32) {
	box.AnimPlayer.Update(deltaTime)
	if frame := box.AnimPlayer.Frame(); frame.Duration > 0.0 {
		box.Src = frame.Rect
	}
}

func (box *Box) SetDestPosition(position mgl32.Vec2) *Box {
	box.Dest = math2.Rect{
		X:      position.X(),
		Y:      position.Y(),
		Width:  box.Dest.Width,
		Height: box.Dest.Height,
	}
	return box
}

func (box *Box) DestPosition() mgl32.Vec2 {
	return mgl32.Vec2{box.Dest.X, box.Dest.Y}
}

func (box *Box) Matrix() mgl32.Mat4 {
	if box.Transform != box.oldTransform {
		box.oldTransform = box.Transform
		bx, by := box.Dest.Center()
		scx := box.Dest.Width / 2.0
		scy := box.Dest.Height / 2.0
		centerScale := box.Scale
		if centerScale == 0.0 {
			// 0 means that a scale was not supplied, so use 1 as default.
			// If the box needs to be completely hidden, use the Hidden property instead of scaling to 0.
			centerScale = 1.0
		}
		box.matrix = mgl32.Translate3D(bx, by, box.Depth). // Move to final position
									Mul4(mgl32.ShearY3D(box.Shear, 0.0)).               // Apply shear
									Mul4(mgl32.Scale3D(scx, scy, 1.0)).                 // Scale by box dimensions according to top left
									Mul4(mgl32.Scale3D(centerScale, centerScale, 0.0)). // Overall scale
									Mul4(mgl32.Rotate3DZ(box.Rotation).Mat4())
	}
	return box.matrix
}

func (box *Box) Render(context *render.Context) {
	if box.Hidden {
		return
	}
	failure.CheckOpenGLError()
	mesh := box.Mesh
	if box.Mesh == nil {
		mesh = cache.QuadMesh
	}
	mesh.Bind()
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
			box.Src.X / texW,
			1.0 - (box.Src.Y / texH),
			box.Src.Width / texW,
			box.Src.Height / texH,
		}
		_ = shaders.UIShader.SetUniformVec4(shaders.UniformSrcRect, srcVec)
	} else {
		_ = shaders.UIShader.SetUniformBool(shaders.UniformNoTexture, true)
	}

	_ = shaders.UIShader.SetUniformBool(shaders.UniformFlipHorz, box.FlippedHorz)

	// Set uniforms
	_ = shaders.UIShader.SetUniformMatrix(shaders.UniformModelMatrix, box.Matrix())
	mesh.DrawAll()
	failure.CheckOpenGLError()
}
