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
	proj.maxLife = 1.5
	proj.onDie = proj.explodeOnDie

	proj.body = comps.Body{
		Transform:   comps.TransformFromTranslationAnglesScale(position, mgl32.Vec3{}, mgl32.Vec3{0.25, 0.25, 0.25}),
		Shape:       collision.NewContinuousSphere(0.25),
		Velocity:    direction.Mul(20.0),
		Layer:       COL_LAYER_PROJECTILES,
		Filter:      COL_LAYER_MAP,
		LockY:       false,
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
		proj.explodeOnDie()
		proj.world.QueueRemoval(proj.id.Handle)
	} else if otherEnt.Body().OnLayer(COL_LAYER_MAP) {
		horzVelocity := math2.Vec3WithY(proj.body.Velocity, 0.0)
		speed := horzVelocity.Len() * 0.8
		if collision.Normal.Y() > 0.1 && proj.fallSpeed < 0.0 {
			if proj.fallSpeed > -0.01 {
				proj.fallSpeed = 0.0
			}
			proj.fallSpeed = -proj.fallSpeed * 0.9
			proj.body.Velocity = math2.Vec3WithY(horzVelocity.Normalize().Mul(speed), proj.fallSpeed)
		} else {
			if speed > 0.01 {
				reflection := math2.Vec3Reflect(horzVelocity.Normalize(), collision.Normal)
				proj.body.Velocity = mgl32.Vec3{reflection.X() * speed, 0.0, reflection.Z() * speed}
			} else {
				proj.body.Velocity = mgl32.Vec3{}
			}
		}
	}
}

func (proj *Projectile) explodeOnDie() {
	proj.body.Transform.Translate(0.0, 0.5, 0.0)
	SpawnSingleExplosion(proj.world, proj.body.Transform)
}
