package world

import (
	"github.com/go-gl/mathgl/mgl32"
	"tophatdemon.com/total-invasion-ii/engine/assets/cache"
	"tophatdemon.com/total-invasion-ii/engine/color"
	"tophatdemon.com/total-invasion-ii/engine/math2"
	"tophatdemon.com/total-invasion-ii/engine/math2/collision"
	"tophatdemon.com/total-invasion-ii/engine/scene"
	"tophatdemon.com/total-invasion-ii/engine/scene/comps"
)

const (
	SFX_CHICKEN = "assets/sounds/chicken.wav"
)

func SpawnChicken(storage *scene.Storage[Enemy], position, angles mgl32.Vec3, world *World) (id scene.Id[*Enemy], enemy *Enemy, err error) {
	id, enemy, err = storage.New()
	if err != nil {
		return
	}

	enemy.initDefaults(world)
	enemy.initBlood(5, color.Red, 0.3)

	tex := cache.GetTexture("assets/textures/sprites/chicken.png")
	walkAnim, _ := tex.GetAnimation("walk;front")
	flyAnim, _ := tex.GetAnimation("fly;front")
	dieAnim, _ := tex.GetAnimation("die;front")

	enemy.states = [...]enemyState{
		ENEMY_STATE_IDLE: {
			anim:       walkAnim,
			leaveSound: cache.GetSfx(SFX_CHICKEN),
		},
		ENEMY_STATE_CHASE: {
			anim: walkAnim,
			updateFunc: func(deltaTime float32) {
				enemy.actor.inputForward = 1.0
			},
		},
		ENEMY_STATE_STUN: {
			anim:       flyAnim,
			enterSound: cache.GetSfx(SFX_CHICKEN),
		},
		ENEMY_STATE_ATTACK: {
			anim: walkAnim,
		},
		ENEMY_STATE_DIE: {
			anim:       dieAnim,
			enterSound: cache.GetSfx(SFX_CHICKEN),
		},
	}

	enemy.actor = Actor{
		body: comps.Body{
			Transform: comps.TransformFromTranslationAnglesScale(
				mgl32.Vec3(position), math2.DegToRadVec3(angles), mgl32.Vec3{0.4, 0.4, 0.4},
			),
			Shape:  collision.NewSphere(0.5),
			Layer:  COL_LAYER_ACTORS | COL_LAYER_NPCS,
			Filter: COL_FILTER_FOR_ACTORS,
			LockY:  false,
		},
		YawAngle:  mgl32.DegToRad(angles[1]),
		AccelRate: 80.0,
		Friction:  20.0,
		MaxSpeed:  2.5,
	}
	enemy.SpriteRender = comps.NewSpriteRender(tex)
	enemy.AnimPlayer = comps.NewAnimationPlayer(walkAnim, false)

	// This will make it so that the chicken never stays in the idle state.
	enemy.WakeTime = 0.0
	enemy.WakeLimit = math2.Inf32()
	enemy.wakeTimer = math2.Inf32()

	enemy.StunChance = 1.0
	enemy.StunTime = 0.5

	enemy.actor.Health = 10.0
	enemy.actor.MaxHealth = enemy.actor.Health

	return
}
