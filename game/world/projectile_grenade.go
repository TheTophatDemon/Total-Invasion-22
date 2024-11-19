package world

import (
	"github.com/go-gl/mathgl/mgl32"
	"tophatdemon.com/total-invasion-ii/engine/assets/cache"
	"tophatdemon.com/total-invasion-ii/engine/math2"
	"tophatdemon.com/total-invasion-ii/engine/math2/collision"
	"tophatdemon.com/total-invasion-ii/engine/scene"
	"tophatdemon.com/total-invasion-ii/engine/scene/comps"
)

const (
	SFX_GRENADE = "assets/sounds/grenadelaunch.wav"
)

func SpawnGrenade(world *World, position, direction mgl32.Vec3, owner scene.Handle) (id scene.Id[*Projectile], proj *Projectile, err error) {
	id, proj, err = world.Projectiles.New()
	if err != nil {
		return
	}

	proj.world = world
	proj.id = id
	proj.owner = owner

	tex := cache.GetTexture("assets/textures/sprites/grenade.png")
	proj.SpriteRender = comps.NewSpriteRender(tex)
	proj.AnimPlayer = comps.NewAnimationPlayer(tex.GetDefaultAnimation(), true)
	proj.voices[0] = cache.GetSfx(SFX_GRENADE).PlayAttenuatedV(position)
	proj.StunChance = 0.0
	proj.Damage = 15
	proj.gravity = -25.0
	proj.maxFallSpeed = 10.0
	proj.moveFunc = proj.applyGravity

	proj.body = comps.Body{
		Transform:   comps.TransformFromTranslationAnglesScale(position, mgl32.Vec3{}, mgl32.Vec3{0.25, 0.25, 0.25}),
		Shape:       collision.NewSphere(0.25),
		Velocity:    direction.Mul(20.0),
		Layer:       COL_LAYER_PROJECTILES,
		Filter:      COL_LAYER_MAP,
		LockY:       true,
		OnIntersect: proj.grenadeHit,
	}
	return
}

func (proj *Projectile) applyGravity(deltaTime float32) {
	proj.fallSpeed = max(-proj.maxFallSpeed, proj.fallSpeed+(proj.gravity*deltaTime))
	proj.body.Velocity = math2.Vec3WithY(proj.body.Velocity, proj.fallSpeed)
}

func (proj *Projectile) grenadeHit(otherEnt comps.HasBody, collision collision.Result, deltaTime float32) {
	_ = deltaTime
	if !proj.shouldIntersect(otherEnt) {
		return
	}
	if damageable, canDamage := otherEnt.(Damageable); canDamage {
		damageable.OnDamage(proj, proj.Damage)
		proj.world.QueueRemoval(proj.id.Handle)
	} else if otherEnt.Body().OnLayer(COL_LAYER_MAP) {
		// if collision.Normal.Y() > 0.1 && proj.fallSpeed < 0.0 {
		// 	if proj.fallSpeed > -0.01 {
		// 		proj.fallSpeed = 0.0
		// 	}
		// 	proj.fallSpeed = -proj.fallSpeed * 0.9
		// } else if math2.Abs(collision.Normal.X())+math2.Abs(collision.Normal.Y()) > 0.5 {
		// 	speed := proj.body.Velocity.Len() * 0.8
		// 	reflection := math2.Vec3Reflect(proj.body.Velocity.Normalize(), collision.Normal)
		// 	proj.body.Velocity = mgl32.Vec3{reflection.X() * speed, proj.body.Velocity.Y(), reflection.Z() * speed}
		// }
		proj.body.Velocity = mgl32.Vec3{}
	}
}
