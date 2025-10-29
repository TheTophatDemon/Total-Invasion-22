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
	body                                           comps.Body
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
	onCollide                                      func(
		movement mgl32.Vec3,
		collidingEntity any,
		mask collision.Mask,
		result collision.Result,
		deltaTime float32,
	) mgl32.Vec3
}

func (proj *Projectile) Body() *comps.Body {
	return &proj.body
}

func (proj *Projectile) Update(deltaTime float32) {
	if lensq := proj.body.Velocity.LenSqr(); lensq > 0.0 {
		proj.Facing = proj.body.Velocity.Mul(1.0 / math2.Sqrt(lensq))
	}

	proj.AnimPlayer.Update(deltaTime)
	for _, vid := range proj.voices {
		vid.SetPositionV(proj.body.Position)
	}

	if proj.moveFunc != nil {
		proj.moveFunc(deltaTime)
	} else if proj.AnimPlayer.IsPlayingAnim(proj.dieAnim) && proj.AnimPlayer.IsAtEnd() {
		gWorld.QueueRemoval(proj.id.Handle)
	}

	movement := proj.body.Velocity.Mul(deltaTime)

	// Respond to collisions
	if proj.onCollide != nil {
		minRes := collision.Result{Distance: math2.Inf32()}
		var minEnt any
		var minLayers collision.Mask

		// Detect intersections with bodies
		bodies := gWorld.bspTree.PotentiallyTouchingEnts(proj.body.Position, proj.body.Shape)
		for handle := range bodies {
			collidingEnt, ok := scene.Get[comps.HasBody](handle)
			if !ok {
				continue
			}

			otherBody := collidingEnt.Body()
			if !otherBody.OnLayer(ColFilterForProjectiles) {
				continue
			}

			owner, hasOwner := scene.Get[comps.HasBody](proj.owner)
			if !proj.hitOwner && hasOwner && otherBody == owner.Body() {
				continue
			}

			res := proj.body.Shape.Sweep(
				proj.body.Position, movement,
				collidingEnt.Body().Position,
				collidingEnt.Body().Shape,
			)
			if res.Hit && res.Distance < minRes.Distance {
				minRes = res
				minEnt = collidingEnt
				minLayers = collidingEnt.Body().Layers
			}
		}

		// Detect intersection with the map
		mapRes := gWorld.GameMap.GridShape.SweepAgainst(mgl32.Vec3{}, proj.body.Position, movement, proj.body.Shape)
		if mapRes.Hit && mapRes.Distance < minRes.Distance {
			minRes = mapRes
			minEnt = gWorld.GameMap
			minLayers = gWorld.GameMap.Layer
		}

		if minRes.Hit {
			movement = proj.onCollide(movement, minEnt, minLayers, minRes, deltaTime)
		}
	}

	proj.body.TranslateV(movement)

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
	proj.SpriteRender.Render(proj.body.Position, &proj.AnimPlayer, context, 0.0)
}

func (proj *Projectile) moveForward(deltaTime float32) {
	proj.body.Velocity = proj.Facing.Mul(proj.forwardSpeed)
}

func (proj *Projectile) removeOnDie(deltaTime float32) {
	_ = deltaTime
	gWorld.QueueRemoval(proj.id.Handle)
}

func (proj *Projectile) playAnimOnDie(deltaTime float32) {
	if !proj.AnimPlayer.IsPlayingAnim(proj.dieAnim) {
		proj.AnimPlayer.PlayNewAnim(proj.dieAnim)
		proj.body.Layers = 0
		proj.onCollide = nil
		proj.moveFunc = nil
		proj.body.Position.Add(proj.body.Velocity.Mul(-deltaTime))
		proj.body.Velocity = mgl32.Vec3{}
	}
}

func (proj *Projectile) stopOnCollide(
	movement mgl32.Vec3,
	collidingEntity any,
	mask collision.Mask,
	result collision.Result,
	deltaTime float32,
) mgl32.Vec3 {
	_ = collidingEntity
	_ = mask
	_ = deltaTime
	if movement.LenSqr() > 0.0 {
		// Move to just before the point of collision
		proj.body.Position = proj.body.Position.Add(movement.Normalize().Mul(result.Distance - 0.1))
		movement = mgl32.Vec3{}
	}
	return movement
}

func (proj *Projectile) dieOnCollide(movement mgl32.Vec3, otherEnt any, mask collision.Mask, result collision.Result, deltaTime float32) mgl32.Vec3 {
	if damageable, canDamage := otherEnt.(Damageable); canDamage {
		damageable.OnDamage(proj, proj.Damage)
	}
	if actorHaver, hasActor := otherEnt.(HasActor); hasActor && proj.knockbackForce != 0.0 && !proj.body.Velocity.ApproxEqual(mgl32.Vec3{}) {
		actorHaver.Actor().ApplyKnockback(proj.body.Velocity.Normalize().Mul(proj.knockbackForce))
	}

	proj.lifeTimer = math2.Inf32()
	return mgl32.Vec3{}
}
