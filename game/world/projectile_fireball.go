package world

import (
	"github.com/go-gl/mathgl/mgl32"
	"tophatdemon.com/total-invasion-ii/engine/assets/cache"
	"tophatdemon.com/total-invasion-ii/engine/math2/collision"
	"tophatdemon.com/total-invasion-ii/engine/scene"
	"tophatdemon.com/total-invasion-ii/engine/scene/comps"
)

const (
	SFX_FIREBALL = "assets/sounds/fireball.wav"
)

func SpawnFireball(world *World, position, rotation mgl32.Vec3, owner scene.Handle) (id scene.Id[*Projectile], proj *Projectile, err error) {
	id, proj, err = world.Projectiles.New()
	if err != nil {
		return
	}

	proj.world = world
	proj.id = id
	proj.owner = owner

	proj.body = comps.Body{
		Transform: comps.TransformFromTranslationAnglesScale(position, rotation, mgl32.Vec3{0.375, 0.375, 0.375}),
		Shape:     collision.NewSphere(0.25),
		Layer:     COL_LAYER_PROJECTILES,
		Filter:    COL_LAYER_NONE,
		LockY:     true,
	}

	tex := cache.GetTexture("assets/textures/sprites/fireball.png")
	proj.SpriteRender = comps.NewSpriteRender(tex)
	proj.AnimPlayer = comps.NewAnimationPlayer(tex.GetDefaultAnimation(), true)
	proj.forwardSpeed = 70.0
	proj.voices[0] = cache.GetSfx(SFX_FIREBALL).PlayAttenuatedV(position)
	proj.StunChance = 0.1
	proj.Damage = 15

	proj.moveFunc = proj.moveForward
	proj.body.OnIntersect = proj.dieOnHit

	return
}
