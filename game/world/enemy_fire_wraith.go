package world

import (
	"math/rand"

	"github.com/go-gl/mathgl/mgl32"
	"tophatdemon.com/total-invasion-ii/engine/assets/cache"
	"tophatdemon.com/total-invasion-ii/engine/color"
)

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
