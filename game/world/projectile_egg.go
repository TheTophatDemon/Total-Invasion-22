package world

import (
	"github.com/go-gl/mathgl/mgl32"
	"tophatdemon.com/total-invasion-ii/engine/assets/cache"
	"tophatdemon.com/total-invasion-ii/engine/math2/collision"
	"tophatdemon.com/total-invasion-ii/engine/scene"
	"tophatdemon.com/total-invasion-ii/engine/scene/comps"
)

const (
	SFX_EGG_SHOOT = "assets/sounds/chickengun.wav"
)

func SpawnEgg(world *World, st *scene.Storage[Projectile], position, rotation mgl32.Vec3, owner scene.Handle) (id scene.Id[*Projectile], proj *Projectile, err error) {
	id, proj, err = st.New()
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
	proj.voices[0] = cache.GetSfx(SFX_EGG_SHOOT).Play()
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
	otherBody := otherEnt.Body()
	owner, hasOwner := scene.Get[comps.HasBody](proj.owner)
	if !hasOwner || (hasOwner && otherBody != owner.Body()) {
		if damageable, canDamage := otherEnt.(Damageable); canDamage {
			damageable.OnDamage(proj, proj.Damage)
		}
		proj.world.QueueRemoval(proj.id.Handle)
	}
}
