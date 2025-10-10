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
	"tophatdemon.com/total-invasion-ii/engine/tdaudio"
	"tophatdemon.com/total-invasion-ii/game"
	"tophatdemon.com/total-invasion-ii/game/settings"
)

/**============================================
 *               Sickle
 *=============================================**/

func SpawnSickle(world *World, position, facing mgl32.Vec3, owner scene.Handle) (id scene.Id[*Projectile], proj *Projectile, err error) {
	id, proj, err = world.Projectiles.New()
	if err != nil {
		return
	}

	*proj = Projectile{
		world: world,
		id:    id,
		owner: owner,
		Body: comps.Body{
			Position: position,
			Shape:    collision.NewBoxShape(0.5, 0.5, 0.5),
			Layer:    ColLayerProjectiles,
		},
		Facing:       facing,
		forwardSpeed: 35.0,
		voices: [4]tdaudio.VoiceId{
			0: cache.GetSfx("assets/sounds/weapon/sickle.wav").PlayAttenuatedV(position),
		},
		StunChance: 0.1,
		Damage:     200.0,
		hitOwner:   true,
		moveFunc:   proj.sickleMove,
		onCollide:  proj.sickleCollide,
	}

	sickleTex := cache.GetTexture("assets/textures/sprites/sickle_thrown.png")
	proj.SpriteRender = comps.NewSpriteRender(sickleTex, nil, nil)

	throwAnim, ok := sickleTex.GetAnimation("throw;front")
	if !ok {
		log.Println("could not find animation for thrown sickle sprite")
	}
	proj.AnimPlayer = comps.NewAnimationPlayer(throwAnim, true)

	return
}

func SpawnIntroSickle(world *World, position, facing mgl32.Vec3, owner scene.Handle) (id scene.Id[*Projectile], proj *Projectile, err error) {
	id, proj, err = SpawnSickle(world, position, facing, owner)
	proj.Layer = ColLayerNone
	proj.moveFunc = proj.introSickleMove
	proj.forwardSpeed = -1.0

	return
}

func (proj *Projectile) sickleMove(deltaTime float32) {
	var decelerationRate float32 = 65.0
	if !input.IsActionPressed(settings.ActionFire) {
		decelerationRate = 100.0
	}
	proj.forwardSpeed = max(-35.0, proj.forwardSpeed-deltaTime*decelerationRate)
	if owner, ok := scene.Get[HasActor](proj.owner); ok {
		if proj.forwardSpeed < 0.0 {
			ownerPos := owner.Body().Position
			projPos := proj.Position
			proj.Facing = ownerPos.Sub(projPos)
			if proj.Facing.LenSqr() > 0 {
				proj.Facing = proj.Facing.Normalize()
			}
		}
	}

	proj.Velocity = proj.Facing.Mul(proj.forwardSpeed)
}

func (proj *Projectile) introSickleMove(deltaTime float32) {
	proj.forwardSpeed = max(-35.0, proj.forwardSpeed-deltaTime*50.0)
	if owner, ok := scene.Get[HasActor](proj.owner); ok {
		if proj.forwardSpeed < 0.0 {
			ownerPos := owner.Body().Position
			projPos := proj.Position
			proj.Facing = ownerPos.Sub(projPos)
			if proj.Facing.LenSqr() > 0 {
				proj.Facing = proj.Facing.Normalize()
			}
		}
	}

	proj.Velocity = proj.Facing.Mul(proj.forwardSpeed)
}

func (proj *Projectile) sickleCollide(otherEnt any, mask collision.Mask, pushVec mgl32.Vec3, deltaTime float32) {
	owner, hasOwner := scene.Get[HasActor](proj.owner)

	if bodyHaver, hasBody := otherEnt.(comps.HasBody); hasBody {
		otherBody := bodyHaver.Body()
		if proj.forwardSpeed <= -1.0 {
			if hasOwner && otherBody == owner.Body() {
				proj.voices[0].Stop()
				proj.id.Remove()
				if player, isPlayer := owner.(*Player); isPlayer {
					player.AddAmmo(game.AmmoTypeSickle, 1)
					if sickleWeapon := proj.world.Hud.Weapons.Get(game.WeaponSickle); sickleWeapon != nil {
						sickleWeapon.Equipped = true
					}
				}
			}
		}
	}

	if (mask&ColLayerMap) != 0 && proj.forwardSpeed > -1.0 {
		// Bounce off of walls
		proj.forwardSpeed = -math2.Abs(proj.forwardSpeed) / 2.0
		proj.voices[1] = cache.GetSfx("assets/sounds/weapon/sickle_clink.wav").
			PlayAttenuatedV(proj.Position)
	}

	// Apply damage per second
	if damageable, canDamage := otherEnt.(Damageable); canDamage && damageable != owner {
		damaged := damageable.OnDamage(proj, proj.Damage*deltaTime)
		if damaged && !proj.voices[2].IsPlaying() {
			proj.voices[2] = cache.GetSfx("assets/sounds/weapon/sickle_cut.wav").
				PlayAttenuatedV(proj.Position)
		}
	}
}

/**============================================
 *               Egg
 *=============================================**/

var timeSinceLastChicken time.Time

func SpawnEgg(world *World, position, facing mgl32.Vec3, owner scene.Handle) (id scene.Id[*Projectile], proj *Projectile, err error) {
	id, proj, err = world.Projectiles.New()
	if err != nil {
		return
	}

	eggTex := cache.GetTexture("assets/textures/sprites/egg.png")
	*proj = Projectile{
		world: world,
		id:    id,
		owner: owner,
		Body: comps.Body{
			Position: position,
			Shape:    collision.NewBoxShape(0.1, 0.1, 0.1),
			Layer:    ColLayerProjectiles,
		},
		Facing:       facing,
		SpriteRender: comps.NewSpriteRender(eggTex, nil, &mgl32.Vec2{0.4, 0.4}),
		forwardSpeed: 100.0,
		StunChance:   0.1,
		Damage:       15,
		moveFunc:     proj.moveForward,
		onCollide:    proj.eggCollide,
	}

	return
}

func (proj *Projectile) eggCollide(otherEnt any, mask collision.Mask, pushVec mgl32.Vec3, deltaTime float32) {
	if damageable, canDamage := otherEnt.(Damageable); canDamage {
		damageable.OnDamage(proj, proj.Damage)
	}

	proj.world.QueueRemoval(proj.id.Handle)
	proj.Layer.SetBypass()

	var backwards mgl32.Vec3
	if proj.Velocity.LenSqr() != 0.0 {
		backwards = proj.Velocity.Normalize().Mul(-1.0)
	}
	SpawnEffect(proj.world,
		comps.TransformFromTranslation(proj.Position.Add(backwards)),
		1.0,
		EggParticles(proj.Shape.Extents().Max[0]))

	chickenSpot := proj.Position.Add(backwards.Mul(1.5))
	bodiesIter := proj.world.IterBodiesInSphere(chickenSpot, 0.5, nil)
	if rand.Float32() < 0.3 && !bodiesIter.HasNext() && time.Since(timeSinceLastChicken).Seconds() > 10.0 {
		SpawnChicken(proj.world, chickenSpot, mgl32.Vec3{0.0, mgl32.RadToDeg(math2.Atan2(-backwards[0], backwards[2])), 0.0})
		timeSinceLastChicken = time.Now()
	}
}

/**============================================
 *               Fireball
 *=============================================**/

func SpawnFireball(world *World, position, facing mgl32.Vec3, owner scene.Handle) (id scene.Id[*Projectile], proj *Projectile, err error) {
	id, proj, err = world.Projectiles.New()
	if err != nil {
		return
	}

	tex := cache.GetTexture("assets/textures/sprites/fireball.png")
	*proj = Projectile{
		world: world,
		id:    id,
		owner: owner,
		Body: comps.Body{
			Position: position,
			Shape:    collision.NewBoxShape(0.25, 0.25, 0.25),
			Layer:    ColLayerProjectiles,
		},
		Facing:       facing,
		SpriteRender: comps.NewSpriteRender(tex, nil, &mgl32.Vec2{0.375, 0.375}),
		AnimPlayer:   comps.NewAnimationPlayer(tex.GetDefaultAnimation(), true),
		forwardSpeed: 70.0,
		voices: [4]tdaudio.VoiceId{
			cache.GetSfx("assets/sounds/fireball.wav").PlayAttenuatedV(position),
		},
		StunChance: 0.1,
		Damage:     15,
		moveFunc:   proj.moveForward,
		onCollide:  proj.dieOnCollide,
	}

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

	tex := cache.GetTexture("assets/textures/sprites/grenade.png")

	*proj = Projectile{
		world: world,
		id:    id,
		Body: comps.Body{
			Position: position,
			Shape:    collision.NewBoxShape(0.25, 0.25, 0.25),
			Velocity: direction.Mul(20.0),
			Layer:    ColLayerProjectiles,
		},
		SpriteRender: comps.NewSpriteRender(tex, nil, &mgl32.Vec2{0.25, 0.25}),
		AnimPlayer:   comps.NewAnimationPlayer(tex.GetDefaultAnimation(), true),
		StunChance:   0.0,
		Damage:       15,
		gravity:      -25.0,
		maxFallSpeed: 10.0,
		moveFunc:     proj.applyGravity,
		maxLife:      1.5,
		onDie:        proj.explodeOnDie,
		onCollide:    proj.grenadeCollide,
	}

	return
}

func (proj *Projectile) applyGravity(deltaTime float32) {
	proj.fallSpeed = max(-proj.maxFallSpeed, proj.fallSpeed+(proj.gravity*deltaTime))
	proj.Velocity = math2.Vec3WithY(proj.Velocity, proj.fallSpeed)
}

func (proj *Projectile) grenadeCollide(otherEnt any, mask collision.Mask, pushVec mgl32.Vec3, deltaTime float32) {
	if damageable, canDamage := otherEnt.(Damageable); canDamage {
		damageable.OnDamage(proj, proj.Damage)
		proj.explodeOnDie(deltaTime)
	} else if (mask & ColLayerKillzone) != 0 {
		proj.explodeOnDie(deltaTime)
	} else if (mask & ColLayerMap) != 0 {
		horzVelocity := math2.Vec3WithY(proj.Velocity, 0.0)
		speed := horzVelocity.Len() * 0.8
		if pushVec[1] > 0.1 && proj.fallSpeed < 0.0 {
			if proj.fallSpeed > -0.01 {
				proj.fallSpeed = 0.0
			}
			proj.fallSpeed = -proj.fallSpeed * 0.9
			proj.Velocity = math2.Vec3WithY(horzVelocity.Normalize().Mul(speed), proj.fallSpeed)
		} else if pushVec.LenSqr() > 0.0 {
			if speed > 0.01 {
				reflection := math2.Vec3Reflect(horzVelocity.Normalize(), pushVec.Normalize())
				proj.Velocity = mgl32.Vec3{reflection.X() * speed, 0.0, reflection.Z() * speed}
			} else {
				proj.Velocity = mgl32.Vec3{}
			}
		}
	}
}

func (proj *Projectile) explodeOnDie(deltaTime float32) {
	_ = deltaTime
	proj.Position = proj.Position.Add(mgl32.Vec3{0.0, 0.5, 0.0})
	SpawnSingleExplosion(proj.world, proj.Position)
	proj.world.QueueRemoval(proj.id.Handle)
}

/**============================================
 *               Plasma Ball
 *=============================================**/

func SpawnPlasmaBall(
	world *World, position, facing mgl32.Vec3,
	owner scene.Handle, bigShot bool,
) (id scene.Id[*Projectile], proj *Projectile, err error) {
	id, proj, err = world.Projectiles.New()
	if err != nil {
		return
	}

	*proj = Projectile{
		world: world,
		id:    id,
		owner: owner,
		Body: comps.Body{
			Position: position,
			Layer:    ColLayerProjectiles,
		},
		Facing:       facing,
		forwardSpeed: 120.0,
		StunChance:   0.1,
		moveFunc:     proj.moveForward,
		onCollide:    proj.dieOnCollide,
		onDie:        proj.playAnimOnDie,
	}

	scale := float32(0.25)

	var tex *textures.Texture
	if bigShot {
		// NOW'S YOUR CHANCE TO BE A BIG SHOT
		scale = 0.35
		proj.Shape = collision.NewBoxShape(0.35, 0.35, 0.35)
		proj.knockbackForce = 5.0
		proj.Damage = 8
		tex = cache.GetTexture("assets/textures/sprites/big_plasma_ball.png")
	} else {
		proj.Shape = collision.NewBoxShape(0.25, 0.25, 0.25)
		proj.knockbackForce = 12.0
		proj.Damage = 5
		tex = cache.GetTexture("assets/textures/sprites/plasma_ball.png")
	}

	proj.SpriteRender = comps.NewSpriteRender(tex, nil, &mgl32.Vec2{scale, scale})
	proj.AnimPlayer = comps.NewAnimationPlayer(tex.GetDefaultAnimation(), true)
	proj.dieAnim, _ = tex.GetAnimation("die")

	return
}

/**============================================
 *               Blessing
 *=============================================**/

func SpawnBlessing(world *World, position, facing mgl32.Vec3, owner scene.Handle) (id scene.Id[*Projectile], proj *Projectile, err error) {
	id, proj, err = world.Projectiles.New()
	if err != nil {
		return
	}

	tex := cache.GetTexture("assets/textures/sprites/blessing.png")

	*proj = Projectile{
		world:  world,
		id:     id,
		owner:  owner,
		Facing: facing,
		Body: comps.Body{
			Position: position,
			Shape:    collision.NewBoxShape(0.5, 0.5, 0.5),
			Layer:    ColLayerProjectiles,
		},
		SpriteRender: comps.NewSpriteRender(tex, nil, &mgl32.Vec2{0.5, 0.5}),
		AnimPlayer:   comps.NewAnimationPlayer(tex.GetDefaultAnimation(), true),
		forwardSpeed: 30.0,
		voices: [4]tdaudio.VoiceId{
			cache.GetSfx("assets/sounds/blessing.wav").PlayAttenuatedV(position),
		},
		StunChance: 0.0,
		Damage:     15,
		moveFunc:   proj.moveForwardAndRevive,
		onCollide:  proj.blessingCollide,
	}

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
			diff := enemy.Body().Position.Sub(proj.Position)
			dist := diff.Len()
			if dist < 2.0 {
				// Ensure we are not reviving an enemy from behind a wall.
				rayHit, _ := proj.world.Raycast(proj.Position, diff.Mul(1.0/dist), ColLayerMap, dist, nil)
				if !rayHit.Hit {
					enemy.changeState(&enemy.reviveState)
				}
			}
		}
	}
}

func (proj *Projectile) blessingCollide(collidingEntity any, mask collision.Mask, pushVec mgl32.Vec3, deltaTime float32) {
	if _, isEnemy := collidingEntity.(*Enemy); !isEnemy {
		proj.dieOnCollide(collidingEntity, mask, pushVec, deltaTime)
	}
}
