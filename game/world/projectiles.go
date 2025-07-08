package world

import (
	"log"
	"math/rand"
	"time"

	"github.com/go-gl/mathgl/mgl32"
	"tophatdemon.com/total-invasion-ii/engine/assets/cache"
	"tophatdemon.com/total-invasion-ii/engine/assets/textures"
	"tophatdemon.com/total-invasion-ii/engine/input"
	"tophatdemon.com/total-invasion-ii/engine/math2"
	"tophatdemon.com/total-invasion-ii/engine/math2/collision"
	"tophatdemon.com/total-invasion-ii/engine/scene"
	"tophatdemon.com/total-invasion-ii/engine/scene/comps"
	"tophatdemon.com/total-invasion-ii/game"
	"tophatdemon.com/total-invasion-ii/game/settings"
)

/**============================================
 *               Sickle
 *=============================================**/

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
		Layer:     ColLayerProjectiles,
		Filter:    ColLayerNone,
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
	proj.voices[0] = cache.GetSfx("assets/sounds/weapon/sickle.wav").PlayAttenuatedV(position)
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
	if !proj.body.OnLayer(ColLayerProjectiles) {
		return
	}

	owner, hasOwner := scene.Get[HasActor](proj.owner)

	otherBody := otherEnt.Body()
	if proj.forwardSpeed <= -1.0 {
		if hasOwner && otherBody == owner.Body() {
			proj.voices[0].Stop()
			proj.id.Remove()
			if player, isPlayer := owner.(*Player); isPlayer {
				player.AddAmmo(game.AmmoTypeSickle, 1)
				if sickleWeapon := proj.world.Hud.Weapons.Get(game.WeaponSickle); sickleWeapon != nil {
					sickleWeapon.Equipped = true
					proj.world.Hud.Weapons.Select(game.WeaponSickle)
				}
			}
		}
	} else if otherBody.OnLayer(ColLayerMap) {
		proj.forwardSpeed = -math2.Abs(proj.forwardSpeed) / 2.0
		proj.voices[1] = cache.GetSfx("assets/sounds/weapon/sickle_clink.wav").
			PlayAttenuatedV(result.Position)
	}

	// Apply damage per second
	if damageable, canDamage := otherEnt.(Damageable); canDamage && damageable != owner {
		damaged := damageable.OnDamage(proj, proj.Damage*deltaTime)
		if damaged && !proj.voices[2].IsPlaying() {
			proj.voices[2] = cache.GetSfx("assets/sounds/weapon/sickle_cut.wav").
				PlayAttenuatedV(proj.body.Transform.Position())
		}
	}
}

/**============================================
 *               Egg
 *=============================================**/

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
		Layer:     ColLayerProjectiles,
		Filter:    ColLayerNone,
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
		EggShardsEffect(proj.body.Shape.(collision.Sphere).Radius()))

	chickenSpot := proj.body.Transform.Position().Add(backwards.Mul(1.5))
	noBlockers := !proj.world.BodiesInSphere(chickenSpot, 0.5, proj).Any()
	if rand.Float32() < 0.3 && noBlockers && time.Since(timeSinceLastChicken).Seconds() > 10.0 {
		SpawnChicken(proj.world, chickenSpot, mgl32.Vec3{0.0, mgl32.RadToDeg(math2.Atan2(-backwards[0], backwards[2])), 0.0})
		timeSinceLastChicken = time.Now()
	}
}

/**============================================
 *               Fireball
 *=============================================**/

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
		Layer:     ColLayerProjectiles,
		Filter:    ColLayerNone,
		LockY:     true,
	}

	tex := cache.GetTexture("assets/textures/sprites/fireball.png")
	proj.SpriteRender = comps.NewSpriteRender(tex)
	proj.AnimPlayer = comps.NewAnimationPlayer(tex.GetDefaultAnimation(), true)
	proj.forwardSpeed = 70.0
	proj.voices[0] = cache.GetSfx("assets/sounds/fireball.wav").PlayAttenuatedV(position)
	proj.StunChance = 0.1
	proj.Damage = 15

	proj.moveFunc = proj.moveForward
	proj.body.OnIntersect = proj.dieOnHit

	return
}

/**============================================
 *               Grenade
 *=============================================**/

func SpawnGrenade(world *World, position, direction mgl32.Vec3) (id scene.Id[*Projectile], proj *Projectile, err error) {
	id, proj, err = world.Projectiles.New()
	if err != nil {
		return
	}

	proj.world = world
	proj.id = id

	tex := cache.GetTexture("assets/textures/sprites/grenade.png")
	proj.SpriteRender = comps.NewSpriteRender(tex)
	proj.AnimPlayer = comps.NewAnimationPlayer(tex.GetDefaultAnimation(), true)
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
		Layer:       ColLayerProjectiles,
		Filter:      ColLayerMap,
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
		proj.explodeOnDie(deltaTime)
	} else if otherEnt.Body().OnLayer(ColLayerKillzone) {
		proj.explodeOnDie(deltaTime)
	} else if otherEnt.Body().OnLayer(ColLayerMap) {
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

func (proj *Projectile) explodeOnDie(deltaTime float32) {
	_ = deltaTime
	proj.body.Transform.Translate(0.0, 0.5, 0.0)
	SpawnSingleExplosion(proj.world, proj.body.Transform)
	proj.world.QueueRemoval(proj.id.Handle)
}

/**============================================
 *               Plasma Ball
 *=============================================**/

func SpawnPlasmaBall(world *World, position, rotation mgl32.Vec3, owner scene.Handle, bigShot bool) (id scene.Id[*Projectile], proj *Projectile, err error) {
	id, proj, err = world.Projectiles.New()
	if err != nil {
		return
	}

	proj.world = world
	proj.id = id
	proj.owner = owner

	proj.body = comps.Body{
		Layer:  ColLayerProjectiles,
		Filter: ColLayerNone,
		LockY:  true,
	}
	var tex *textures.Texture
	if bigShot {
		// NOW'S YOUR CHANCE TO BE A BIG SHOT
		proj.body.Transform = comps.TransformFromTranslationAnglesScale(position, rotation, mgl32.Vec3{0.35, 0.35, 0.35})
		proj.body.Shape = collision.NewSphere(0.35)
		proj.knockbackForce = 5.0
		tex = cache.GetTexture("assets/textures/sprites/big_plasma_ball.png")
	} else {
		proj.body.Transform = comps.TransformFromTranslationAnglesScale(position, rotation, mgl32.Vec3{0.25, 0.25, 0.25})
		proj.body.Shape = collision.NewSphere(0.25)
		proj.knockbackForce = 12.0
		tex = cache.GetTexture("assets/textures/sprites/plasma_ball.png")
	}

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

/**============================================
 *               Blessing
 *=============================================**/

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
		Layer:     ColLayerProjectiles,
		Filter:    ColLayerNone,
		LockY:     true,
	}

	tex := cache.GetTexture("assets/textures/sprites/blessing.png")
	proj.SpriteRender = comps.NewSpriteRender(tex)
	proj.AnimPlayer = comps.NewAnimationPlayer(tex.GetDefaultAnimation(), true)
	proj.forwardSpeed = 20.0
	//TODO: Add unique sound effect
	proj.voices[0] = cache.GetSfx("assets/sounds/fireball.wav").PlayAttenuatedV(position)
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
		if enemy == nil {
			break
		}
		if handle.Equals(proj.owner) {
			continue
		}
		if enemy.actor.Health <= 0.0 {
			diff := enemy.Body().Transform.Position().Sub(proj.body.Transform.Position())
			dist := diff.Len()
			if dist < 2.0 {
				// Ensure we are not reviving an enemy from behind a wall.
				rayHit, _ := proj.world.Raycast(proj.body.Transform.Position(), diff.Mul(1.0/dist), ColLayerMap, dist, nil)
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
