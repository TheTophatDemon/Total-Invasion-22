package ui

import (
	"github.com/go-gl/gl/v3.3-core/gl"
	"github.com/go-gl/mathgl/mgl32"
	"tophatdemon.com/total-invasion-ii/engine/assets"
	"tophatdemon.com/total-invasion-ii/engine/ecomps"
	"tophatdemon.com/total-invasion-ii/engine/ecs"
	"tophatdemon.com/total-invasion-ii/engine/render"
)

type Scene struct {
	ecs.Scene
	Boxes            *ecs.ComponentStorage[Box]
	AnimationPlayers *ecs.ComponentStorage[ecomps.AnimationPlayer]
}

func NewUIScene(maxEnts uint) Scene {
	return Scene{
		ecs.NewScene(maxEnts),
		ecs.NewStorage[Box](maxEnts),
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

	// Render boxes
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
			r, g, b, a := box.Color.RGBA()
			colorVec := mgl32.Vec4{
				float32(r) / float32(0xffff),
				float32(g) / float32(0xffff),
				float32(b) / float32(0xffff),
				float32(a) / float32(0xffff),
			}
			_ = assets.UIShader.SetUniformVec4(assets.UniformDiffuseColor, colorVec)

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
}
