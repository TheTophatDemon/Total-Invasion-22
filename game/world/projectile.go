package world

import (
	"tophatdemon.com/total-invasion-ii/engine/assets/audio"
	"tophatdemon.com/total-invasion-ii/engine/render"
	"tophatdemon.com/total-invasion-ii/engine/scene"
	"tophatdemon.com/total-invasion-ii/engine/scene/comps"
)

type Projectile struct {
	world        *World
	id           scene.Id[*Projectile]
	SpriteRender comps.SpriteRender
	AnimPlayer   comps.AnimationPlayer
	StunChance   float32 // Probability from 0-1 that this projectile will cause enemies to stun. Multiplied with the enemy's pain chance.
	body         comps.Body
	owner        scene.Handle
	moveFunc     func(deltaTime float32)
	speed        float32
	voices       [4]audio.VoiceId
}

var _ comps.HasBody = (*Projectile)(nil)

func (proj *Projectile) Body() *comps.Body {
	return &proj.body
}

func (proj *Projectile) Update(deltaTime float32) {
	proj.AnimPlayer.Update(deltaTime)
	for _, vid := range proj.voices {
		vid.Attenuate(proj.Body().Transform.Position(), proj.world.ListenerTransform())
	}
	if proj.moveFunc != nil {
		proj.moveFunc(deltaTime)
	}
}

func (proj *Projectile) Render(context *render.Context) {
	proj.SpriteRender.Render(&proj.body.Transform, &proj.AnimPlayer, context, proj.body.Transform.Yaw())
}
