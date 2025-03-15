package world

import (
	"github.com/go-gl/mathgl/mgl32"
	"tophatdemon.com/total-invasion-ii/engine/assets/cache"
	"tophatdemon.com/total-invasion-ii/engine/math2/collision"
	"tophatdemon.com/total-invasion-ii/engine/scene"
	"tophatdemon.com/total-invasion-ii/engine/scene/comps"
)

func SpawnBlessing(world *World, position, rotation mgl32.Vec3, owner scene.Handle) (id scene.Id[*Projectile], proj *Projectile, err error) {
	id, proj, err = world.Projectiles.New()
	if err != nil {
		return
	}

	proj.world = world
	proj.id = id
	proj.owner = owner

	proj.body = comps.Body{
		Transform: comps.TransformFromTranslationAnglesScale(position, rotation, mgl32.Vec3{0.5, 0.5, 0.5}),
		Shape:     collision.NewSphere(0.5),
		Layer:     COL_LAYER_PROJECTILES,
		Filter:    COL_LAYER_NONE,
		LockY:     true,
	}

	tex := cache.GetTexture("assets/textures/sprites/blessing.png")
	proj.SpriteRender = comps.NewSpriteRender(tex)
	proj.AnimPlayer = comps.NewAnimationPlayer(tex.GetDefaultAnimation(), true)
	proj.forwardSpeed = 20.0
	proj.voices[0] = cache.GetSfx(SFX_FIREBALL).PlayAttenuatedV(position)
	proj.StunChance = 0.0
	proj.Damage = 15

	proj.moveFunc = proj.moveForwardAndRevive
	proj.body.OnIntersect = proj.blessingOnHit

	return
}

func (proj *Projectile) moveForwardAndRevive(deltaTime float32) {
	proj.moveForward(deltaTime)

	enemiesIter := proj.world.Enemies.Iter()
	for {
		enemy, handle := enemiesIter.Next()
		if enemy == nil || handle.Equals(proj.owner) {
			break
		}
		if enemy.actor.Health <= 0.0 {
			diff := enemy.Body().Transform.Position().Sub(proj.body.Transform.Position())
			dist := diff.Len()
			if dist < 2.0 {
				// Ensure we are not reviving an enemy from behind a wall.
				rayHit, _ := proj.world.Raycast(proj.body.Transform.Position(), diff.Mul(1.0/dist), COL_LAYER_MAP, dist, nil)
				if !rayHit.Hit {
					enemy.changeState(&enemy.reviveState)
				}
			}
		}
	}
}

func (proj *Projectile) blessingOnHit(collidingEntity comps.HasBody, collision collision.Result, deltaTime float32) {
	if _, isEnemy := collidingEntity.(*Enemy); !isEnemy {
		proj.dieOnHit(collidingEntity, collision, deltaTime)
	}
}
