package comps

import (
	"github.com/go-gl/gl/v3.3-core/gl"
	"tophatdemon.com/total-invasion-ii/engine/assets"
	"tophatdemon.com/total-invasion-ii/engine/math2"
	"tophatdemon.com/total-invasion-ii/engine/scene"
)

type SpriteRender struct {
	*MeshRender
	atlas *assets.Texture
	Anim  *AnimationPlayer
}

func NewSpriteRender(atlas *assets.Texture) *SpriteRender {
	// If the texture isn't actually an atlas, then the animation player will be nil.
	var animPlayer *AnimationPlayer = nil
	if atlas.IsAtlas() {
		animPlayer = NewAnimationPlayer(atlas.GetAnimation(0))
	}

	return &SpriteRender{
		&MeshRender{
			mesh:   assets.SpriteMesh,
			shader: assets.SpriteShader,
		},
		atlas,
		animPlayer,
	}
}

func (sr *SpriteRender) UpdateComponent(sc *scene.Scene, ent scene.Entity, deltaTime float32) {
	if sr.Anim != nil {
		sr.Anim.Update(deltaTime)
	}
}

func (sr *SpriteRender) RenderComponent(sc *scene.Scene, ent scene.Entity, ctx *scene.RenderContext) {
	var transform *Transform
	transform, err := scene.GetComponent(sc, ent, transform)
	if err != nil {
		transform = TransformFromTranslation(math2.Vec3Zero())
	}
	modelMatrix := transform.GetMatrix()
	gl.UniformMatrix4fv(sr.shader.GetUniformLoc("uModelTransform"), 1, false, &modelMatrix[0])

	gl.UniformMatrix4fv(sr.shader.GetUniformLoc("uViewTransform"), 1, false, &ctx.View[0])
	gl.UniformMatrix4fv(sr.shader.GetUniformLoc("uProjectionTransform"), 1, false, &ctx.Projection[0])
	if sr.atlas.IsAtlas() {
		gl.Uniform1i(sr.shader.GetUniformLoc("uAtlasUsed"), 1)
		frame := 0
		if sr.Anim != nil {
			frame = sr.Anim.Frame()
		}
		gl.Uniform1i(sr.shader.GetUniformLoc("uFrame"), int32(frame))
		gl.ActiveTexture(gl.TEXTURE1)
	} else {
		gl.Uniform1i(sr.shader.GetUniformLoc("uAtlasUsed"), 0)
		gl.ActiveTexture(gl.TEXTURE0)
	}
	gl.BindTexture(sr.atlas.Target(), sr.atlas.ID())
	sr.MeshRender.RenderComponent(sc, ent, ctx)
}

func (sr *SpriteRender) PrepareRender(ctx *scene.RenderContext) {
	sr.MeshRender.PrepareRender(ctx)
	gl.Uniform1f(sr.shader.GetUniformLoc("uFogStart"), ctx.FogStart)
	gl.Uniform1f(sr.shader.GetUniformLoc("uFogLength"), ctx.FogLength)
	gl.Uniform1i(sr.shader.GetUniformLoc("uTex"), 0)
	gl.Uniform1i(sr.shader.GetUniformLoc("uAtlas"), 1)
}
