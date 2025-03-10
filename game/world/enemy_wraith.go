package world

import (
	"tophatdemon.com/total-invasion-ii/engine/assets/cache"
	"tophatdemon.com/total-invasion-ii/engine/color"
	"tophatdemon.com/total-invasion-ii/game/settings"
)

const (
	WRAITH_MELEE_RANGE = 2.5
)

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
	if enemy.distToTarget < WRAITH_MELEE_RANGE {
		enemy.changeState(&enemy.attackState)
	}
}

func wraithAttackUpdate(enemy *Enemy, deltaTime float32) {
	enemy.actor.inputForward, enemy.actor.inputStrafe = 0.0, 0.0
	if enemy.AnimPlayer.HitATriggerFrame() {
		if enemy.distToTarget >= WRAITH_MELEE_RANGE {
			enemy.changeState(&enemy.chaseState)
		} else if player, ok := enemy.world.CurrentPlayer.Get(); ok {
			player.OnDamage(enemy, settings.CurrDifficulty().WraithMeleeDamage)
		}
	}
}
