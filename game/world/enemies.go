package world

import (
	"math"
	"math/rand"

	"github.com/go-gl/mathgl/mgl32"
	"tophatdemon.com/total-invasion-ii/engine/assets/cache"
	"tophatdemon.com/total-invasion-ii/engine/color"
	"tophatdemon.com/total-invasion-ii/engine/math2"
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
		} else if player, ok := enemy.world.CurrentPlayer.Get(); ok {
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
	enemy.actor.MaxHealth = 150.0

	return
}

func fireWraithEnterChase(enemy *Enemy, oldState *enemyState) {
	enemy.attackTimer = rand.Float32() + 0.5
}

func fireWraithUpdateChase(enemy *Enemy, deltaTime float32) {
	enemy.stalk(deltaTime, 1.0)
	enemy.attackTimer -= deltaTime
	if enemy.attackTimer <= 0.0 {
		hit, _ := enemy.world.Raycast(
			enemy.actor.Position(),
			enemy.dirToTarget,
			COL_LAYER_MAP|COL_LAYER_NPCS,
			enemy.distToTarget,
			enemy,
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
		SpawnFireball(enemy.world, enemy.actor.Position(), mgl32.Vec3{0.0, enemy.actor.YawAngle, 0.0}, enemy.id.Handle)
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

	return
}

func motherWraithEnterChase(enemy *Enemy, oldState *enemyState) {
	enemy.attackTimer = rand.Float32() + 1.5

	// Switch periodically between shooting at the player and shooting to revive nearby enemies.
	if rand.Float32() < 0.5 {
		enemy.targetHandle = enemy.world.CurrentPlayer.Handle
	} else {
		nearbyEnemiesIter := enemy.world.Enemies.Iter()
		var nearestCorpseHandle scene.Handle
		nearestCorpseDistance := float32(math.MaxFloat32)
		for {
			corpse, handle := nearbyEnemiesIter.Next()
			if corpse == nil {
				break
			}

			if corpse.actor.Health <= 0 {
				diff := corpse.Body().Transform.Position().Sub(enemy.Body().Transform.Position())
				distSq := diff.LenSqr()
				if distSq < nearestCorpseDistance {
					dist := math2.Sqrt(distSq)
					hit, _ := enemy.world.Raycast(
						enemy.actor.Position(),
						diff.Mul(1.0/dist),
						COL_LAYER_MAP,
						dist,
						enemy,
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
			enemy.targetHandle = enemy.world.CurrentPlayer.Handle
		}
	}
}

func motherWraithUpdateChase(enemy *Enemy, deltaTime float32) {
	enemy.stalk(deltaTime, 1.0)
	enemy.attackTimer -= deltaTime
	if enemy.attackTimer <= 0.0 {
		hit, _ := enemy.world.Raycast(
			enemy.actor.Position(),
			enemy.dirToTarget,
			COL_LAYER_MAP|COL_LAYER_NPCS,
			enemy.distToTarget,
			enemy,
		)
		if !hit.Hit {
			enemy.changeState(&enemy.attackState)
		} else {
			// Reset focus to the player when a corpse is out of reach that needs reviving.
			enemy.targetHandle = enemy.world.CurrentPlayer.Handle
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
		SpawnBlessing(enemy.world, enemy.actor.Position(), mgl32.Vec3{0.0, enemy.actor.YawAngle, 0.0}, enemy.id.Handle)
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
		hit, _ := enemy.world.Raycast(
			enemy.actor.Position(),
			enemy.dirToTarget,
			COL_LAYER_MAP|COL_LAYER_NPCS,
			enemy.distToTarget,
			enemy,
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
			SpawnPlasmaBall(enemy.world,
				enemy.actor.Position().Add(mgl32.Vec3{0.0, 0.25, 0.0}),
				mgl32.Vec3{0.0, enemy.actor.YawAngle, 0.0},
				enemy.id.Handle,
				true)
		}
		if enemy.stateTimer > 1.6 {
			enemy.AnimPlayer.PlayNewAnim(attackEndAnim)
		}
	} else if enemy.AnimPlayer.IsPlayingAnim(attackEndAnim) && enemy.AnimPlayer.IsAtEnd() {
		enemy.changeState(&enemy.chaseState)
	}
}
