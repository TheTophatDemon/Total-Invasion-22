package comps

import (
	"tophatdemon.com/total-invasion-ii/engine/assets"
	"tophatdemon.com/total-invasion-ii/engine/scene"
)

type SpriteRender struct {
	*MeshRender
	atlas *assets.Texture
	anim  *AnimationPlayer
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
	if sr.anim != nil {
		sr.anim.Update(deltaTime)
	}
}

func (sr *SpriteRender) RenderComponent(sc *scene.Scene, ent scene.Entity) {
	// TODO: Set sprite specific uniforms
	sr.MeshRender.RenderComponent(sc, ent)
}

func (sr *SpriteRender) PrepareRender() {
	sr.MeshRender.PrepareRender()
	// TODO: Set sprite shader specific uniforms
}
