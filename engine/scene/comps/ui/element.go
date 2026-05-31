package ui

import (
	"math"

	"github.com/go-gl/gl/v3.3-core/gl"
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
	TextAlignBottom
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
	BgFlippedHorz             bool
	Scissor                   math2.Rect
	text                      string
	textMesh                  *geom.Mesh
	textConfig                maybe.T[TextConfig]
	textMatrix                mgl32.Mat4
	sliceMatrices             [textures.SliceIndexCount]mgl32.Mat4
	transform                 Transform
	transformClean, textClean bool
	onScreenBox               math2.Rect
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

func NewColorBox(transform Transform, clr color.Color) Element {
	return Element{
		BgMesh:    cache.QuadMesh,
		BgColor:   maybe.Some(clr),
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

func (el *Element) TextConfig() *maybe.T[TextConfig] {
	return &el.textConfig
}

func (el *Element) SetTextConfig(config TextConfig) {
	if el.textConfig != maybe.Some(config) {
		el.textConfig = maybe.Some(config)
		el.textClean = false
	}
}

func (el *Element) recalculateMatrices() {
	if !el.transformClean {
		minX := float32(math.MaxFloat32)
		minY := float32(math.MaxFloat32)
		maxX := -minX
		maxY := -minY

		// Calculate bg matrix
		if el.BgMesh != nil {
			var ninePatchSlice textures.Slice
			if el.BgTexture != nil {
				ninePatchSlice = el.BgTexture.FindSlice("ninePatch")
				if ninePatchSlice == (textures.Slice{}) {
					ninePatchSlice.Bounds = el.BgTexture.Rect()
				}
			} else {
				ninePatchSlice.Bounds = math2.Rect{Width: 1, Height: 1}
			}
			boxSize := el.BgMesh.BoundingBox().Size()
			displacement := el.BgMesh.BoundingBox().Min.Vec2().Add(math2.ElemMul2(boxSize.Vec2(), mgl32.Vec2(el.transform.Origin))).Mul(-1.0)

			patches := ninePatchSlice.NinePatchRects()
			innerWidth := el.Size()[0] - patches[textures.SliceLeftMiddle].Width - patches[textures.SliceRightMiddle].Width
			innerHeight := el.Size()[1] - patches[textures.SliceTopMiddle].Height - patches[textures.SliceBottomMiddle].Height
			screenW, screenH := engine.ScreenSize()
			fScreenWidth, fScreenHeight := float32(screenW), float32(screenH)
			for i, patch := range patches {
				sliceIndex := textures.SliceIndex(i)

				var matrix mgl32.Mat4 = mgl32.Ident4()

				matrix = mgl32.Translate3D(displacement[0], displacement[1], 0.0).Mul4(matrix)

				var width, height float32
				switch sliceIndex {
				case textures.SliceTopLeft, textures.SliceTopRight, textures.SliceBottomRight, textures.SliceBottomLeft:
					width, height = patch.Width, patch.Height
				case textures.SliceLeftMiddle, textures.SliceRightMiddle:
					width = patch.Width
					height = innerHeight
				case textures.SliceTopMiddle, textures.SliceBottomMiddle:
					width = innerWidth
					height = patch.Height
				case textures.SliceCenter:
					width = innerWidth
					height = innerHeight
				}
				matrix = mgl32.Scale3D(width/boxSize[0], height/boxSize[1], 1.0).Mul4(matrix)

				var xOff float32
				switch sliceIndex {
				case textures.SliceTopMiddle, textures.SliceCenter, textures.SliceBottomMiddle:
					xOff = patches[textures.SliceLeftMiddle].Width
				case textures.SliceTopRight, textures.SliceRightMiddle, textures.SliceBottomRight:
					xOff = patches[textures.SliceLeftMiddle].Width + innerWidth
				}

				var yOff float32
				switch sliceIndex {
				case textures.SliceLeftMiddle, textures.SliceCenter, textures.SliceRightMiddle:
					yOff = patches[textures.SliceTopMiddle].Height
				case textures.SliceBottomLeft, textures.SliceBottomMiddle, textures.SliceBottomRight:
					yOff = patches[textures.SliceTopMiddle].Height + innerHeight
				}

				matrix = mgl32.Translate3D(xOff, yOff, 0.0).Mul4(matrix)

				// Rotate
				if el.Rotation() != 0 {
					matrix = mgl32.Rotate3DZ(float32(el.Rotation())).Mat4().Mul4(matrix)
				}

				// Apply shear
				if el.Shear() != (mgl32.Vec2{}) {
					matrix = mgl32.ShearY3D(el.Shear()[0], el.Shear()[1]).Mul4(matrix)
				}

				// Translate
				matrix = mgl32.Translate3D(
					el.Position()[0]+(el.Anchor()[0]*fScreenWidth),
					el.Position()[1]+(el.Anchor()[1]*fScreenHeight),
					el.Depth(),
				).Mul4(matrix)

				el.sliceMatrices[sliceIndex] = matrix
			}
		}

		// Calculate text matrix
		{
			// Displace by origin
			displacement := math2.ElemMul2(el.transform.Size, mgl32.Vec2(el.transform.Origin)).Mul(-1.0)
			el.textMatrix = mgl32.Translate3D(displacement[0], displacement[1], 0.0)

			// Rotate
			if el.Rotation() != 0 {
				el.textMatrix = mgl32.Rotate3DZ(float32(el.Rotation())).Mat4().Mul4(el.textMatrix)
			}

			// Translate
			screenW, screenH := engine.ScreenSize()
			fWidth, fHeight := float32(screenW), float32(screenH)
			el.textMatrix = mgl32.Translate3D(
				el.Position()[0]+(el.Anchor()[0]*fWidth),
				el.Position()[1]+(el.Anchor()[1]*fHeight),
				el.Depth(),
			).Mul4(el.textMatrix)

			// Calculate mesh boundaries
			for _, vert := range [...]mgl32.Vec3{
				{},
				el.transform.Size.Vec3(0.0),
				{el.transform.Size[0]},
				{0.0, el.transform.Size[1]},
			} {
				vert = mgl32.TransformCoordinate(vert, el.textMatrix)
				minX = min(minX, vert[0])
				maxX = max(maxX, vert[0])
				minY = min(minY, vert[1])
				maxY = max(maxY, vert[1])
			}
		}

		el.onScreenBox = math2.Rect{
			X:      minX,
			Y:      minY,
			Width:  maxX - minX,
			Height: maxY - minY,
		}

		el.transformClean = true
	}
}

// Force the element to recalculate its matrices next time it is rendered.
func (el *Element) MarkDirty() {
	el.textClean = false
	el.transformClean = false
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
	el.recalculateMatrices()
	return el.onScreenBox
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

func (el *Element) X() float32 {
	return el.transform.Position[0]
}

func (el *Element) Y() float32 {
	return el.transform.Position[1]
}

func (el *Element) Center() mgl32.Vec2 {
	return el.transform.Position.Add(math2.ElemMul2(mgl32.Vec2(el.transform.Origin).Mul(-1.0).Add(mgl32.Vec2{0.5, 0.5}), el.transform.Size))
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

func (el *Element) Width() float32 {
	return el.transform.Size[0]
}

func (el *Element) Height() float32 {
	return el.transform.Size[1]
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

func (el *Element) SetWidth(width float32) {
	el.transform.Size[0] = width
	el.transformClean = false
	el.textClean = false
}

func (el *Element) SetHeight(height float32) {
	el.transform.Size[1] = height
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
	el.recalculateMatrices()

	failure.CheckOpenGLError()

	if el.Scissor.Width > 0 || el.Scissor.Height > 0 {
		gl.Enable(gl.SCISSOR_TEST)
		_, scrHeight := engine.ScreenSize()
		gl.Scissor(int32(el.Scissor.X), int32(scrHeight)-int32(el.Scissor.Y)-int32(el.Scissor.Height), int32(el.Scissor.Width), int32(el.Scissor.Height))
		failure.CheckOpenGLError()
		defer func() {
			gl.Disable(gl.SCISSOR_TEST)
			failure.CheckOpenGLError()
		}()
	}

	shaders.UIShader.Use()
	failure.CheckOpenGLError()

	_ = context.SetUniforms(shaders.UIShader)
	failure.CheckOpenGLError()
	_ = shaders.UIShader.SetUniformInt(shaders.UniformTex, 0)
	failure.CheckOpenGLError()

	if el.BgMesh != nil {
		el.BgMesh.Bind()

		_ = shaders.UIShader.SetUniformVec4(shaders.UniformDiffuseColor, el.BgColor.Or(color.White).Vector())
		failure.CheckOpenGLError()

		var ninePatchSlice textures.Slice
		if el.BgTexture != nil {
			ninePatchSlice = el.BgTexture.FindSlice("ninePatch")
			if ninePatchSlice == (textures.Slice{}) {
				ninePatchSlice.Bounds = el.BgTexture.Rect()
			}
		} else {
			ninePatchSlice.Bounds = math2.Rect{Width: 1, Height: 1}
		}

		patches := ninePatchSlice.NinePatchRects()
		for i, patch := range patches {
			if patch.Width == 0 || patch.Height == 0 {
				continue
			}

			if el.BgTexture != nil {
				el.BgTexture.Bind()
				failure.CheckOpenGLError()
				texRect := el.BgTexture.Rect()
				_ = shaders.UIShader.SetUniformBool(shaders.UniformNoTexture, false)
				failure.CheckOpenGLError()
				_ = shaders.UIShader.SetUniformBool(shaders.UniformFlipHorz, el.BgFlippedHorz)
				failure.CheckOpenGLError()
				animFrame := el.AnimPlayer.FrameUV()
				_ = shaders.UIShader.SetUniformVec4(shaders.UniformSrcRect, mgl32.Vec4{
					animFrame.X + (patch.X/texRect.Width)*animFrame.Width,
					animFrame.Y - ((patch.Y / texRect.Height) * animFrame.Height),
					(patch.Width / texRect.Width) * animFrame.Width,
					(patch.Height / texRect.Height) * animFrame.Height,
				})
				failure.CheckOpenGLError()
			} else {
				_ = shaders.UIShader.SetUniformBool(shaders.UniformNoTexture, true)
				failure.CheckOpenGLError()
			}

			_ = shaders.UIShader.SetUniformMatrix(shaders.UniformModelMatrix, el.sliceMatrices[i])
			failure.CheckOpenGLError()

			el.BgMesh.DrawAll()
			failure.CheckOpenGLError()
		}
	}

	textConfig, hasTextConfig := el.textConfig.Get()
	if el.TextMesh() != nil && hasTextConfig && textConfig.Font != nil {
		el.TextMesh().Bind()
		failure.CheckOpenGLError()

		cache.GetTexture(textConfig.Font.TexturePath()).Bind()
		_ = shaders.UIShader.SetUniformBool(shaders.UniformNoTexture, false)
		_ = shaders.UIShader.SetUniformVec4(shaders.UniformSrcRect, el.AnimPlayer.FrameUV().Vec4())
		failure.CheckOpenGLError()

		// Draw drop shadow
		if !textConfig.DisableShadow {
			shadowMatrix := mgl32.Translate3D(4.0, 4.0, 0.0).Mul4(el.textMatrix)
			_ = shaders.UIShader.SetUniformMatrix(shaders.UniformModelMatrix, shadowMatrix)
			_ = shaders.UIShader.SetUniformVec4(shaders.UniformDiffuseColor, textShadowColor.Vector())
			failure.CheckOpenGLError()
			el.TextMesh().DrawAll()
		}

		_ = shaders.UIShader.SetUniformVec4(shaders.UniformDiffuseColor, textConfig.Color.Or(color.White).Vector())
		failure.CheckOpenGLError()

		_ = shaders.UIShader.SetUniformMatrix(shaders.UniformModelMatrix, el.textMatrix)
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
