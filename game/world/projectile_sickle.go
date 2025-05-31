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
	"tophatdemon.com/total-invasion-ii/game"
	"tophatdemon.com/total-invasion-ii/game/settings"
)

const (
	SFX_SICKLE_THROW = "assets/sounds/weapon/sickle.wav"
	SFX_SICKLE_CLINK = "assets/sounds/weapon/sickle_clink.wav"
	SFX_SICKLE_CUT   = "assets/sounds/weapon/sickle_cut.wav"
)

func SpawnSickle(world *World, position, rotation mgl32.Vec3, owner scene.Handle) (id scene.Id[*Projectile], proj *Projectile, err error) {
	id, proj, err = world.Projectiles.New()
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
	proj.forwardSpeed = 35.0
	proj.voices[0] = cache.GetSfx(SFX_SICKLE_THROW).PlayAttenuatedV(position)
	proj.StunChance = 0.1
	proj.Damage = 200.0

	proj.moveFunc = proj.sickleMove
	proj.body.OnIntersect = proj.sickleIntersect

	return
}

func SpawnIntroSickle(world *World, position, rotation mgl32.Vec3, owner scene.Handle) (id scene.Id[*Projectile], proj *Projectile, err error) {
	id, proj, err = SpawnSickle(world, position, rotation, owner)
	proj.moveFunc = proj.introSickleMove
	proj.forwardSpeed = -1.0
	return
}

func (proj *Projectile) sickleMove(deltaTime float32) {
	var decelerationRate float32 = 50.0
	if !input.IsActionPressed(settings.ACTION_FIRE) {
		decelerationRate = 100.0
	}
	proj.forwardSpeed = max(-35.0, proj.forwardSpeed-deltaTime*decelerationRate)
	if owner, ok := scene.Get[HasActor](proj.owner); ok {
		if proj.forwardSpeed < 0.0 {
			ownerPos := owner.Body().Transform.Position()
			projPos := proj.body.Transform.Position()
			proj.body.Transform.SetRotation(0.0, math2.Atan2(projPos.Z()-ownerPos.Z(), ownerPos.X()-projPos.X())+math2.HALF_PI, 0.0)
		}
	}

	proj.body.Velocity = mgl32.TransformNormal(mgl32.Vec3{0.0, 0.0, -proj.forwardSpeed}, proj.body.Transform.Matrix())
}

func (proj *Projectile) introSickleMove(deltaTime float32) {
	proj.forwardSpeed = max(-35.0, proj.forwardSpeed-deltaTime*50.0)
	if owner, ok := scene.Get[HasActor](proj.owner); ok {
		if proj.forwardSpeed < 0.0 {
			ownerPos := owner.Body().Transform.Position()
			projPos := proj.body.Transform.Position()
			proj.body.Transform.SetRotation(0.0, math2.Atan2(projPos.Z()-ownerPos.Z(), ownerPos.X()-projPos.X())+math2.HALF_PI, 0.0)
		}
	}

	proj.body.Velocity = mgl32.TransformNormal(mgl32.Vec3{0.0, 0.0, -proj.forwardSpeed}, proj.body.Transform.Matrix())
}

func (proj *Projectile) sickleIntersect(otherEnt comps.HasBody, result collision.Result, deltaTime float32) {
	if !proj.body.OnLayer(COL_LAYER_PROJECTILES) {
		return
	}

	owner, hasOwner := scene.Get[HasActor](proj.owner)

	otherBody := otherEnt.Body()
	if proj.forwardSpeed <= -1.0 {
		if hasOwner && otherBody == owner.Body() {
			proj.voices[0].Stop()
			proj.id.Remove()
			if player, isPlayer := owner.(*Player); isPlayer {
				player.AddAmmo(game.AMMO_TYPE_SICKLE, 1)
			}
		}
	} else if otherBody.OnLayer(COL_LAYER_MAP) {
		proj.forwardSpeed = -math2.Abs(proj.forwardSpeed) / 2.0
		proj.voices[1] = cache.GetSfx(SFX_SICKLE_CLINK).PlayAttenuatedV(result.Position)
	}

	// Apply damage per second
	if damageable, canDamage := otherEnt.(Damageable); canDamage && damageable != owner {
		damaged := damageable.OnDamage(proj, proj.Damage*deltaTime)
		if damaged && !proj.voices[2].IsPlaying() {
			proj.voices[2] = cache.GetSfx(SFX_SICKLE_CUT).PlayAttenuatedV(proj.body.Transform.Position())
		}
	}
}
