package world

import (
	"tophatdemon.com/total-invasion-ii/engine/render"
	"tophatdemon.com/total-invasion-ii/engine/scene"
	"tophatdemon.com/total-invasion-ii/engine/scene/comps"
	"tophatdemon.com/total-invasion-ii/engine/tdaudio"
)

type Projectile struct {
	world                                          *World
	id                                             scene.Id[*Projectile]
	SpriteRender                                   comps.SpriteRender
	AnimPlayer                                     comps.AnimationPlayer
	StunChance                                     float32 // Probability from 0-1 that this projectile will cause enemies to stun. Multiplied with the enemy's pain chance.
	Damage                                         float32 // Damage done to actors.
	body                                           comps.Body
	owner                                          scene.Handle
	moveFunc                                       func(deltaTime float32)
	forwardSpeed, fallSpeed, maxFallSpeed, gravity float32
	voices                                         [4]tdaudio.VoiceId
}

var _ comps.HasBody = (*Projectile)(nil)

func (proj *Projectile) Body() *comps.Body {
	return &proj.body
}

func (proj *Projectile) Update(deltaTime float32) {
	proj.AnimPlayer.Update(deltaTime)
	for _, vid := range proj.voices {
		vid.SetPositionV(proj.Body().Transform.Position())
	}
	if proj.moveFunc != nil {
		proj.moveFunc(deltaTime)
	}
}

func (proj *Projectile) Render(context *render.Context) {
	proj.SpriteRender.Render(&proj.body.Transform, &proj.AnimPlayer, context, proj.body.Transform.Yaw())
}

func (proj *Projectile) shouldIntersect(otherEnt comps.HasBody) bool {
	if !proj.body.OnLayer(COL_LAYER_PROJECTILES) {
		return false
	}
	otherBody := otherEnt.Body()
	if otherBody.Layer == COL_LAYER_NONE || otherBody.OnLayer(COL_LAYER_INVISIBLE|COL_LAYER_PROJECTILES) {
		return false
	}
	owner, hasOwner := scene.Get[comps.HasBody](proj.owner)
	if !hasOwner || (hasOwner && otherBody != owner.Body()) {
		return true
	}
	return false
}
