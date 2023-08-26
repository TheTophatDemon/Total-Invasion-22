package ui

import (
	"log"

	"github.com/go-gl/gl/v3.3-core/gl"
	"github.com/go-gl/mathgl/mgl32"
	"tophatdemon.com/total-invasion-ii/engine/assets"
	"tophatdemon.com/total-invasion-ii/engine/math2"
	"tophatdemon.com/total-invasion-ii/engine/render"
	"tophatdemon.com/total-invasion-ii/engine/world"
)

type Scene struct {
	Boxes *world.Storage[Box]
	Texts *world.Storage[Text]
}

func NewUIScene(maxBoxes, maxTexts uint) *Scene {
	return &Scene{
		Boxes: world.NewStorage[Box](maxBoxes),
		Texts: world.NewStorage[Text](maxTexts),
	}
}

func (scene *Scene) Update(deltaTime float32) {
	scene.Boxes.Update((*Box).Update, deltaTime)
}

func (scene *Scene) Render(context *render.Context) {
	gl.Clear(gl.DEPTH_BUFFER_BIT)
	gl.CullFace(gl.FRONT)

	// Render boxes in one large batch
	assets.QuadMesh.Bind()
	assets.UIShader.Use()

	// Box uniforms
	_ = context.SetUniforms(assets.UIShader)
	_ = assets.UIShader.SetUniformInt(assets.UniformTex, 0)
	_ = assets.UIShader.SetUniformInt(assets.UniformAtlas, 1)
	_ = assets.UIShader.SetUniformInt(assets.UniformFrame, 0)

	scene.Boxes.ForEach(func(box *Box) {
		// Set animation frame.
		_ = assets.UIShader.SetUniformInt(assets.UniformFrame, box.AnimPlayer.Frame())

		// Set color.
		_ = assets.UIShader.SetUniformVec4(assets.UniformDiffuseColor, math2.ColorToVec4(box.Color))

		// Set texture.
		if box.Texture != nil {
			box.Texture.Bind()
			texW, texH := float32(box.Texture.Width()), float32(box.Texture.Height())
			_ = assets.UIShader.SetUniformBool(assets.UniformAtlasUsed, box.Texture.IsAtlas())
			srcVec := mgl32.Vec4{
				box.src.X / texW,
				box.src.Y / texH,
				box.src.Width / texW,
				box.src.Height / texH,
			}
			_ = assets.UIShader.SetUniformVec4(assets.UniformSrcRect, srcVec)
		}

		// Set uniforms
		_ = assets.UIShader.SetUniformMatrix(assets.UniformModelMatrix, box.Transform())
		assets.QuadMesh.DrawAll()
	})

	_ = assets.UIShader.SetUniformInt(assets.UniformFrame, 0)

	// Render text
	scene.Texts.ForEach(func(text *Text) {
		// Set color
		_ = assets.UIShader.SetUniformVec4(assets.UniformDiffuseColor, math2.ColorToVec4(text.Color()))

		// Set texture
		text.texture.Bind()
		_ = assets.UIShader.SetUniformBool(assets.UniformAtlasUsed, text.texture.IsAtlas())
		srcRect := math2.Rect{X: 0.0, Y: 0.0, Width: 1.0, Height: 1.0}
		_ = assets.UIShader.SetUniformVec4(assets.UniformSrcRect, srcRect.Vec4())

		// Set transform
		_ = assets.UIShader.SetUniformMatrix(assets.UniformModelMatrix, text.Transform())

		// Draw
		if mesh, err := text.Mesh(); err == nil {
			mesh.Bind()
			mesh.DrawAll()
		} else {
			log.Println(err)
		}
	})
}
