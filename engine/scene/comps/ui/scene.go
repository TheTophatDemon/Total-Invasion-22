package ui

import (
	"log"

	"github.com/go-gl/gl/v3.3-core/gl"
	"github.com/go-gl/mathgl/mgl32"
	"tophatdemon.com/total-invasion-ii/engine/assets/cache"
	"tophatdemon.com/total-invasion-ii/engine/assets/shaders"
	"tophatdemon.com/total-invasion-ii/engine/render"
	world "tophatdemon.com/total-invasion-ii/engine/scene"
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
	gl.Disable(gl.DEPTH_TEST)

	// Render boxes in one large batch
	cache.QuadMesh.Bind()
	shaders.UIShader.Use()

	// Box uniforms
	_ = context.SetUniforms(shaders.UIShader)
	_ = shaders.UIShader.SetUniformInt(shaders.UniformTex, 0)

	scene.Boxes.ForEach(func(box *Box) {
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
	})

	_ = shaders.UIShader.SetUniformBool(shaders.UniformNoTexture, false)

	// Render text
	scene.Texts.ForEach(func(text *Text) {
		if len(text.text) == 0 {
			return
		}

		// Set color
		_ = shaders.UIShader.SetUniformVec4(shaders.UniformDiffuseColor, text.Color().Vector())

		_ = shaders.UIShader.SetUniformBool(shaders.UniformFlipHorz, false)

		// Set texture
		text.texture.Bind()
		_ = shaders.UIShader.SetUniformVec4(shaders.UniformSrcRect, mgl32.Vec4{0.0, 1.0, 1.0, 1.0})

		// Set transform
		_ = shaders.UIShader.SetUniformMatrix(shaders.UniformModelMatrix, text.Transform())

		// Draw
		if mesh, err := text.Mesh(); err == nil {
			mesh.Bind()
			mesh.DrawAll()
		} else {
			log.Println(err)
		}
	})

	gl.Enable(gl.DEPTH_TEST)
}
