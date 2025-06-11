package world

import (
	"github.com/go-gl/mathgl/mgl32"
	"tophatdemon.com/total-invasion-ii/engine/assets/cache"
	"tophatdemon.com/total-invasion-ii/engine/color"
	"tophatdemon.com/total-invasion-ii/game"
)

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
	enemy.spawnAmmo = game.AMMO_TYPE_PLASMA
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
