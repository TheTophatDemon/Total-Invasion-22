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

const (
	ColFilterForProjectiles = ColLayerActors | ColLayerMap | ColLayerNPCs | ColLayerPlayers
)

type Projectile struct {
	comps.Body
	world                                          *World
	id                                             scene.Id[*Projectile]
	SpriteRender                                   comps.SpriteRender
	AnimPlayer                                     comps.AnimationPlayer
	StunChance                                     float32 // Probability from 0-1 that this projectile will cause enemies to stun. Multiplied with the enemy's pain chance.
	Damage                                         float32 // Damage done to actors.
	Facing                                         mgl32.Vec3
	owner                                          scene.Handle
	hitOwner                                       bool
	moveFunc                                       func(deltaTime float32)
	onDie                                          func(deltaTime float32)
	forwardSpeed, fallSpeed, maxFallSpeed, gravity float32
	knockbackForce                                 float32
	voices                                         [4]tdaudio.VoiceId
	maxLife, lifeTimer                             float32
	dieAnim                                        textures.Animation
	onCollide                                      func(collidingEntity any, mask collision.Mask, pushVec mgl32.Vec3, deltaTime float32)
}

func (proj *Projectile) Update(deltaTime float32) {
	if lensq := proj.Velocity.LenSqr(); lensq > 0.0 {
		proj.Facing = proj.Velocity.Mul(1.0 / math2.Sqrt(lensq))
	}

	proj.AnimPlayer.Update(deltaTime)
	for _, vid := range proj.voices {
		vid.SetPositionV(proj.Position)
	}

	if proj.moveFunc != nil {
		proj.moveFunc(deltaTime)
	} else if proj.AnimPlayer.IsPlayingAnim(proj.dieAnim) && proj.AnimPlayer.IsAtEnd() {
		proj.world.QueueRemoval(proj.id.Handle)
	}

	proj.Position = proj.Position.Add(proj.Velocity.Mul(deltaTime))

	// Respond to collisions
	if proj.Layer.On(ColLayerProjectiles) && proj.onCollide != nil {
		// Detect intersections with bodies
		bodies := proj.world.bspTree.PotentiallyTouchingEnts(proj.Position, proj.Shape)
		for handle := range bodies {
			collidingEnt, ok := scene.Get[comps.HasBody](handle)
			if !ok {
				continue
			}

			otherBody := collidingEnt.Body()
			if otherBody.Layer == ColLayerNone || !otherBody.OnLayer(ColFilterForProjectiles) {
				continue
			}

			owner, hasOwner := scene.Get[comps.HasBody](proj.owner)
			if !proj.hitOwner && hasOwner && otherBody == owner.Body() {
				continue
			}

			hit, pushVec := proj.Shape.PushOutOf(proj.Position,
				collidingEnt.Body().Position,
				collidingEnt.Body().Shape,
			)
			if hit {
				proj.onCollide(collidingEnt, collidingEnt.Body().Layer, pushVec, deltaTime)
			}
		}

		// Detect intersection with the map
		layers := proj.world.MapLayers.Iter()
		for layer, _ := layers.Next(); layer != nil; layer, _ = layers.Next() {
			if !layer.Layer.On(ColFilterForProjectiles) {
				continue
			}
			pushOut := layer.GridShape.PushOut(mgl32.Vec3{}, proj.Position, proj.Shape)
			if !pushOut.ApproxEqual(mgl32.Vec3{}) {
				proj.onCollide(layer, layer.Layer, pushOut, deltaTime)
			}
		}
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
	proj.SpriteRender.Render(proj.Position, &proj.AnimPlayer, context, 0.0)
}

func (proj *Projectile) moveForward(deltaTime float32) {
	proj.Velocity = proj.Facing.Mul(proj.forwardSpeed)
}

func (proj *Projectile) removeOnDie(deltaTime float32) {
	proj.world.QueueRemoval(proj.id.Handle)
}

func (proj *Projectile) playAnimOnDie(deltaTime float32) {
	if !proj.AnimPlayer.IsPlayingAnim(proj.dieAnim) {
		proj.AnimPlayer.PlayNewAnim(proj.dieAnim)
		proj.Layer.SetBypass()
		proj.Position.Add(proj.Velocity.Mul(-deltaTime))
		proj.Velocity = mgl32.Vec3{}
		proj.moveFunc = nil
	}
}

func (proj *Projectile) dieOnCollide(otherEnt any, mask collision.Mask, pushVec mgl32.Vec3, deltaTime float32) {
	if damageable, canDamage := otherEnt.(Damageable); canDamage {
		damageable.OnDamage(proj, proj.Damage)
	}
	if actorHaver, hasActor := otherEnt.(HasActor); hasActor && proj.knockbackForce != 0.0 && !proj.Velocity.ApproxEqual(mgl32.Vec3{}) {
		actorHaver.Actor().ApplyKnockback(proj.Velocity.Normalize().Mul(proj.knockbackForce))
	}

	proj.lifeTimer = math2.Inf32()
}
