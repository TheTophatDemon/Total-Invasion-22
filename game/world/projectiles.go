package world

import (
	"log"
	"math/rand"
	"time"

	"github.com/go-gl/mathgl/mgl32"
	"tophatdemon.com/total-invasion-ii/engine/assets/cache"
	"tophatdemon.com/total-invasion-ii/engine/assets/textures"
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

func SpawnSickle(position, facing mgl32.Vec3, owner scene.Handle) (id scene.Id[*Projectile], proj *Projectile, err error) {
	id, proj, err = gWorld.Projectiles.New()
	if err != nil {
		return
	}

	*proj = Projectile{
		id:    id,
		owner: owner,
		body: comps.Body{
			Position: position,
			Shape:    collision.NewBoxShape(0.5, 0.5, 0.5),
			Layers:   ColLayerProjectiles,
		},
		Facing:       facing,
		forwardSpeed: 35.0,
		voices: [4]tdaudio.VoiceId{
			0: cache.GetSfx("assets/sounds/weapon/sickle.wav").PlayAttenuatedV(position),
		},
		StunChance: 0.1,
		Damage:     200.0,
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

func SpawnIntroSickle(position, facing mgl32.Vec3, owner scene.Handle) (id scene.Id[*Projectile], proj *Projectile, err error) {
	id, proj, err = SpawnSickle(position, facing, owner)
	proj.body.Layers = ColLayerNone
	proj.moveFunc = proj.introSickleMove
	proj.forwardSpeed = -1.0

	return
}

func (proj *Projectile) sickleMove(deltaTime float32) {
	var decelerationRate float32 = 65.0
	if !settings.Current.ActionFire.Pressed() {
		decelerationRate = 100.0
	}
	proj.forwardSpeed = max(-35.0, proj.forwardSpeed-deltaTime*decelerationRate)
	if owner, ok := proj.owner.Get[HasActor](); ok {
		if proj.forwardSpeed < 0.0 {
			proj.hitOwner = true
			ownerPos := owner.Body().Position
			projPos := proj.body.Position
			proj.Facing = projPos.Sub(ownerPos)
			if proj.Facing.LenSqr() > 0 {
				proj.Facing = proj.Facing.Normalize()
			}
		}
	}

	proj.body.Velocity = proj.Facing.Mul(proj.forwardSpeed)
}

func (proj *Projectile) introSickleMove(deltaTime float32) {
	proj.forwardSpeed = max(-35.0, proj.forwardSpeed-deltaTime*50.0)
	if owner, ok := proj.owner.Get[HasActor](); ok {
		if proj.forwardSpeed < 0.0 {
			proj.hitOwner = true
			ownerPos := owner.Body().Position
			projPos := proj.body.Position
			proj.Facing = projPos.Sub(ownerPos)
			if proj.Facing.LenSqr() > 0 {
				proj.Facing = proj.Facing.Normalize()
			}
		}
	}

	proj.body.Velocity = proj.Facing.Mul(proj.forwardSpeed)
}

func (proj *Projectile) sickleCollide(movement mgl32.Vec3, otherEnt any, mask collision.Mask, result collision.Result, deltaTime float32) mgl32.Vec3 {
	owner, hasOwner := proj.owner.Get[HasActor]()

	if bodyHaver, hasBody := otherEnt.(comps.HasBody); hasBody {
		otherBody := bodyHaver.Body()
		if proj.forwardSpeed <= -1.0 {
			if hasOwner && otherBody == owner.Body() {
				proj.voices[0].Stop()
				proj.id.Remove()
				if player, isPlayer := owner.(*Player); isPlayer {
					player.AddAmmo(game.AmmoTypeSickle, 1)
					player.Sickle.Equipped = true
				}
			}
		}
	}

	if (mask&ColLayerMap) != 0 && proj.forwardSpeed > -1.0 {
		// Bounce off of walls
		proj.forwardSpeed = -math2.Abs(proj.forwardSpeed) / 2.0
		proj.voices[1] = cache.GetSfx("assets/sounds/weapon/sickle_clink.wav").
			PlayAttenuatedV(proj.body.Position)
	}

	// Apply damage per second
	if damageable, canDamage := otherEnt.(Damageable); canDamage && damageable != owner && proj.body.Layers.On(ColLayerProjectiles) {
		damaged := damageable.OnDamage(proj, proj.Damage*deltaTime)
		if damaged && !proj.voices[2].IsPlaying() {
			proj.voices[2] = cache.GetSfx("assets/sounds/weapon/sickle_cut.wav").
				PlayAttenuatedV(proj.body.Position)
		}
	}

	return movement
}

/**============================================
 *               Egg
 *=============================================**/

var timeSinceLastChicken time.Time

func SpawnEgg(position, facing mgl32.Vec3, owner scene.Handle) (id scene.Id[*Projectile], proj *Projectile, err error) {
	id, proj, err = gWorld.Projectiles.New()
	if err != nil {
		return
	}

	eggTex := cache.GetTexture("assets/textures/sprites/egg.png")
	*proj = Projectile{
		id:    id,
		owner: owner,
		body: comps.Body{
			Position: position,
			Shape:    collision.NewBoxShape(0.1, 0.1, 0.1),
			Layers:   ColLayerProjectiles,
		},
		Facing:       facing,
		SpriteRender: comps.NewSpriteRender(eggTex, nil, &mgl32.Vec2{0.4, 0.4}),
		forwardSpeed: 50.0,
		StunChance:   0.1,
		Damage:       15,
		moveFunc:     proj.moveForward,
		onCollide:    proj.eggCollide,
	}

	return
}

func (proj *Projectile) eggCollide(movement mgl32.Vec3, otherEnt any, mask collision.Mask, result collision.Result, deltaTime float32) mgl32.Vec3 {
	if damageable, canDamage := otherEnt.(Damageable); canDamage {
		damageable.OnDamage(proj, proj.Damage)
	}

	gWorld.QueueRemoval(proj.id.Handle)
	proj.body.Layers = 0
	proj.onCollide = nil
	proj.moveFunc = nil

	var backwards mgl32.Vec3
	if proj.body.Velocity.LenSqr() != 0.0 {
		backwards = proj.body.Velocity.Normalize().Mul(-1.0)
	}
	SpawnEffect(proj.body.Position.Add(backwards), 1.0, EggParticles(proj.body.Shape.Extents().Max[0]))

	chickenSpot := proj.body.Position.Add(backwards.Mul(1.5))
	bodiesIter := gWorld.IterBodiesInSphere(chickenSpot, 0.5, nil)
	if rand.Float32() < 0.3 && !bodiesIter.HasNext() && time.Since(timeSinceLastChicken).Seconds() > 10.0 {
		SpawnChicken(chickenSpot, [3]math2.Degrees{0.0, math2.ToDegrees(math2.Radians(math2.Atan2(-backwards[0], backwards[2]))), 0.0})
		timeSinceLastChicken = time.Now()
	}

	return proj.stopOnCollide(movement, otherEnt, mask, result, deltaTime)
}

/**============================================
 *               Fireball
 *=============================================**/

func SpawnFireball(position, facing mgl32.Vec3, owner scene.Handle) (id scene.Id[*Projectile], proj *Projectile, err error) {
	id, proj, err = gWorld.Projectiles.New()
	if err != nil {
		return
	}

	tex := cache.GetTexture("assets/textures/sprites/fireball.png")
	*proj = Projectile{
		id:    id,
		owner: owner,
		body: comps.Body{
			Position: position,
			Shape:    collision.NewBoxShape(0.25, 0.25, 0.25),
			Layers:   ColLayerProjectiles,
		},
		Facing:       facing,
		SpriteRender: comps.NewSpriteRender(tex, nil, &mgl32.Vec2{0.375, 0.375}),
		AnimPlayer:   comps.NewAnimationPlayer(tex.GetDefaultAnimation(), true),
		forwardSpeed: 35.0,
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

func SpawnGrenade(position, direction mgl32.Vec3, owner scene.Handle) (id scene.Id[*Projectile], proj *Projectile, err error) {
	id, proj, err = gWorld.Projectiles.New()
	if err != nil {
		return
	}

	tex := cache.GetTexture("assets/textures/sprites/grenade.png")

	*proj = Projectile{
		id: id,
		body: comps.Body{
			Position: position,
			Shape:    collision.NewBoxShape(0.25, 0.25, 0.25),
			Velocity: direction.Mul(20.0),
			Layers:   ColLayerProjectiles,
		},
		SpriteRender: comps.NewSpriteRender(tex, nil, &mgl32.Vec2{0.25, 0.25}),
		AnimPlayer:   comps.NewAnimationPlayer(tex.GetDefaultAnimation(), true),
		StunChance:   0.0,
		Damage:       15,
		gravity:      -25.0,
		maxFallSpeed: 10.0,
		owner:        owner,
		hitOwner:     false,
		moveFunc:     proj.applyGravity,
		maxLife:      1.5,
		onDie:        proj.explodeOnDie,
		onCollide:    proj.grenadeCollide,
	}

	return
}

func (proj *Projectile) applyGravity(deltaTime float32) {
	proj.fallSpeed = max(-proj.maxFallSpeed, proj.fallSpeed+(proj.gravity*deltaTime))
	proj.body.Velocity = math2.Vec3WithY(proj.body.Velocity, proj.fallSpeed)
}

func (proj *Projectile) grenadeCollide(movement mgl32.Vec3, otherEnt any, mask collision.Mask, result collision.Result, deltaTime float32) mgl32.Vec3 {
	if damageable, canDamage := otherEnt.(Damageable); canDamage {
		damageable.OnDamage(proj, proj.Damage)
		proj.explodeOnDie(deltaTime)
	} else if (mask & ColLayerKillzone) != 0 {
		proj.explodeOnDie(deltaTime)
	} else if (mask & ColLayerMap) != 0 {
		// The grenade won't hit its owner until its first bounce to prevent grenades exploding in the player's face
		proj.hitOwner = true

		fuzzyNormal := result.Normal.Add(math2.RandomDir().Mul(0.1)).Normalize()

		horzVelocity := math2.Vec3WithY(proj.body.Velocity, 0.0)
		speed := horzVelocity.Len() * 0.7
		var direction mgl32.Vec3
		if speed > 0 {
			direction = horzVelocity.Normalize()
		}

		if fuzzyNormal[1] > 0.1 && proj.fallSpeed < 0.0 {
			if proj.fallSpeed > -0.01 {
				proj.fallSpeed = 0.0
			}
			proj.fallSpeed = -proj.fallSpeed * 0.9
			proj.body.Velocity = math2.Vec3WithY(direction.Mul(speed), proj.fallSpeed)
		} else if fuzzyNormal.LenSqr() > 0.0 {
			if speed > 0.01 {
				reflection := math2.Vec3Reflect(direction, fuzzyNormal)
				proj.body.Velocity = mgl32.Vec3{reflection.X() * speed, 0.0, reflection.Z() * speed}
			} else {
				proj.body.Velocity = mgl32.Vec3{}
			}
		}

	}
	return proj.stopOnCollide(movement, otherEnt, mask, result, deltaTime)
}

func (proj *Projectile) explodeOnDie(deltaTime float32) {
	_ = deltaTime
	SpawnSingleExplosion(proj.body.Position.Add(mgl32.Vec3{0.0, 0.5, 0.0}))
	gWorld.QueueRemoval(proj.id.Handle)
}

/**============================================
 *               Plasma Ball
 *=============================================**/

func SpawnPlasmaBall(
	position, facing mgl32.Vec3,
	owner scene.Handle, bigShot bool,
) (id scene.Id[*Projectile], proj *Projectile, err error) {
	id, proj, err = gWorld.Projectiles.New()
	if err != nil {
		return
	}

	*proj = Projectile{
		id:    id,
		owner: owner,
		body: comps.Body{
			Position: position,
			Layers:   ColLayerProjectiles,
		},
		Facing:       facing,
		forwardSpeed: 60.0,
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
		proj.body.Shape = collision.NewBoxShape(0.35, 0.35, 0.35)
		proj.knockbackForce = 5.0
		proj.Damage = 8
		tex = cache.GetTexture("assets/textures/sprites/big_plasma_ball.png")
	} else {
		proj.body.Shape = collision.NewBoxShape(0.25, 0.25, 0.25)
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

func SpawnBlessing(position, facing mgl32.Vec3, owner scene.Handle) (id scene.Id[*Projectile], proj *Projectile, err error) {
	id, proj, err = gWorld.Projectiles.New()
	if err != nil {
		return
	}

	tex := cache.GetTexture("assets/textures/sprites/blessing.png")

	*proj = Projectile{
		id:     id,
		owner:  owner,
		Facing: facing,
		body: comps.Body{
			Position: position,
			Shape:    collision.NewBoxShape(0.5, 0.5, 0.5),
			Layers:   ColLayerProjectiles,
		},
		SpriteRender: comps.NewSpriteRender(tex, nil, &mgl32.Vec2{0.5, 0.5}),
		AnimPlayer:   comps.NewAnimationPlayer(tex.GetDefaultAnimation(), true),
		forwardSpeed: 15.0,
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

	enemiesIter := gWorld.Enemies.Iter()
	for {
		enemy, handle := enemiesIter.Next()
		if enemy == nil {
			break
		}
		if handle.Equals(proj.owner) {
			continue
		}
		if enemy.actor.Health <= 0.0 {
			diff := enemy.Body().Position.Sub(proj.body.Position)
			dist := diff.Len()
			if dist < 2.0 {
				// Ensure we are not reviving an enemy from behind a wall.
				rayHit, _ := gWorld.Raycast(proj.body.Position, diff.Mul(1.0/dist), ColLayerMap, dist, nil)
				if !rayHit.Hit {
					enemy.changeState(&enemy.reviveState)
				}
			}
		}
	}
}

func (proj *Projectile) blessingCollide(movement mgl32.Vec3, collidingEntity any, mask collision.Mask, result collision.Result, deltaTime float32) mgl32.Vec3 {
	if _, isEnemy := collidingEntity.(*Enemy); !isEnemy {
		return proj.dieOnCollide(movement, collidingEntity, mask, result, deltaTime)
	}
	return movement
}
