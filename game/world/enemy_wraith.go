package world

import (
	"math/rand"

	"github.com/go-gl/mathgl/mgl32"
	"tophatdemon.com/total-invasion-ii/engine/assets/cache"
	"tophatdemon.com/total-invasion-ii/engine/math2"
	"tophatdemon.com/total-invasion-ii/engine/math2/collision"
	"tophatdemon.com/total-invasion-ii/engine/scene"
	"tophatdemon.com/total-invasion-ii/engine/scene/comps"
	"tophatdemon.com/total-invasion-ii/game/settings"
)

const (
	SFX_WRAITH_WAKE = "assets/sounds/enemy/wraith/wraith_greeting.wav"
	SFX_WRAITH_HURT = "assets/sounds/enemy/wraith/wraith_hurt.wav"
	SFX_WRAITH_DIE  = "assets/sounds/enemy/wraith/wraith_die.wav"
)

const (
	WRAITH_MELEE_RANGE = 2.5
)

func SpawnWraith(storage *scene.Storage[Enemy], position, angles mgl32.Vec3, world *World) (id scene.Id[*Enemy], enemy *Enemy, err error) {
	id, enemy, err = storage.New()
	if err != nil {
		return
	}

	enemy.initDefaults(world)

	wraithTexture := cache.GetTexture("assets/textures/sprites/wraith.png")
	walkAnim, _ := wraithTexture.GetAnimation("walk;front")
	attackAnim, _ := wraithTexture.GetAnimation("attack;front")
	stunAnim, _ := wraithTexture.GetAnimation("hurt;front")
	dieAnim, _ := wraithTexture.GetAnimation("die;front")

	enemy.states = [...]enemyState{
		ENEMY_STATE_IDLE: {
			anim:       walkAnim,
			stopAnim:   true,
			leaveSound: cache.GetSfx(SFX_WRAITH_WAKE),
		},
		ENEMY_STATE_CHASE: {
			anim: walkAnim,
			updateFunc: func(deltaTime float32) {
				enemy.chase(deltaTime, 3.0, 1.0)
				if enemy.distToTarget < WRAITH_MELEE_RANGE {
					enemy.changeState(ENEMY_STATE_ATTACK)
				}
			},
		},
		ENEMY_STATE_STUN: {
			enterSound: cache.GetSfx(SFX_WRAITH_HURT),
			anim:       stunAnim,
		},
		ENEMY_STATE_ATTACK: {
			anim: attackAnim,
			updateFunc: func(deltaTime float32) {
				enemy.actor.inputForward, enemy.actor.inputStrafe = 0.0, 0.0
				if enemy.stateTimer > 0.5 {
					if enemy.distToTarget >= WRAITH_MELEE_RANGE {
						enemy.changeState(ENEMY_STATE_CHASE)
					} else if player, ok := enemy.world.CurrentPlayer.Get(); ok {
						player.OnDamage(enemy, settings.CurrDifficulty().WraithMeleeDamage)
					}
					enemy.stateTimer = 0.0
				}
			},
		},
		ENEMY_STATE_DIE: {
			enterSound: cache.GetSfx(SFX_WRAITH_DIE),
			anim:       dieAnim,
		},
	}

	enemy.actor = Actor{
		body: comps.Body{
			Transform: comps.TransformFromTranslationAnglesScale(
				mgl32.Vec3(position).Add(mgl32.Vec3{0.0, -0.1, 0.0}), math2.DegToRadVec3(angles), mgl32.Vec3{0.9, 0.9, 0.9},
			),
			Shape:  collision.NewSphere(0.7),
			Layer:  COL_LAYER_ACTORS | COL_LAYER_NPCS,
			Filter: COL_FILTER_FOR_ACTORS,
			LockY:  true,
		},
		YawAngle:  mgl32.DegToRad(angles[1]),
		AccelRate: 80.0,
		Friction:  20.0,
		MaxSpeed:  5.5,
	}
	enemy.SpriteRender = comps.NewSpriteRender(wraithTexture)
	enemy.AnimPlayer = comps.NewAnimationPlayer(walkAnim, false)
	enemy.WakeTime = 0.5
	enemy.WakeLimit = 5.0
	enemy.StunChance = 1.0
	enemy.StunTime = 0.5
	enemy.chaseTimer = rand.Float32() * 10.0
	enemy.actor.Health = 100.0
	enemy.actor.MaxHealth = enemy.actor.Health

	return
}
