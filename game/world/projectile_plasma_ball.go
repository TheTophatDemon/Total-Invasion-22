package world

import (
	"github.com/go-gl/mathgl/mgl32"
	"tophatdemon.com/total-invasion-ii/engine/assets/cache"
	"tophatdemon.com/total-invasion-ii/engine/math2/collision"
	"tophatdemon.com/total-invasion-ii/engine/scene"
	"tophatdemon.com/total-invasion-ii/engine/scene/comps"
)

func SpawnPlasmaBall(world *World, position, rotation mgl32.Vec3, owner scene.Handle, bigShot bool) (id scene.Id[*Projectile], proj *Projectile, err error) {
	id, proj, err = world.Projectiles.New()
	if err != nil {
		return
	}

	proj.world = world
	proj.id = id
	proj.owner = owner

	proj.body = comps.Body{
		Layer:  COL_LAYER_PROJECTILES,
		Filter: COL_LAYER_NONE,
		LockY:  true,
	}
	if bigShot {
		// NOW'S YOUR CHANCE TO BE A BIG SHOT
		proj.body.Transform = comps.TransformFromTranslationAnglesScale(position, rotation, mgl32.Vec3{0.35, 0.35, 0.35})
		proj.body.Shape = collision.NewSphere(0.35)
		proj.knockbackForce = 5.0
	} else {
		proj.body.Transform = comps.TransformFromTranslationAnglesScale(position, rotation, mgl32.Vec3{0.25, 0.25, 0.25})
		proj.body.Shape = collision.NewSphere(0.25)
		proj.knockbackForce = 10.0
	}

	tex := cache.GetTexture("assets/textures/sprites/plasma_ball.png")
	proj.SpriteRender = comps.NewSpriteRender(tex)
	proj.AnimPlayer = comps.NewAnimationPlayer(tex.GetDefaultAnimation(), true)
	proj.dieAnim, _ = tex.GetAnimation("die")
	proj.forwardSpeed = 120.0
	proj.StunChance = 0.1
	proj.Damage = 5

	proj.moveFunc = proj.moveForward
	proj.body.OnIntersect = proj.dieOnHit
	proj.onDie = proj.playAnimOnDie

	return
}
