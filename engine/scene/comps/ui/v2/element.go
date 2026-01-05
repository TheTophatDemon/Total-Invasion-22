package ui

import (
	"github.com/go-gl/glfw/v3.3/glfw"
	"github.com/go-gl/mathgl/mgl32"
	"tophatdemon.com/total-invasion-ii/engine"
	"tophatdemon.com/total-invasion-ii/engine/assets/cache"
	"tophatdemon.com/total-invasion-ii/engine/assets/fonts"
	"tophatdemon.com/total-invasion-ii/engine/assets/geom"
	"tophatdemon.com/total-invasion-ii/engine/assets/shaders"
	"tophatdemon.com/total-invasion-ii/engine/assets/textures"
	"tophatdemon.com/total-invasion-ii/engine/color"
	"tophatdemon.com/total-invasion-ii/engine/containers/maybe"
	"tophatdemon.com/total-invasion-ii/engine/failure"
	"tophatdemon.com/total-invasion-ii/engine/input"
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

// IDEA: Would it be useful to be able to draw both the quad and the text at the same time on one element?
type Element struct {
	Color                     color.Color
	AnimPlayer                comps.AnimationPlayer
	ShadowColor               color.Color // Color of the drop shadow. Set to transparent to disable.
	ShadowOffset              mgl32.Vec2
	Texture                   *textures.Texture
	Mesh                      *geom.Mesh
	font                      *fonts.Font
	text                      string
	textConfig                maybe.T[TextConfig]
	matrix                    mgl32.Mat4
	transform                 Transform
	transformClean, textClean bool
}

func NewBox(transform Transform, texture *textures.Texture) Element {
	return Element{
		Mesh:      cache.QuadMesh,
		Texture:   texture,
		transform: transform,
	}
}

func (el *Element) HasText() bool {
	return el.textConfig.IsSome()
}

func (el *Element) SetText(text string, config TextConfig) {
	if config.Font == nil {
		el.font = cache.DefaultFont
		el.Mesh = nil // Prevents previous mesh from getting freed by generateTextMesh()
	}
	el.Texture = cache.GetTexture(el.font.TexturePath())
	if el.text != text || el.textConfig != maybe.Some(config) {
		el.textConfig = maybe.Some(config)
		el.textClean = false
		el.text = text
	}
}

func (el *Element) Text() string {
	return el.text
}

func (el *Element) ShrinkToFitText() {
	if el.Mesh == nil || !el.HasText() {
		return
	}
	el.transform.Size = el.Mesh.BoundingBox().Size().Vec2()
}

// Gets a displacement vector that goes from the element's origin to the mesh's origin
func (el *Element) getDisplacement() mgl32.Vec2 {
	if !el.HasText() {
		bbox := el.Mesh.BoundingBox()
		return bbox.Min.Vec2().Add(math2.ElemMul2(bbox.Size().Vec2(), mgl32.Vec2(el.transform.Origin))).Mul(-1.0)
	} else {
		return math2.ElemMul2(el.transform.Size, mgl32.Vec2(el.transform.Origin)).Mul(-1.0)
	}
}

// Returns the rectangle that the element occupies on the screen
func (el *Element) OnScreenBox() math2.Rect {
	min := el.transform.Position.Add(el.getDisplacement())
	return math2.Rect{
		X:      min[0],
		Y:      min[1],
		Width:  el.transform.Size[0],
		Height: el.transform.Size[1],
	}
}

func (el *Element) Transform() Transform {
	return el.transform
}

func (el *Element) SetTransform(trans Transform) {
	if el.transform != trans {
		el.transform = trans
		el.transformClean = false
	}
}

func (el *Element) Position() mgl32.Vec2 {
	return el.transform.Position
}

func (el *Element) Rotation() math2.Radians {
	return el.transform.Rotation
}

func (el *Element) Size() mgl32.Vec2 {
	return el.transform.Size
}

func (el *Element) Anchor() Ratios {
	return el.transform.Anchor
}

func (el *Element) Origin() Ratios {
	return el.transform.Origin
}

func (el *Element) Shear() mgl32.Vec2 {
	return el.transform.Shear
}

func (el *Element) Depth() float32 {
	return el.transform.Depth
}

func (el *Element) SetPosition(value mgl32.Vec2) {
	el.transform.Position = value
	el.transformClean = false
}

func (el *Element) Translate(offset mgl32.Vec2) {
	el.transform.Position = el.transform.Position.Add(offset)
	el.transformClean = false
}

func (el *Element) SetX(x float32) {
	el.transform.Position[0] = x
	el.transformClean = false
}

func (el *Element) SetY(y float32) {
	el.transform.Position[1] = y
	el.transformClean = false
}

func (el *Element) SetRotation(value math2.Radians) {
	el.transform.Rotation = value
	el.transformClean = false
}

func (el *Element) Rotate(amount math2.Radians) {
	el.transform.Rotation += amount
	el.transformClean = false
}

func (el *Element) SetSize(value mgl32.Vec2) {
	el.transform.Size = value
	el.transformClean = false
}

func (el *Element) SetAnchor(value Ratios) {
	el.transform.Anchor = value
	el.transformClean = false
}

func (el *Element) SetOrigin(value Ratios) {
	el.transform.Origin = value
	el.transformClean = false
}

func (el *Element) SetShear(value mgl32.Vec2) {
	el.transform.Shear = value
	el.transformClean = false
}

func (el *Element) SetDepth(value float32) {
	el.transform.Depth = value
	el.transformClean = false
}

func (el *Element) Render(context *render.Context) {
	if el.textConfig.IsSome() && (!el.transformClean || !el.textClean) {
		el.generateTextMesh(*el.textConfig.Unwrap())
		el.textClean = true
	}

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

	if el.Texture != nil {
		el.Texture.Bind()
		_ = shaders.UIShader.SetUniformBool(shaders.UniformNoTexture, false)
		_ = shaders.UIShader.SetUniformVec4(shaders.UniformSrcRect, el.AnimPlayer.FrameUV().Vec4())
	} else {
		_ = shaders.UIShader.SetUniformBool(shaders.UniformNoTexture, true)
	}
	failure.CheckOpenGLError()

	if !el.transformClean {
		// Displace by origin
		displacement := el.getDisplacement()
		el.matrix = mgl32.Translate3D(displacement[0], displacement[1], 0.0)

		// Rotate
		if el.Rotation() != 0 {
			el.matrix = mgl32.Rotate3DZ(float32(el.Rotation())).Mat4().Mul4(el.matrix)
		}

		if !el.HasText() {
			// Scale
			boxSize := el.Mesh.BoundingBox().Size()
			el.matrix = mgl32.Scale3D(el.Size()[0]/boxSize[0], el.Size()[1]/boxSize[1], 1.0).Mul4(el.matrix)
		}

		// Apply shear
		if el.Shear() != (mgl32.Vec2{}) {
			el.matrix = mgl32.ShearY3D(el.Shear()[0], el.Shear()[1]).Mul4(el.matrix)
		}

		// Translate
		screenW, screenH := engine.ScreenSize()
		fWidth, fHeight := float32(screenW), float32(screenH)
		el.matrix = mgl32.Translate3D(
			el.Position()[0]+(el.Anchor()[0]*fWidth),
			el.Position()[1]+(el.Anchor()[1]*fHeight),
			el.Depth(),
		).Mul4(el.matrix)

		el.transformClean = true
	}

	el.Mesh.Bind()

	// Draw drop shadow
	if el.ShadowColor.A > 0.0 {
		shadowMatrix := mgl32.Translate3D(el.ShadowOffset[0], el.ShadowOffset[1], 0.0).Mul4(el.matrix)
		_ = shaders.UIShader.SetUniformMatrix(shaders.UniformModelMatrix, shadowMatrix)
		_ = shaders.UIShader.SetUniformVec4(shaders.UniformDiffuseColor, el.ShadowColor.Vector())
		failure.CheckOpenGLError()
		el.Mesh.DrawAll()
	}

	if el.Color == (color.Color{}) {
		// Set default color if unset
		el.Color = color.White
	}
	_ = shaders.UIShader.SetUniformVec4(shaders.UniformDiffuseColor, el.Color.Vector())
	failure.CheckOpenGLError()

	_ = shaders.UIShader.SetUniformMatrix(shaders.UniformModelMatrix, el.matrix)
	failure.CheckOpenGLError()

	el.Mesh.DrawAll()
	failure.CheckOpenGLError()

	if engine.InDebugMode() && input.IsMouseButtonDown(glfw.MouseButton3) {
		var pozzy []mgl32.Vec3

		if el.HasText() {
			pozzy = []mgl32.Vec3{
				{0.0, 0.0, 0.0},
				{el.Size()[0], 0.0, 0.0},
				{el.Size()[0], el.Size()[1], 0.0},
				{0.0, el.Size()[1], 0.0},
			}
		} else {
			pozzy = []mgl32.Vec3{
				{-1.0, -1.0, 0.0},
				{1.0, -1.0, 0.0},
				{1.0, 1.0, 0.0},
				{-1.0, 1.0, 0.0},
			}
		}
		wireMesh := geom.CreateWireMesh(geom.Vertices{
			Pos: pozzy,
			Color: []mgl32.Vec4{
				{1.0, 1.0, 1.0, 1.0},
				{1.0, 1.0, 1.0, 1.0},
				{1.0, 1.0, 1.0, 1.0},
				{1.0, 1.0, 1.0, 1.0},
			},
		}, []uint32{0, 1, 1, 2, 2, 3, 3, 0})
		shaders.DebugShader.Use()
		wireMesh.Bind()
		failure.CheckOpenGLError()
		_ = context.SetUniforms(shaders.DebugShader)
		_ = shaders.DebugShader.SetUniformMatrix(shaders.UniformModelMatrix, el.matrix)
		failure.CheckOpenGLError()
		wireMesh.DrawAll()
		wireMesh.Free()
	}
	failure.CheckOpenGLError()
}
