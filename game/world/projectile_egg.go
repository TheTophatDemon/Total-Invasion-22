package world

import (
	"math/rand"
	"time"

	"github.com/go-gl/mathgl/mgl32"
	"tophatdemon.com/total-invasion-ii/engine/assets/cache"
	"tophatdemon.com/total-invasion-ii/engine/math2"
	"tophatdemon.com/total-invasion-ii/engine/math2/collision"
	"tophatdemon.com/total-invasion-ii/engine/scene"
	"tophatdemon.com/total-invasion-ii/engine/scene/comps"
	"tophatdemon.com/total-invasion-ii/game/world/effects"
)

const (
	CHICKEN_SPAWN_CHANCE = 0.3
)

var timeSinceLastChicken time.Time

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
	proj.forwardSpeed = 100.0
	proj.StunChance = 0.1
	proj.Damage = 15

	proj.moveFunc = proj.moveForward
	proj.body.OnIntersect = proj.eggIntersect

	return
}

func (proj *Projectile) moveForward(deltaTime float32) {
	proj.body.Velocity = mgl32.TransformNormal(mgl32.Vec3{0.0, 0.0, -proj.forwardSpeed}, proj.body.Transform.Matrix())
}

func (proj *Projectile) eggIntersect(otherEnt comps.HasBody, result collision.Result, deltaTime float32) {
	if !proj.shouldIntersect(otherEnt) {
		return
	}

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
	noBlockers := len(proj.world.BodiesInSphere(chickenSpot, 0.5, proj)) == 0
	if rand.Float32() < CHICKEN_SPAWN_CHANCE && noBlockers && time.Now().Sub(timeSinceLastChicken).Seconds() > 10.0 {
		SpawnChicken(proj.world, chickenSpot, mgl32.Vec3{0.0, mgl32.RadToDeg(math2.Atan2(-backwards[0], backwards[2])), 0.0})
		timeSinceLastChicken = time.Now()
	}
}
