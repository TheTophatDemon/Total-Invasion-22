package world

import (
	"math"
	"math/rand"
	"strings"

	"github.com/go-gl/mathgl/mgl32"
	"tophatdemon.com/total-invasion-ii/engine/assets/cache"
	"tophatdemon.com/total-invasion-ii/engine/color"
	"tophatdemon.com/total-invasion-ii/engine/math2"
	"tophatdemon.com/total-invasion-ii/engine/math2/collision"
	"tophatdemon.com/total-invasion-ii/engine/scene"
	"tophatdemon.com/total-invasion-ii/game"
	"tophatdemon.com/total-invasion-ii/game/settings"
)

/**============================================
 *               Wraith
 *=============================================**/

const wraithMeleeRange = 2.5

func configureWraith(enemy *Enemy) (params enemyConfig) {
	params.texture = cache.GetTexture("assets/textures/sprites/wraith.png")
	walkAnim, _ := params.texture.GetAnimation("walk;front")
	attackAnim, _ := params.texture.GetAnimation("attack;front")
	stunAnim, _ := params.texture.GetAnimation("hurt;front")
	dieAnim, _ := params.texture.GetAnimation("die;front")
	params.defaultAnim = walkAnim
	params.bloodColor = color.Red

	enemy.idleState = enemyState{
		anim:       walkAnim,
		stopAnim:   true,
		leaveSound: cache.GetSfx("assets/sounds/enemy/wraith/wraith_greeting.wav"),
	}
	enemy.chaseState = enemyState{
		anim:       walkAnim,
		updateFunc: wraithChaseUpdate,
	}
	enemy.stunState = enemyState{
		enterSound: cache.GetSfx("assets/sounds/enemy/wraith/wraith_hurt.wav"),
		anim:       stunAnim,
	}
	enemy.attackState = enemyState{
		anim:       attackAnim,
		updateFunc: wraithAttackUpdate,
	}
	enemy.dieState = enemyState{
		enterSound: cache.GetSfx("assets/sounds/enemy/wraith/wraith_die.wav"),
		anim:       dieAnim,
	}

	reviveAnim, _ := params.texture.GetAnimation("revive;front")
	enemy.reviveState = enemyState{
		enterSound: cache.GetSfx("assets/sounds/enemy/wraith/wraith_revive.wav"),
		anim:       reviveAnim,
	}

	enemy.actor.MaxHealth = 90.0

	return
}

func wraithChaseUpdate(enemy *Enemy, deltaTime float32) {
	enemy.chase(deltaTime, 3.0, 1.0)
	if enemy.distToTarget < wraithMeleeRange {
		enemy.changeState(&enemy.attackState)
	}
}

func wraithAttackUpdate(enemy *Enemy, deltaTime float32) {
	enemy.actor.inputForward, enemy.actor.inputStrafe = 0.0, 0.0
	if enemy.AnimPlayer.HitATriggerFrame() {
		if enemy.distToTarget >= wraithMeleeRange {
			enemy.changeState(&enemy.chaseState)
		} else if player, ok := gWorld.CurrentPlayer.Get(); ok {
			player.OnDamage(enemy, settings.CurrDifficulty().WraithMeleeDamage)
		}
	}
}

/**============================================
 *               Fire Wraith
 *=============================================**/

func configureFireWraith(enemy *Enemy) (params enemyConfig) {
	params.bloodColor = color.Blue
	params.texture = cache.GetTexture("assets/textures/sprites/fire_wraith.png")
	walkAnim, _ := params.texture.GetAnimation("walk;front")
	attackAnim, _ := params.texture.GetAnimation("attack;front")
	stunAnim, _ := params.texture.GetAnimation("hurt;front")
	dieAnim, _ := params.texture.GetAnimation("die;front")
	params.defaultAnim = walkAnim

	enemy.idleState = enemyState{
		anim:       walkAnim,
		stopAnim:   true,
		leaveSound: cache.GetSfx("assets/sounds/enemy/fire_wraith/fire_wraith_greeting.wav"),
	}
	enemy.chaseState = enemyState{
		anim:       walkAnim,
		enterFunc:  fireWraithEnterChase,
		updateFunc: fireWraithUpdateChase,
	}
	enemy.stunState = enemyState{
		enterSound: cache.GetSfx("assets/sounds/enemy/fire_wraith/fire_wraith_hurt.wav"),
		anim:       stunAnim,
	}
	enemy.attackState = enemyState{
		anim:       attackAnim,
		enterFunc:  fireWraithEnterAttack,
		updateFunc: fireWraithUpdateAttack,
	}
	enemy.dieState = enemyState{
		enterSound: cache.GetSfx("assets/sounds/enemy/fire_wraith/fire_wraith_die.wav"),
		anim:       dieAnim,
	}

	reviveAnim, _ := params.texture.GetAnimation("revive;front")
	enemy.reviveState = enemyState{
		enterSound: cache.GetSfx("assets/sounds/enemy/fire_wraith/fire_wraith_revive.wav"),
		anim:       reviveAnim,
	}

	enemy.actor.MaxSpeed = 6.0
	enemy.actor.MaxHealth = 150.0 //40

	return
}

func fireWraithEnterChase(enemy *Enemy, oldState *enemyState) {
	enemy.attackTimer = rand.Float32() + 0.5
}

func fireWraithUpdateChase(enemy *Enemy, deltaTime float32) {
	enemy.stalk(deltaTime, 1.0)
	enemy.attackTimer -= deltaTime
	if enemy.attackTimer <= 0.0 {
		hit, _ := gWorld.Raycast(
			enemy.actor.Position(),
			enemy.dirToTarget,
			ColLayerMap|ColLayerNPCs,
			enemy.distToTarget,
			enemy.Body(),
		)
		if !hit.Hit {
			enemy.changeState(&enemy.attackState)
		}
	}
}

func fireWraithEnterAttack(enemy *Enemy, oldState *enemyState) {
	enemy.attackTimer = 0.0
	enemy.faceTarget()
}

func fireWraithUpdateAttack(enemy *Enemy, deltaTime float32) {
	enemy.actor.inputForward, enemy.actor.inputStrafe = 0.0, 0.0
	if enemy.AnimPlayer.HitTriggerFrame(0) {
		enemy.faceTarget()
		SpawnFireball(enemy.actor.Position(), enemy.actor.FacingVec(), enemy.id.Handle)
	}
	if enemy.AnimPlayer.IsAtEnd() {
		enemy.changeState(&enemy.chaseState)
	}
}

/**============================================
 *               Mother Wraith
 *=============================================**/

func configureMotherWraith(enemy *Enemy) (params enemyConfig) {
	params.texture = cache.GetTexture("assets/textures/sprites/mother_wraith.png")
	floatAnim, _ := params.texture.GetAnimation("float;front")
	attackAnim, _ := params.texture.GetAnimation("attack;front")
	stunAnim, _ := params.texture.GetAnimation("hurt;front")
	dieAnim, _ := params.texture.GetAnimation("die;front")
	params.defaultAnim = floatAnim
	params.bloodColor = color.Color{G: 1.0, B: 0.51, A: 1.0}

	enemy.idleState = enemyState{
		anim:       floatAnim,
		stopAnim:   true,
		leaveSound: cache.GetSfx("assets/sounds/enemy/mother_wraith/mother_wraith_greeting.wav"),
	}
	enemy.chaseState = enemyState{
		anim:       floatAnim,
		enterFunc:  motherWraithEnterChase,
		updateFunc: motherWraithUpdateChase,
	}
	enemy.stunState = enemyState{
		enterSound: cache.GetSfx("assets/sounds/enemy/mother_wraith/mother_wraith_hurt.wav"),
		anim:       stunAnim,
	}
	enemy.attackState = enemyState{
		anim:       attackAnim,
		enterFunc:  motherWraithEnterAttack,
		updateFunc: motherWraithUpdateAttack,
	}
	enemy.dieState = enemyState{
		enterSound: cache.GetSfx("assets/sounds/enemy/mother_wraith/mother_wraith_die.wav"),
		anim:       dieAnim,
	}

	reviveAnim, _ := params.texture.GetAnimation("revive;front")
	enemy.reviveState = enemyState{
		enterSound: cache.GetSfx("assets/sounds/enemy/mother_wraith/mother_wraith_revive.wav"),
		anim:       reviveAnim,
	}

	enemy.actor.AccelRate = 50.0
	enemy.actor.MaxSpeed = 4.0
	enemy.actor.MaxHealth = 350.0
	enemy.StunChance = 0.5

	return
}

func motherWraithEnterChase(enemy *Enemy, oldState *enemyState) {
	if oldState == &enemy.idleState {
		enemy.attackTimer = 1.0
	} else {
		enemy.attackTimer = rand.Float32() + 1.5
	}

	// Switch periodically between shooting at the player and shooting to revive nearby enemies.
	if rand.Float32() < 0.5 {
		enemy.targetHandle = gWorld.CurrentPlayer.Handle
	} else {
		nearbyEnemiesIter := gWorld.Enemies.Iter()
		var nearestCorpseHandle scene.Handle
		nearestCorpseDistance := float32(math.MaxFloat32)
		for {
			corpse, handle := nearbyEnemiesIter.Next()
			if corpse == nil {
				break
			}

			if corpse.actor.Health <= 0 {
				diff := corpse.Body().Position.Sub(enemy.Body().Position)
				distSq := diff.LenSqr()
				if distSq < nearestCorpseDistance {
					dist := math2.Sqrt(distSq)
					hit, _ := gWorld.Raycast(
						enemy.actor.Position(),
						diff.Mul(1.0/dist),
						ColLayerMap,
						dist,
						enemy.Body(),
					)
					if !hit.Hit {
						dist = nearestCorpseDistance
						nearestCorpseHandle = handle
					}
				}
			}
		}
		if !nearestCorpseHandle.IsNil() {
			enemy.targetHandle = nearestCorpseHandle
		} else {
			enemy.targetHandle = gWorld.CurrentPlayer.Handle
		}
	}
}

func motherWraithUpdateChase(enemy *Enemy, deltaTime float32) {
	enemy.stalk(deltaTime, 1.0)
	enemy.attackTimer -= deltaTime
	if enemy.attackTimer <= 0.0 {
		hit, _ := gWorld.Raycast(
			enemy.actor.Position(),
			enemy.dirToTarget,
			ColLayerMap|ColLayerNPCs,
			enemy.distToTarget,
			enemy.Body(),
		)
		if !hit.Hit {
			enemy.changeState(&enemy.attackState)
		} else {
			// Reset focus to the player when a corpse is out of reach that needs reviving.
			enemy.targetHandle = gWorld.CurrentPlayer.Handle
		}
	}
}

func motherWraithEnterAttack(enemy *Enemy, oldState *enemyState) {
	enemy.attackTimer = 0.0
	enemy.faceTarget()
}

func motherWraithUpdateAttack(enemy *Enemy, deltaTime float32) {
	enemy.actor.inputForward, enemy.actor.inputStrafe = 0.0, 0.0
	if enemy.AnimPlayer.HitTriggerFrame(0) {
		enemy.faceTarget()
		SpawnBlessing(enemy.actor.Position(), enemy.actor.FacingVec(), enemy.id.Handle)
	}
	if enemy.AnimPlayer.IsAtEnd() {
		enemy.changeState(&enemy.chaseState)
	}
}

/**============================================
 *               Dummkopf
 *=============================================**/

func configureDummkopf(enemy *Enemy) (params enemyConfig) {
	params.texture = cache.GetTexture("assets/textures/sprites/dummkopf.png")
	idleAnim, _ := params.texture.GetAnimation("idle;front")
	unwakeAnim, _ := params.texture.GetAnimation("unwake;front")
	attackStartAnim, _ := params.texture.GetAnimation("attack start;front")
	stunAnim, _ := params.texture.GetAnimation("hurt;front")
	dieAnim, _ := params.texture.GetAnimation("die;front")
	params.defaultAnim = idleAnim
	params.bloodColor = color.FromBytes(65, 255, 0, 255)

	enemy.idleState = enemyState{
		anim:       unwakeAnim,
		leaveSound: cache.GetSfx("assets/sounds/enemy/dummkopf/dummkopf_greeting.wav"),
	}
	enemy.chaseState = enemyState{
		enterFunc:  dummkopfEnterChase,
		updateFunc: dummkopfUpdateChase,
	}
	enemy.stunState = enemyState{
		enterSound: cache.GetSfx("assets/sounds/enemy/dummkopf/dummkopf_hurt.wav"),
		anim:       stunAnim,
	}
	enemy.attackState = enemyState{
		anim:       attackStartAnim,
		enterFunc:  dummkopfEnterAttack,
		updateFunc: dummkopfUpdateAttack,
	}
	enemy.dieState = enemyState{
		enterSound: cache.GetSfx("assets/sounds/enemy/dummkopf/dummkopf_die.wav"),
		anim:       dieAnim,
	}

	reviveAnim, _ := params.texture.GetAnimation("revive;front")
	enemy.reviveState = enemyState{
		enterSound: cache.GetSfx("assets/sounds/enemy/dummkopf/dummkopf_revive.wav"),
		anim:       reviveAnim,
	}

	enemy.actor.MaxHealth = 250.0
	enemy.StunChance = 0.25
	enemy.spawnAmmo = game.AmmoTypePlasma
	enemy.spawnAmmoChance = 0.5

	return
}

func dummkopfEnterChase(enemy *Enemy, oldState *enemyState) {
	if oldState == &enemy.attackState {
		enemy.attackTimer = 1.0
	} else {
		enemy.attackTimer = 0.5
		wakeAnim, _ := enemy.SpriteRender.Texture().GetAnimation("wake;front")
		enemy.AnimPlayer.PlayNewAnim(wakeAnim)
	}
	enemy.faceTarget()
}

func dummkopfUpdateChase(enemy *Enemy, deltaTime float32) {
	enemy.attackTimer -= deltaTime
	if enemy.attackTimer <= 0.0 {
		hit, _ := gWorld.Raycast(
			enemy.actor.Position(),
			enemy.dirToTarget,
			ColLayerMap|ColLayerNPCs,
			enemy.distToTarget,
			enemy.Body(),
		)
		if !hit.Hit {
			enemy.changeState(&enemy.attackState)
		}
	}
}

func dummkopfEnterAttack(enemy *Enemy, oldState *enemyState) {
	enemy.attackTimer = 0.0
	enemy.faceTarget()
}

func dummkopfUpdateAttack(enemy *Enemy, deltaTime float32) {
	attackLoopAnim, _ := enemy.SpriteRender.Texture().GetAnimation("attack;front")
	attackEndAnim, _ := enemy.SpriteRender.Texture().GetAnimation("attack end;front")
	if enemy.AnimPlayer.IsPlayingAnim(enemy.attackState.anim) && enemy.AnimPlayer.IsAtEnd() {
		enemy.AnimPlayer.PlayNewAnim(attackLoopAnim)
		enemy.attackTimer = 0.0
	} else if enemy.AnimPlayer.IsPlayingAnim(attackLoopAnim) {
		enemy.attackTimer -= deltaTime
		if enemy.attackTimer < 0.0 {
			enemy.attackTimer = 0.15
			enemy.faceTarget()
			enemy.voice = cache.GetSfx("assets/sounds/enemy/dummkopf/dummkopf_spit.wav").
				PlayAttenuatedV(enemy.actor.Position())
			SpawnPlasmaBall(
				enemy.actor.Position().Add(mgl32.Vec3{0.0, 0.25, 0.0}),
				enemy.actor.FacingVec(),
				enemy.id.Handle,
				true,
			)
		}
		if enemy.stateTimer > 1.6 {
			enemy.AnimPlayer.PlayNewAnim(attackEndAnim)
		}
	} else if enemy.AnimPlayer.IsPlayingAnim(attackEndAnim) && enemy.AnimPlayer.IsAtEnd() {
		enemy.changeState(&enemy.chaseState)
	}
}

/**============================================
 *               Prisrak
 *=============================================**/

func configurePrisrak(enemy *Enemy) (params enemyConfig) {
	params.bloodColor = color.Yellow
	params.texture = cache.GetTexture("assets/textures/sprites/prisrak.png")
	floatAnim, _ := params.texture.GetAnimation("float;front")
	attackAnim, _ := params.texture.GetAnimation("attack;front")
	stunAnim, _ := params.texture.GetAnimation("hurt;front")
	dieAnim, _ := params.texture.GetAnimation("die;front")
	teleportAnim, _ := params.texture.GetAnimation("teleport;front")
	params.defaultAnim = floatAnim

	enemy.idleState = enemyState{
		anim:       floatAnim,
		stopAnim:   true,
		leaveSound: cache.GetSfx("assets/sounds/enemy/prisrak/prisrak_greeting.wav"),
	}
	enemy.chaseState = enemyState{
		anim:       floatAnim,
		enterFunc:  prisrakEnterChase,
		updateFunc: prisrakUpdateChase,
	}
	enemy.stunState = enemyState{
		enterSound: cache.GetSfx("assets/sounds/enemy/prisrak/prisrak_hurt.wav"),
		anim:       stunAnim,
	}
	enemy.attackState = enemyState{
		anim:       attackAnim,
		enterFunc:  fireWraithEnterAttack,
		updateFunc: prisrakUpdateAttack,
	}
	enemy.dodgeState = enemyState{
		anim:       teleportAnim,
		updateFunc: prisrakUpdateDodge,
	}
	enemy.dieState = enemyState{
		enterSound: cache.GetSfx("assets/sounds/enemy/prisrak/prisrak_die.wav"),
		anim:       dieAnim,
	}

	reviveAnim, _ := params.texture.GetAnimation("revive;front")
	enemy.reviveState = enemyState{
		enterSound: cache.GetSfx("assets/sounds/enemy/prisrak/prisrak_revive.wav"),
		anim:       reviveAnim,
	}

	enemy.defaultCollisionFilters = ColLayerMap | ColLayerActors

	enemy.actor.MaxSpeed = 4.0
	enemy.actor.MaxHealth = 225.0
	enemy.actor.body.Shape = collision.NewBoxShape(0.6, 0.5, 0.6)
	enemy.StunChance = 0.5

	return
}

func prisrakEnterChase(enemy *Enemy, oldState *enemyState) {
	enemy.attackTimer = rand.Float32() + 1.0
}

func prisrakUpdateChase(enemy *Enemy, deltaTime float32) {
	enemy.stalk(deltaTime, 1.0)
	enemy.attackTimer -= deltaTime
	if enemy.attackTimer <= 0.0 {
		hit, _ := gWorld.Raycast(
			enemy.actor.Position(),
			enemy.dirToTarget,
			ColLayerMap|ColLayerNPCs,
			enemy.distToTarget,
			enemy.Body(),
		)
		if !hit.Hit {
			enemy.changeState(&enemy.attackState)
		}
	}
}

func prisrakUpdateAttack(enemy *Enemy, deltaTime float32) {
	enemy.actor.inputForward, enemy.actor.inputStrafe = 0.0, 0.0
	for triggerIndex := range enemy.AnimPlayer.CurrentAnimation().TriggerFrames {
		if enemy.AnimPlayer.HitTriggerFrame(triggerIndex) {
			enemy.faceTarget()
			shootDir := enemy.actor.FacingVec()
			const spreadAngle = 22.5
			angle := math2.Degrees((-spreadAngle * 1.5) + (float32(triggerIndex) * spreadAngle) + (rand.Float32() * spreadAngle))
			shootDir = mgl32.TransformNormal(shootDir, mgl32.HomogRotate3DY(float32(math2.ToRadians(angle))))
			SpawnFireball(enemy.actor.Position(), shootDir, enemy.id.Handle)
		}
	}
	if enemy.AnimPlayer.IsAtEnd() {
		enemy.changeState(&enemy.dodgeState)
	}
}

func prisrakUpdateDodge(enemy *Enemy, deltaTime float32) {
	appearAnim, _ := enemy.SpriteRender.Texture().GetAnimation("appear")
	if enemy.AnimPlayer.IsAtEnd() {
		if !strings.HasPrefix(enemy.AnimPlayer.CurrentAnimation().Name, "appear") {
			leftRay := enemy.actor.FacingVec().Mul(12.0)
			rightRay := leftRay
			leftRay[0], leftRay[2] = -leftRay[2], leftRay[0]
			rightRay[0], rightRay[2] = rightRay[2], -rightRay[0]
			// Enlarged shape used to avoid getting stuck
			sweepShape := enemy.Body().Shape.ShrunkenBy(-0.1)

			farthestCast := collision.Result{Distance: 0.0, Position: enemy.actor.Position()}
			for _, ray := range [...]mgl32.Vec3{leftRay, rightRay} {
				sweepResult, _ := gWorld.GameMap.GridShape.SweepAgainst(mgl32.Vec3{}, enemy.actor.Position(), ray, sweepShape, ColLayerMap)
				if sweepResult.Distance > farthestCast.Distance {
					farthestCast = sweepResult
				}
			}
			farthestCast.Position[1] = enemy.actor.Position()[1]

			SpawnEffect(enemy.Body().Position, 2.0, WarpParticles(0.5, farthestCast.Position.Sub(enemy.Body().Position)))

			enemy.Body().Position = farthestCast.Position
			enemy.AnimPlayer.PlayNewAnim(appearAnim)
		} else {
			enemy.changeState(&enemy.chaseState)
		}
	}
}
