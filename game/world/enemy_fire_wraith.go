package world

import (
	"math/rand"

	"github.com/go-gl/mathgl/mgl32"
	"tophatdemon.com/total-invasion-ii/engine/assets/cache"
	"tophatdemon.com/total-invasion-ii/engine/color"
	"tophatdemon.com/total-invasion-ii/engine/math2/collision"
	"tophatdemon.com/total-invasion-ii/engine/scene"
	"tophatdemon.com/total-invasion-ii/engine/scene/comps"
	"tophatdemon.com/total-invasion-ii/game/world/effects"
)

const (
	SFX_FIRE_WRAITH_WAKE = "assets/sounds/enemy/fire_wraith/fire_wraith_greeting.wav"
	SFX_FIRE_WRAITH_HURT = "assets/sounds/enemy/fire_wraith/fire_wraith_hurt.wav"
	SFX_FIRE_WRAITH_DIE  = "assets/sounds/enemy/fire_wraith/fire_wraith_die.wav"
)

func SpawnFireWraith(world *World, position, angles mgl32.Vec3) (id scene.Id[*Enemy], enemy *Enemy, err error) {
	id, enemy, err = world.Enemies.New()
	if err != nil {
		return
	}

	enemy.initDefaults(world, id)
	enemy.bloodParticles = effects.Blood(15, color.Blue, 0.5)
	enemy.bloodParticles.Init()

	texture := cache.GetTexture("assets/textures/sprites/fire_wraith.png")
	walkAnim, _ := texture.GetAnimation("walk;front")
	attackAnim, _ := texture.GetAnimation("attack;front")
	stunAnim, _ := texture.GetAnimation("hurt;front")
	dieAnim, _ := texture.GetAnimation("die;front")

	enemy.states = [...]enemyState{
		ENEMY_STATE_IDLE: {
			anim:       walkAnim,
			stopAnim:   true,
			leaveSound: cache.GetSfx(SFX_FIRE_WRAITH_WAKE),
		},
		ENEMY_STATE_CHASE: {
			anim: walkAnim,
			enterFunc: func(oldState EnemyState) {
				enemy.attackTimer = rand.Float32() + 0.5
			},
			updateFunc: func(deltaTime float32) {
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
						enemy.changeState(ENEMY_STATE_ATTACK)
					}
				}
			},
		},
		ENEMY_STATE_STUN: {
			enterSound: cache.GetSfx(SFX_FIRE_WRAITH_HURT),
			anim:       stunAnim,
		},
		ENEMY_STATE_ATTACK: {
			anim: attackAnim,
			enterFunc: func(oldState EnemyState) {
				enemy.attackTimer = 0.0
				enemy.faceTarget()
			},
			updateFunc: func(deltaTime float32) {
				enemy.actor.inputForward, enemy.actor.inputStrafe = 0.0, 0.0
				if enemy.AnimPlayer.HitTriggerFrame(0) {
					enemy.faceTarget()
					SpawnFireball(enemy.world, enemy.actor.Position(), mgl32.Vec3{0.0, enemy.actor.YawAngle, 0.0}, enemy.id.Handle)
				}
				if enemy.AnimPlayer.IsAtEnd() {
					enemy.changeState(ENEMY_STATE_CHASE)
				}
			},
		},
		ENEMY_STATE_DIE: {
			enterSound: cache.GetSfx(SFX_FIRE_WRAITH_DIE),
			anim:       dieAnim,
		},
	}

	enemy.actor = Actor{
		body: comps.Body{
			Transform: comps.TransformFromTranslationAnglesScale(
				mgl32.Vec3(position).Add(mgl32.Vec3{0.0, -0.1, 0.0}), mgl32.Vec3{}, mgl32.Vec3{0.9, 0.9, 0.9},
			),
			Shape:  collision.NewSphere(0.7),
			Layer:  COL_LAYER_ACTORS | COL_LAYER_NPCS,
			Filter: COL_FILTER_FOR_ACTORS,
			LockY:  true,
		},
		YawAngle:  angles[1],
		AccelRate: 80.0,
		Friction:  20.0,
		MaxSpeed:  6.0,
		world:     world,
	}
	enemy.SpriteRender = comps.NewSpriteRender(texture)
	enemy.AnimPlayer = comps.NewAnimationPlayer(walkAnim, false)
	enemy.WakeTime = 0.5
	enemy.WakeLimit = 5.0
	enemy.StunChance = 1.0
	enemy.StunTime = 0.5
	enemy.chaseTimer = rand.Float32() * 10.0
	enemy.actor.Health = 175.0
	enemy.actor.MaxHealth = enemy.actor.Health

	return
}
