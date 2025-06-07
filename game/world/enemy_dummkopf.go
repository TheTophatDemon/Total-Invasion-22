package world

import (
	"tophatdemon.com/total-invasion-ii/engine/assets/cache"
	"tophatdemon.com/total-invasion-ii/engine/color"
)

func configureDummkopf(enemy *Enemy) (params enemyConfig) {
	params.texture = cache.GetTexture("assets/textures/sprites/dummkopf.png")
	idleAnim, _ := params.texture.GetAnimation("idle;front")
	wakeAnim, _ := params.texture.GetAnimation("wake;front")

	stunAnim, _ := params.texture.GetAnimation("hurt;front")
	dieAnim, _ := params.texture.GetAnimation("die;front")
	params.defaultAnim = idleAnim
	params.bloodColor = color.FromBytes(65, 255, 0, 255)

	enemy.idleState = enemyState{
		anim:       idleAnim,
		stopAnim:   true,
		leaveSound: cache.GetSfx("assets/sounds/enemy/dummkopf/dummkopf_greeting.wav"),
	}
	enemy.chaseState = enemyState{
		anim:       wakeAnim,
		updateFunc: nil,
	}
	enemy.stunState = enemyState{
		enterSound: cache.GetSfx("assets/sounds/enemy/dummkopf/dummkopf_hurt.wav"),
		anim:       stunAnim,
	}
	// enemy.attackState = enemyState{
	// 	anim:       attackAnim,
	// 	updateFunc: wraithAttackUpdate,
	// }
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

	return
}
