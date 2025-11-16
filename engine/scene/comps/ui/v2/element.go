package ui

import (
	"github.com/go-gl/mathgl/mgl32"
	"tophatdemon.com/total-invasion-ii/engine"
	"tophatdemon.com/total-invasion-ii/engine/assets/cache"
	"tophatdemon.com/total-invasion-ii/engine/assets/fonts"
	"tophatdemon.com/total-invasion-ii/engine/assets/geom"
	"tophatdemon.com/total-invasion-ii/engine/assets/shaders"
	"tophatdemon.com/total-invasion-ii/engine/assets/textures"
	"tophatdemon.com/total-invasion-ii/engine/color"
	"tophatdemon.com/total-invasion-ii/engine/failure"
	"tophatdemon.com/total-invasion-ii/engine/math2"
	"tophatdemon.com/total-invasion-ii/engine/render"
	"tophatdemon.com/total-invasion-ii/engine/scene/comps"
)

type TextAlign uint8

const (
	TextAlignTopLeft TextAlign = 0
	TextAlignCenterH TextAlign = 1 << (iota - 1)
	TextAlignCenterV
	TextAlignRight
)

type Ratios [2]float32 // Defines a 2d vector representing fractions of a rectangle

type Transform struct {
	Position    mgl32.Vec2 // Position of origin point on screen in pixels
	Origin      Ratios     // Dimensionally independent point of origin on the element itself
	Anchor      Ratios     // Dimensionally independent point of origin on the screen
	Rotation    math2.Radians
	Size, Shear mgl32.Vec2 // Size and shear in pixels
	Depth       float32    // Stacking order of the element
}

type Element struct {
	Transform
	Color         color.Color
	AnimPlayer    comps.AnimationPlayer
	TextAlignment TextAlign
	ShadowColor   color.Color // Color of the drop shadow. Set to transparent to disable.
	ShadowOffset  mgl32.Vec2
	WrapWords     bool // Whether text wrapping around the boundary is done by word instead of by character
	Mesh          *geom.Mesh
	Texture       *textures.Texture
	font          *fonts.Font
	text          string
	matrix        mgl32.Mat4
	oldTransform  Transform
}

func NewBox(transform Transform, texture *textures.Texture) Element {
	return Element{
		Transform: transform,
		Mesh:      cache.QuadMesh,
		Texture:   texture,
	}
}

func (el *Element) IsText() bool {
	return el.font != nil
}

func (el *Element) SetText(text string) {
	el.SetTextWithFont(text, cache.DefaultFont)
}

func (el *Element) Text() string {
	return el.text
}

func (el *Element) SetTextWithFont(text string, font *fonts.Font) {
	if font == nil {
		font = cache.DefaultFont
		el.Mesh = nil // Prevents previous mesh from getting freed by generateTextMesh()
	}
	el.Texture = cache.GetTexture(font.TexturePath())
	if el.text != text {
		defer el.generateTextMesh()
	}
	el.text = text
	el.font = font
}

func (el *Element) ShrinkToFitText() {
	if el.Mesh == nil || !el.IsText() {
		return
	}
	el.Size = el.Mesh.BoundingBox().Size().Vec2()
}

// Gets a displacement vector that goes from the element's origin to the mesh's origin
func (el *Element) getDisplacement() mgl32.Vec2 {
	if !el.IsText() {
		bbox := el.Mesh.BoundingBox()
		return bbox.Min.Vec2().Add(math2.ElemMul2(bbox.Size().Vec2(), mgl32.Vec2(el.Origin))).Mul(-1.0)
	} else {
		return math2.ElemMul2(el.Size, mgl32.Vec2(el.Origin)).Mul(-1.0)
	}
}

// Returns the rectangle that the element occupies on the screen
func (el *Element) OnScreenBox() math2.Rect {
	min := el.Position.Add(el.getDisplacement())
	return math2.Rect{
		X:      min[0],
		Y:      min[1],
		Width:  el.Size[0],
		Height: el.Size[1],
	}
}

func (el *Element) Render(context *render.Context) {
	if el.Mesh == nil {
		failure.LogWarningWithLocation("UI element rendering without mesh")
		return
	}
	failure.CheckOpenGLError()

	shaders.UIShader.Use()
	failure.CheckOpenGLError()

	_ = context.SetUniforms(shaders.UIShader)
	failure.CheckOpenGLError()
	_ = shaders.UIShader.SetUniformInt(shaders.UniformTex, 0)
	failure.CheckOpenGLError()

	if el.Color == (color.Color{}) {
		// Set default color if unset
		el.Color = color.White
	}
	_ = shaders.UIShader.SetUniformVec4(shaders.UniformDiffuseColor, el.Color.Vector())
	failure.CheckOpenGLError()

	if el.Texture != nil {
		el.Texture.Bind()
		_ = shaders.UIShader.SetUniformBool(shaders.UniformNoTexture, false)
		_ = shaders.UIShader.SetUniformVec4(shaders.UniformSrcRect, el.AnimPlayer.FrameUV().Vec4())
	} else {
		_ = shaders.UIShader.SetUniformBool(shaders.UniformNoTexture, true)
	}
	failure.CheckOpenGLError()

	if el.oldTransform != el.Transform {
		el.oldTransform = el.Transform

		// Displace by origin
		displacement := el.getDisplacement()
		el.matrix = mgl32.Translate3D(displacement[0], displacement[1], 0.0)

		// Rotate
		if el.Rotation != 0 {
			el.matrix = mgl32.Rotate3DZ(float32(el.Rotation)).Mat4().Mul4(el.matrix)
		}

		if !el.IsText() {
			// Scale
			el.matrix = mgl32.Scale3D(el.Size[0], el.Size[1], 1.0).Mul4(el.matrix)
		}

		// Apply shear
		if el.Shear != (mgl32.Vec2{}) {
			el.matrix = mgl32.ShearY3D(el.Shear[0], el.Shear[1]).Mul4(el.matrix)
		}

		// Translate
		screenW, screenH := engine.ScreenSize()
		fWidth, fHeight := float32(screenW), float32(screenH)
		el.matrix = mgl32.Translate3D(
			el.Position[0]+(el.Anchor[0]*fWidth),
			el.Position[1]+(el.Anchor[1]*fHeight),
			el.Depth,
		).Mul4(el.matrix)
	}

	_ = shaders.UIShader.SetUniformMatrix(shaders.UniformModelMatrix, el.matrix)
	failure.CheckOpenGLError()

	el.Mesh.Bind()
	el.Mesh.DrawAll()
	failure.CheckOpenGLError()

	// TODO: Generate drop shadow

	failure.CheckOpenGLError()
}
