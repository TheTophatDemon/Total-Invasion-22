package world

import (
	"math/rand"

	"github.com/go-gl/mathgl/mgl32"
	"tophatdemon.com/total-invasion-ii/engine/assets/cache"
	"tophatdemon.com/total-invasion-ii/engine/color"
	"tophatdemon.com/total-invasion-ii/engine/math2"
	"tophatdemon.com/total-invasion-ii/engine/math2/collision"
	"tophatdemon.com/total-invasion-ii/engine/scene"
	"tophatdemon.com/total-invasion-ii/engine/scene/comps"
)

const (
	SFX_EGG_SHOOT  = "assets/sounds/chickengun.wav"
	TEX_EGG_SHARDS = "assets/textures/sprites/egg_shards.png"
)

const (
	CHICKEN_SPAWN_CHANCE = 0.1
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
		proj.emitEggShards(proj.body.Transform.Position().Add(backwards.Mul(1.0)))
		if rand.Float32() < CHICKEN_SPAWN_CHANCE {
			SpawnChicken(proj.world,
				proj.body.Transform.Position().Add(backwards),
				mgl32.Vec3{0.0, mgl32.RadToDeg(math2.Atan2(-backwards[0], backwards[2])), 0.0})
		}
	}
}

func (proj *Projectile) emitEggShards(position mgl32.Vec3) {
	particleTex := cache.GetTexture(TEX_EGG_SHARDS)
	SpawnEffect(proj.world, &proj.world.Effects, comps.TransformFromTranslation(position), 1.0, comps.ParticleRender{
		Texture:       particleTex,
		EmissionTimer: 0.2,
		MaxCount:      4,
		SpawnRadius:   proj.body.Shape.(collision.Sphere).Radius(),
		SpawnRate:     1.0,
		SpawnCount:    6,
		SpawnFunc: func(index int, form *comps.ParticleForm, info *comps.ParticleInfo) {
			form.Color = color.White.Vector()
			form.Size = mgl32.Vec2{0.2, 0.2}
			info.Velocity = info.Velocity.Mul(rand.Float32()*1 + 5.0)
			info.Acceleration = mgl32.Vec3{0.0, -20.0, 0.0}
			info.Lifetime = 1.0
			info.AnimPlayer = comps.NewAnimationPlayer(particleTex.GetDefaultAnimation(), false)
			info.AnimPlayer.MoveToRandomFrame()
		},
		UpdateFunc: func(deltaTime float32, form *comps.ParticleForm, info *comps.ParticleInfo) {
			const SHRINK_RATE = 0.75
			form.Size[0] -= deltaTime * SHRINK_RATE
			form.Size[1] -= deltaTime * SHRINK_RATE
			if form.Size[0] <= 0.1 {
				form.Size = mgl32.Vec2{}
				info.Lifetime = 0.0
			}
		},
	})
}
