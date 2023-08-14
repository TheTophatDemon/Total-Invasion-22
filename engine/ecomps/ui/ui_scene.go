package ui

import (
	"log"

	"github.com/go-gl/gl/v3.3-core/gl"
	"github.com/go-gl/mathgl/mgl32"
	"tophatdemon.com/total-invasion-ii/engine/assets"
	"tophatdemon.com/total-invasion-ii/engine/ecomps"
	"tophatdemon.com/total-invasion-ii/engine/ecs"
	"tophatdemon.com/total-invasion-ii/engine/math2"
	"tophatdemon.com/total-invasion-ii/engine/render"
)

type Scene struct {
	ecs.Scene
	Boxes            *ecs.ComponentStorage[Box]
	Texts            *ecs.ComponentStorage[Text]
	AnimationPlayers *ecs.ComponentStorage[ecomps.AnimationPlayer]
}

func NewUIScene(maxEnts uint) Scene {
	return Scene{
		ecs.NewScene(maxEnts),
		ecs.NewStorage[Box](maxEnts),
		ecs.NewStorage[Text](maxEnts),
		ecs.NewStorage[ecomps.AnimationPlayer](maxEnts),
	}
}

func (scene *Scene) UpdateAll(deltaTime float32) {
	for iter := scene.EntsIter(); iter.Valid(); iter = iter.Next() {
		animPlayer, hasAnim := scene.AnimationPlayers.Get(iter.Entity())
		if hasAnim {
			animPlayer.Update(deltaTime)
		}
	}
}

func (scene *Scene) RenderAll(context *render.Context) {
	gl.Clear(gl.DEPTH_BUFFER_BIT)
	gl.CullFace(gl.FRONT)
	// gl.Disable(gl.CULL_FACE)

	// Render boxes in one large batch
	assets.QuadMesh.Bind()
	assets.UIShader.Use()

	// Box uniforms
	_ = context.SetUniforms(assets.UIShader)
	_ = assets.UIShader.SetUniformInt(assets.UniformTex, 0)
	_ = assets.UIShader.SetUniformInt(assets.UniformAtlas, 1)
	_ = assets.UIShader.SetUniformInt(assets.UniformFrame, 0)

	for iter := scene.EntsIter(); iter.Valid(); iter = iter.Next() {
		box, hasBox := scene.Boxes.Get(iter.Entity())
		animPlayer, hasAnim := scene.AnimationPlayers.Get(iter.Entity())

		if hasAnim {
			_ = assets.UIShader.SetUniformInt(assets.UniformFrame, animPlayer.Frame())
		}

		if hasBox {
			// Set color
			_ = assets.UIShader.SetUniformVec4(assets.UniformDiffuseColor, math2.ColorToVec4(box.Color))

			// Set texture
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
		}
	}

	_ = assets.UIShader.SetUniformInt(assets.UniformFrame, 0)

	// Render other UI elements
	for iter := scene.EntsIter(); iter.Valid(); iter = iter.Next() {
		text, hasText := scene.Texts.Get(iter.Entity())
		animPlayer, hasAnim := scene.AnimationPlayers.Get(iter.Entity())

		if hasAnim {
			_ = assets.UIShader.SetUniformInt(assets.UniformFrame, animPlayer.Frame())
		}

		if hasText {
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
		}
	}
}
