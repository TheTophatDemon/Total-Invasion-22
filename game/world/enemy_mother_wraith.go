package world

import (
	"fmt"
	"math"
	"math/rand"

	"github.com/go-gl/mathgl/mgl32"
	"tophatdemon.com/total-invasion-ii/engine/assets/cache"
	"tophatdemon.com/total-invasion-ii/engine/color"
	"tophatdemon.com/total-invasion-ii/engine/math2"
	"tophatdemon.com/total-invasion-ii/engine/scene"
)

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
		fmt.Println("Now we lookin' at the playah")
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
			fmt.Println("Now we lookin' at the enumee")
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
