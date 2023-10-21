package ui

import (
	"log"

	"github.com/go-gl/gl/v3.3-core/gl"
	"github.com/go-gl/mathgl/mgl32"
	"tophatdemon.com/total-invasion-ii/engine/assets/cache"
	"tophatdemon.com/total-invasion-ii/engine/assets/shaders"
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
	cache.QuadMesh.Bind()
	shaders.UIShader.Use()

	// Box uniforms
	_ = context.SetUniforms(shaders.UIShader)
	_ = shaders.UIShader.SetUniformInt(shaders.UniformTex, 0)

	scene.Boxes.ForEach(func(box *Box) {
		// Set color.
		_ = shaders.UIShader.SetUniformVec4(shaders.UniformDiffuseColor, math2.ColorToVec4(box.Color))

		// Set texture.
		if box.Texture != nil {
			box.Texture.Bind()
			texW, texH := float32(box.Texture.Width()), float32(box.Texture.Height())
			srcVec := mgl32.Vec4{
				box.src.X / texW,
				box.src.Y / texH,
				box.src.Width / texW,
				box.src.Height / texH,
			}
			_ = shaders.UIShader.SetUniformVec4(shaders.UniformSrcRect, srcVec)
		}

		// Set uniforms
		_ = shaders.UIShader.SetUniformMatrix(shaders.UniformModelMatrix, box.Transform())
		cache.QuadMesh.DrawAll()
	})

	// Render text
	scene.Texts.ForEach(func(text *Text) {
		// Set color
		_ = shaders.UIShader.SetUniformVec4(shaders.UniformDiffuseColor, math2.ColorToVec4(text.Color()))

		// Set texture
		text.texture.Bind()
		srcRect := math2.Rect{X: 0.0, Y: 0.0, Width: 1.0, Height: 1.0}
		_ = shaders.UIShader.SetUniformVec4(shaders.UniformSrcRect, srcRect.Vec4())

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
}
