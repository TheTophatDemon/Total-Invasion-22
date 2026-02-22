package ui

import (
	"math"

	"github.com/go-gl/glfw/v3.3/glfw"
	"github.com/go-gl/mathgl/mgl32"
	"tophatdemon.com/total-invasion-ii/engine"
	"tophatdemon.com/total-invasion-ii/engine/assets/cache"
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

type Element struct {
	BgColor                   maybe.T[color.Color]
	AnimPlayer                comps.AnimationPlayer
	BgTexture                 *textures.Texture
	BgMesh                    *geom.Mesh
	text                      string
	textMesh                  *geom.Mesh
	textConfig                maybe.T[TextConfig]
	boxMatrix, textMatrix     mgl32.Mat4
	transform                 Transform
	transformClean, textClean bool
}

func NewBox(transform Transform, texture *textures.Texture) Element {
	if texture != nil && transform.Size.ApproxEqual(mgl32.Vec2{}) {
		transform.Size = mgl32.Vec2{float32(texture.Width()), float32(texture.Height())}
	}
	return Element{
		BgMesh:    cache.QuadMesh,
		BgTexture: texture,
		transform: transform,
	}
}

func (el *Element) HasText() bool {
	return el.textConfig.IsSome()
}

func (el *Element) SetText(text string) {
	if !el.textConfig.IsSome() {
		el.textConfig = maybe.Some(TextConfig{
			Font: cache.DefaultFont,
		})
	}
	if el.text != text {
		el.text = text
		el.textClean = false
	}
}

func (el *Element) Text() string {
	return el.text
}

func (el *Element) TextConfig() maybe.T[TextConfig] {
	return el.textConfig
}

func (el *Element) SetTextConfig(config TextConfig) {
	if el.textConfig != maybe.Some(config) {
		el.textConfig = maybe.Some(config)
		el.textClean = false
	}
}

// Gets a displacement vector that goes from the element's origin to the BgMesh's origin
func (el *Element) getBoxDisplacement() mgl32.Vec2 {
	var mesh = el.BgMesh
	if mesh == nil {
		mesh = cache.QuadMesh
	}
	bbox := mesh.BoundingBox()
	return bbox.Min.Vec2().Add(math2.ElemMul2(bbox.Size().Vec2(), mgl32.Vec2(el.transform.Origin))).Mul(-1.0)
}

// Gets a displacement vector that goes from the element's origin to the text mesh's origin
func (el *Element) getTextDisplacement() mgl32.Vec2 {
	return math2.ElemMul2(el.transform.Size, mgl32.Vec2(el.transform.Origin)).Mul(-1.0)
}

func (el *Element) BgMatrix() mgl32.Mat4 {
	if !el.transformClean {
		// Calculate bg matrix
		{
			// Displace by origin
			displacement := el.getBoxDisplacement()
			el.boxMatrix = mgl32.Translate3D(displacement[0], displacement[1], 0.0)

			// Rotate
			if el.Rotation() != 0 {
				el.boxMatrix = mgl32.Rotate3DZ(float32(el.Rotation())).Mat4().Mul4(el.boxMatrix)
			}

			// Scale
			mesh := el.BgMesh
			if mesh == nil {
				mesh = cache.QuadMesh
			}
			boxSize := mesh.BoundingBox().Size()
			el.boxMatrix = mgl32.Scale3D(el.Size()[0]/boxSize[0], el.Size()[1]/boxSize[1], 1.0).Mul4(el.boxMatrix)

			// Apply shear
			if el.Shear() != (mgl32.Vec2{}) {
				el.boxMatrix = mgl32.ShearY3D(el.Shear()[0], el.Shear()[1]).Mul4(el.boxMatrix)
			}

			// Translate
			screenW, screenH := engine.ScreenSize()
			fWidth, fHeight := float32(screenW), float32(screenH)
			el.boxMatrix = mgl32.Translate3D(
				el.Position()[0]+(el.Anchor()[0]*fWidth),
				el.Position()[1]+(el.Anchor()[1]*fHeight),
				el.Depth(),
			).Mul4(el.boxMatrix)
		}

		// Calculate text matrix
		{
			// Displace by origin
			displacement := el.getTextDisplacement()
			el.textMatrix = mgl32.Translate3D(displacement[0], displacement[1], 0.0)

			// Rotate
			if el.Rotation() != 0 {
				el.textMatrix = mgl32.Rotate3DZ(float32(el.Rotation())).Mat4().Mul4(el.textMatrix)
			}

			// Apply shear
			if el.Shear() != (mgl32.Vec2{}) {
				el.textMatrix = mgl32.ShearY3D(el.Shear()[0], el.Shear()[1]).Mul4(el.textMatrix)
			}

			// Translate
			screenW, screenH := engine.ScreenSize()
			fWidth, fHeight := float32(screenW), float32(screenH)
			el.textMatrix = mgl32.Translate3D(
				el.Position()[0]+(el.Anchor()[0]*fWidth),
				el.Position()[1]+(el.Anchor()[1]*fHeight),
				el.Depth(),
			).Mul4(el.textMatrix)
		}

		el.transformClean = true
	}
	return el.boxMatrix
}

func (el *Element) TextMatrix() mgl32.Mat4 {
	el.BgMatrix()
	return el.textMatrix
}

func (el *Element) TextMesh() *geom.Mesh {
	_, hasTextConfig := el.textConfig.Get()
	if hasTextConfig && (!el.transformClean || !el.textClean) {
		el.generateTextMesh()
		el.textClean = true
	}
	return el.textMesh
}

// Returns the rectangle that the element occupies on the screen
func (el *Element) OnScreenBox() math2.Rect {
	minX := float32(math.MaxFloat32)
	minY := float32(math.MaxFloat32)
	maxX := float32(-math.MaxFloat32)
	maxY := float32(-math.MaxFloat32)
	for _, pos := range cache.QuadMesh.Verts().Pos {
		pos = mgl32.TransformCoordinate(pos, el.BgMatrix())
		minX = min(pos[0], minX)
		minY = min(pos[1], minY)
		maxX = max(pos[0], maxX)
		maxY = max(pos[1], maxY)
	}
	return math2.Rect{
		X:      minX,
		Y:      minY,
		Width:  maxX - minX,
		Height: maxY - minY,
	}
}

func (el *Element) Transform() Transform {
	return el.transform
}

func (el *Element) SetTransform(trans Transform) {
	if el.transform != trans {
		if el.Size() != trans.Size {
			el.textClean = false
		}
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
	el.textClean = false
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

// Expands the element to the specified height while keeping its current aspect ratio.
func (el *Element) FitHeight(targetHeight float32) {
	size := el.Size()
	el.SetSize(size.Mul(targetHeight / size[1]))
}

func (el *Element) Render(context *render.Context) {
	if el.BgMesh == nil && el.TextMesh() == nil {
		return
	}
	failure.CheckOpenGLError()

	shaders.UIShader.Use()
	failure.CheckOpenGLError()

	_ = context.SetUniforms(shaders.UIShader)
	failure.CheckOpenGLError()
	_ = shaders.UIShader.SetUniformInt(shaders.UniformTex, 0)
	failure.CheckOpenGLError()

	if el.BgMesh != nil {
		if el.BgTexture != nil {
			el.BgTexture.Bind()
			_ = shaders.UIShader.SetUniformBool(shaders.UniformNoTexture, false)
			_ = shaders.UIShader.SetUniformVec4(shaders.UniformSrcRect, el.AnimPlayer.FrameUV().Vec4())
		} else {
			_ = shaders.UIShader.SetUniformBool(shaders.UniformNoTexture, true)
		}
		failure.CheckOpenGLError()

		el.BgMesh.Bind()

		_ = shaders.UIShader.SetUniformVec4(shaders.UniformDiffuseColor, el.BgColor.Or(color.White).Vector())
		failure.CheckOpenGLError()

		_ = shaders.UIShader.SetUniformMatrix(shaders.UniformModelMatrix, el.BgMatrix())
		failure.CheckOpenGLError()

		el.BgMesh.DrawAll()
		failure.CheckOpenGLError()
	}

	textConfig, hasTextConfig := el.textConfig.Get()
	if el.TextMesh() != nil && hasTextConfig && textConfig.Font != nil {
		el.TextMesh().Bind()
		failure.CheckOpenGLError()

		cache.GetTexture(textConfig.Font.TexturePath()).Bind()
		_ = shaders.UIShader.SetUniformBool(shaders.UniformNoTexture, false)
		_ = shaders.UIShader.SetUniformVec4(shaders.UniformSrcRect, el.AnimPlayer.FrameUV().Vec4())
		failure.CheckOpenGLError()

		textMatrix := el.TextMatrix()

		// Draw drop shadow
		if !textConfig.DisableShadow {
			shadowMatrix := mgl32.Translate3D(4.0, 4.0, 0.0).Mul4(textMatrix)
			_ = shaders.UIShader.SetUniformMatrix(shaders.UniformModelMatrix, shadowMatrix)
			_ = shaders.UIShader.SetUniformVec4(shaders.UniformDiffuseColor, textShadowColor.Vector())
			failure.CheckOpenGLError()
			el.TextMesh().DrawAll()
		}

		_ = shaders.UIShader.SetUniformVec4(shaders.UniformDiffuseColor, textConfig.Color.Or(color.White).Vector())
		failure.CheckOpenGLError()

		_ = shaders.UIShader.SetUniformMatrix(shaders.UniformModelMatrix, textMatrix)
		failure.CheckOpenGLError()

		el.TextMesh().DrawAll()
		failure.CheckOpenGLError()
	}

	if engine.InDebugMode() && input.IsMouseButtonDown(glfw.MouseButton3) {
		screenRect := el.OnScreenBox()
		wireMesh := geom.CreateWireMesh(geom.Vertices{
			Pos: []mgl32.Vec3{
				{screenRect.X, screenRect.Y, 0.0},
				{screenRect.X + screenRect.Width, screenRect.Y, 0.0},
				{screenRect.X + screenRect.Width, screenRect.Y + screenRect.Height, 0.0},
				{screenRect.X, screenRect.Y + screenRect.Height, 0.0},
			},
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
		_ = shaders.DebugShader.SetUniformMatrix(shaders.UniformModelMatrix, mgl32.Ident4())
		failure.CheckOpenGLError()
		wireMesh.DrawAll()
		wireMesh.Free()
	}
	failure.CheckOpenGLError()
}
