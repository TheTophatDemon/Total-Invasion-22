package world

import (
	"log"

	"github.com/go-gl/mathgl/mgl32"
	"tophatdemon.com/total-invasion-ii/engine/assets/cache"
	"tophatdemon.com/total-invasion-ii/engine/input"
	"tophatdemon.com/total-invasion-ii/engine/math2"
	"tophatdemon.com/total-invasion-ii/engine/math2/collision"
	"tophatdemon.com/total-invasion-ii/engine/scene"
	"tophatdemon.com/total-invasion-ii/engine/scene/comps"
	"tophatdemon.com/total-invasion-ii/game/settings"
)

const (
	SFX_SICKLE_THROW = "assets/sounds/sickle.wav"
	SFX_SICKLE_CLINK = "assets/sounds/sickle_clink.wav"
)

func SpawnSickle(world *World, st *scene.Storage[Projectile], position, rotation mgl32.Vec3, owner scene.Handle) (id scene.Id[*Projectile], proj *Projectile, err error) {
	id, proj, err = st.New()
	if err != nil {
		return
	}

	proj.world = world
	proj.id = id
	proj.owner = owner

	proj.body = comps.Body{
		Transform: comps.TransformFromTranslationAngles(position, rotation),
		Shape:     collision.NewSphere(0.5),
		Layer:     COL_LAYER_PROJECTILES,
		Filter:    COL_LAYER_NONE,
		LockY:     true,
	}

	sickleTex := cache.GetTexture("assets/textures/sprites/sickle_thrown.png")
	proj.SpriteRender = comps.NewSpriteRender(sickleTex)

	throwAnim, ok := sickleTex.GetAnimation("throw;front")
	if !ok {
		log.Println("could not find animation for thrown sickle sprite")
	}
	proj.AnimPlayer = comps.NewAnimationPlayer(throwAnim, true)
	proj.speed = 35.0
	proj.voices[0] = cache.GetSfx(SFX_SICKLE_THROW).PlayAttenuatedV(position)
	proj.StunChance = 0.1
	proj.Damage = 200.0

	proj.moveFunc = proj.sickleMove
	proj.body.OnIntersect = proj.sickleIntersect

	return
}

func (proj *Projectile) sickleMove(deltaTime float32) {
	var decelerationRate float32 = 50.0
	if !input.IsActionPressed(settings.ACTION_FIRE) {
		decelerationRate = 100.0
	}
	proj.speed = max(-35.0, proj.speed-deltaTime*decelerationRate)
	if owner, ok := scene.Get[HasActor](proj.owner); ok {
		if proj.speed < 0.0 {
			ownerPos := owner.Body().Transform.Position()
			projPos := proj.body.Transform.Position()
			proj.body.Transform.SetRotation(0.0, math2.Atan2(projPos.Z()-ownerPos.Z(), ownerPos.X()-projPos.X())+math2.HALF_PI, 0.0)
		}
	}

	proj.body.Velocity = mgl32.TransformNormal(mgl32.Vec3{0.0, 0.0, -proj.speed}, proj.body.Transform.Matrix())
}

func (proj *Projectile) sickleIntersect(otherEnt comps.HasBody, result collision.Result, deltaTime float32) {
	owner, hasOwner := scene.Get[HasActor](proj.owner)

	otherBody := otherEnt.Body()
	if proj.speed <= -1.0 {
		if hasOwner && otherBody == owner.Body() {
			proj.voices[0].Stop()
			proj.id.Remove()
		}
	} else if otherBody.OnLayer(COL_LAYER_MAP) {
		proj.speed = -math2.Abs(proj.speed) / 2.0
		proj.voices[1] = cache.GetSfx(SFX_SICKLE_CLINK).PlayAttenuatedV(result.Position)
	}

	// Apply damage per second
	if damageable, canDamage := otherEnt.(Damageable); canDamage && damageable != owner {
		damageable.OnDamage(proj, proj.Damage*deltaTime)
	}
}
