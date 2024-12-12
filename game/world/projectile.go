package world

import (
	"github.com/go-gl/mathgl/mgl32"
	"tophatdemon.com/total-invasion-ii/engine/assets/textures"
	"tophatdemon.com/total-invasion-ii/engine/math2"
	"tophatdemon.com/total-invasion-ii/engine/math2/collision"
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
	onDie                                          func(deltaTime float32)
	forwardSpeed, fallSpeed, maxFallSpeed, gravity float32
	knockbackForce                                 float32
	voices                                         [4]tdaudio.VoiceId
	maxLife, lifeTimer                             float32
	dieAnim                                        textures.Animation
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
	} else if proj.AnimPlayer.IsPlayingAnim(proj.dieAnim) && proj.AnimPlayer.IsAtEnd() {
		proj.world.QueueRemoval(proj.id.Handle)
	}
	proj.lifeTimer += deltaTime
	if (proj.lifeTimer > proj.maxLife && proj.maxLife > 0) || proj.lifeTimer > 10.0 {
		if proj.onDie != nil {
			proj.onDie(deltaTime)
		} else {
			proj.removeOnDie(deltaTime)
		}
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

func (proj *Projectile) moveForward(deltaTime float32) {
	_ = deltaTime
	proj.body.Velocity = mgl32.TransformNormal(mgl32.Vec3{0.0, 0.0, -proj.forwardSpeed}, proj.body.Transform.Matrix())
}

func (proj *Projectile) removeOnDie(deltaTime float32) {
	_ = deltaTime
	proj.world.QueueRemoval(proj.id.Handle)
}

func (proj *Projectile) playAnimOnDie(deltaTime float32) {
	if !proj.AnimPlayer.IsPlayingAnim(proj.dieAnim) {
		proj.AnimPlayer.PlayNewAnim(proj.dieAnim)
		proj.body.Layer = 0
		proj.body.Filter = 0
		proj.body.OnIntersect = nil
		proj.body.Transform.TranslateV(proj.body.Velocity.Mul(-deltaTime))
		proj.body.Velocity = mgl32.Vec3{}
		proj.moveFunc = nil
	}
}

func (proj *Projectile) dieOnHit(otherEnt comps.HasBody, result collision.Result, deltaTime float32) {
	_ = deltaTime
	if !proj.shouldIntersect(otherEnt) {
		return
	}
	if damageable, canDamage := otherEnt.(Damageable); canDamage {
		damageable.OnDamage(proj, proj.Damage)
	}
	if actorHaver, hasActor := otherEnt.(HasActor); hasActor && proj.knockbackForce != 0.0 && !proj.body.Velocity.ApproxEqual(mgl32.Vec3{}) {
		actorHaver.Actor().ApplyKnockback(proj.body.Velocity.Normalize().Mul(proj.knockbackForce))
	}

	proj.lifeTimer = math2.Inf32()
}
