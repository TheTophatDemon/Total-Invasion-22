package world

import (
	"math/rand"

	"github.com/go-gl/mathgl/mgl32"
	"tophatdemon.com/total-invasion-ii/engine/assets/cache"
	"tophatdemon.com/total-invasion-ii/engine/math2"
	"tophatdemon.com/total-invasion-ii/engine/math2/collision"
	"tophatdemon.com/total-invasion-ii/engine/scene"
	"tophatdemon.com/total-invasion-ii/engine/scene/comps"
	"tophatdemon.com/total-invasion-ii/game/world/effects"
)

const (
	SFX_EGG_SHOOT = "assets/sounds/chickengun.wav"
)

const (
	CHICKEN_SPAWN_CHANCE = 0.1
)

func SpawnEgg(world *World, position, rotation mgl32.Vec3, owner scene.Handle) (id scene.Id[*Projectile], proj *Projectile, err error) {
	id, proj, err = world.Projectiles.New()
	if err != nil {
		return
	}

	proj.world = world
	proj.id = id
	proj.owner = owner

	proj.body = comps.Body{
		Transform: comps.TransformFromTranslationAnglesScale(position, rotation, mgl32.Vec3{0.4, 0.4, 0.4}),
		Shape:     collision.NewSphere(0.1),
		Layer:     COL_LAYER_PROJECTILES,
		Filter:    COL_LAYER_NONE,
		LockY:     true,
	}

	eggTex := cache.GetTexture("assets/textures/sprites/egg.png")
	proj.SpriteRender = comps.NewSpriteRender(eggTex)
	proj.speed = 100.0
	proj.voices[0] = cache.GetSfx(SFX_EGG_SHOOT).PlayAttenuatedV(position)
	proj.StunChance = 0.1
	proj.Damage = 15

	proj.moveFunc = proj.eggMove
	proj.body.OnIntersect = proj.eggIntersect

	return
}

func (proj *Projectile) eggMove(deltaTime float32) {
	proj.body.Velocity = mgl32.TransformNormal(mgl32.Vec3{0.0, 0.0, -proj.speed}, proj.body.Transform.Matrix())
}

func (proj *Projectile) eggIntersect(otherEnt comps.HasBody, result collision.Result, deltaTime float32) {
	if !proj.body.OnLayer(COL_LAYER_PROJECTILES) {
		return
	}
	otherBody := otherEnt.Body()
	if otherBody.Layer == COL_LAYER_NONE || otherBody.Layer == COL_LAYER_INVISIBLE {
		return
	}
	owner, hasOwner := scene.Get[comps.HasBody](proj.owner)
	if !hasOwner || (hasOwner && otherBody != owner.Body()) {
		if damageable, canDamage := otherEnt.(Damageable); canDamage {
			damageable.OnDamage(proj, proj.Damage)
		}

		proj.world.QueueRemoval(proj.id.Handle)
		proj.body.Layer = 0
		proj.body.Filter = 0
		var backwards mgl32.Vec3
		if proj.body.Velocity.LenSqr() != 0.0 {
			backwards = proj.body.Velocity.Normalize().Mul(-1.0)
		}
		SpawnEffect(proj.world,
			comps.TransformFromTranslation(proj.body.Transform.Position().Add(backwards)),
			1.0,
			effects.EggShards(proj.body.Shape.(collision.Sphere).Radius()))

		chickenSpot := proj.body.Transform.Position().Add(backwards.Mul(1.5))
		if rand.Float32() < CHICKEN_SPAWN_CHANCE && len(proj.world.BodiesInSphere(chickenSpot, 0.5, proj)) == 0 {
			SpawnChicken(proj.world, chickenSpot, mgl32.Vec3{0.0, mgl32.RadToDeg(math2.Atan2(-backwards[0], backwards[2])), 0.0})
		}
	}
}
