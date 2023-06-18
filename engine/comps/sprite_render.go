package comps

import (
	"tophatdemon.com/total-invasion-ii/engine/assets"
	"tophatdemon.com/total-invasion-ii/engine/scene"
)

type SpriteRender struct {
	*MeshRender
	atlas *assets.AtlasTexture
	anim  *AnimationPlayer
}

func NewSpriteRender(atlas *assets.AtlasTexture) *SpriteRender {
	return &SpriteRender{
		&MeshRender{
			mesh:   assets.SpriteMesh,
			shader: assets.SpriteShader,
		},
		atlas,
		NewAnimationPlayer(atlas.GetAnimation(0)),
	}
}

func (sr *SpriteRender) UpdateComponent(sc *scene.Scene, ent scene.Entity, deltaTime float32) {
	sr.anim.Update(deltaTime)
}

func (sr *SpriteRender) RenderComponent(sc *scene.Scene, ent scene.Entity) {
	// TODO: Set sprite specific uniforms
	sr.MeshRender.RenderComponent(sc, ent)
}

func (sr *SpriteRender) PrepareRender() {
	sr.MeshRender.PrepareRender()
	// TODO: Set sprite shader specific uniforms
}
