package world

import (
	"math"
	"math/rand"

	"github.com/go-gl/mathgl/mgl32"
	"tophatdemon.com/total-invasion-ii/engine/assets/te3"
	"tophatdemon.com/total-invasion-ii/engine/assets/textures"
	"tophatdemon.com/total-invasion-ii/engine/color"
	"tophatdemon.com/total-invasion-ii/engine/math2"
	"tophatdemon.com/total-invasion-ii/engine/math2/collision"
	"tophatdemon.com/total-invasion-ii/engine/render"
	"tophatdemon.com/total-invasion-ii/engine/scene"
	"tophatdemon.com/total-invasion-ii/engine/scene/comps"
	"tophatdemon.com/total-invasion-ii/engine/tdaudio"
	"tophatdemon.com/total-invasion-ii/game"
	"tophatdemon.com/total-invasion-ii/game/settings"
	"tophatdemon.com/total-invasion-ii/game/world/effects"
)

const (
	ENEMY_FOV_RADS         = math.Pi
	ENEMY_WAKE_PROXIMITY   = 1.7
	ENEMY_COL_LAYERS       = COL_LAYER_ACTORS | COL_LAYER_NPCS
	ENEMY_NOTICE_PROXIMITY = 25.0
)

type Enemy struct {
	SpriteRender                                   comps.SpriteRender
	AnimPlayer                                     comps.AnimationPlayer
	WakeTime                                       float32 // Number of seconds player must be in sight before enemy begins to pursue.
	WakeLimit                                      float32 // Maximum number of seconds after losing sight of player before giving up.
	StunTime                                       float32 // Number of seconds the enemy stays in the 'stunned' state after getting hurt.
	StunChance                                     float32 // The probability from 0 to 1 of the enemy getting stunned when hurt.
	bloodParticles                                 comps.ParticleRender
	bloodOffset                                    mgl32.Vec3
	actor                                          Actor
	id                                             scene.Id[*Enemy]
	world                                          *World
	wakeTimer, chaseTimer, stateTimer, attackTimer float32
	chaseStrafeDir                                 float32 // 1.0 to strafe right, -1.0 to strafe left while chasing player.
	spriteAngle                                    float32 // Yaw angle on the Y axis determining where the sprite faces. Sometimes corresponds with actor.YawAngle
	idleState, chaseState, stunState               enemyState
	attackState, dieState, reviveState             enemyState
	state, previousState                           *enemyState
	voice                                          tdaudio.VoiceId

	// Player or target tracking variables
	targetHandle                scene.Handle
	dirToTarget                 mgl32.Vec3
	distToTarget                float32
	canSeeTarget, canHearTarget bool
	variant                     game.EnemyType
}

type enemyState struct {
	anim                   textures.Animation
	stopAnim               bool // Set to true to leave the animation on its first frame without playing it.
	enterSound, leaveSound tdaudio.SoundId
	updateFunc             func(enemy *Enemy, deltaTime float32)
	enterFunc              func(enemy *Enemy, oldState *enemyState)
	leaveFunc              func(enemy *Enemy, newState *enemyState)
}

type enemyConfig struct {
	bloodColor  color.Color
	texture     *textures.Texture
	defaultAnim textures.Animation
}

var enemyTypeConfigFuncs = [game.ENEMY_TYPE_COUNT]func(enemy *Enemy) enemyConfig{
	game.ENEMY_TYPE_WRAITH:        configureWraith,
	game.ENEMY_TYPE_FIRE_WRAITH:   configureFireWraith,
	game.ENEMY_TYPE_MOTHER_WRAITH: configureMotherWraith,
	game.ENEMY_TYPE_DUMMKOPF:      configureDummkopf,
}

var _ HasActor = (*Enemy)(nil)
var _ comps.HasBody = (*Enemy)(nil)

func SpawnEnemyFromTE3(world *World, ent te3.Ent) (scene.Id[*Enemy], *Enemy, error) {
	var variant game.EnemyType
	switch ent.Properties["enemy"] {
	case "fire wraith":
		variant = game.ENEMY_TYPE_FIRE_WRAITH
	case "mother wraith":
		variant = game.ENEMY_TYPE_MOTHER_WRAITH
	case "dummkopf":
		variant = game.ENEMY_TYPE_DUMMKOPF
	default:
		variant = game.ENEMY_TYPE_WRAITH
	}
	return SpawnEnemy(world, ent.Position, ent.AnglesInRadians(), variant)
}

func (enemy *Enemy) Actor() *Actor {
	return &enemy.actor
}

func (enemy *Enemy) Body() *comps.Body {
	return &enemy.actor.body
}

func SpawnEnemy(world *World, position, angles mgl32.Vec3, variant game.EnemyType) (id scene.Id[*Enemy], enemy *Enemy, err error) {
	id, enemy, err = world.Enemies.New()
	if err != nil {
		return
	}

	world.Hud.EnemiesTotal++
	enemy.world = world
	enemy.variant = variant
	enemy.state = &enemy.idleState
	enemy.id = id

	enemy.actor = Actor{
		body: comps.Body{
			Transform: comps.TransformFromTranslationAnglesScale(
				mgl32.Vec3(position).Add(mgl32.Vec3{0.0, -0.1, 0.0}), mgl32.Vec3{}, mgl32.Vec3{0.9, 0.9, 0.9},
			),
			Shape:  collision.NewSphere(0.7),
			Layer:  ENEMY_COL_LAYERS,
			Filter: COL_FILTER_FOR_ACTORS,
			LockY:  true,
		},
		YawAngle:  angles[1],
		AccelRate: 80.0,
		Friction:  20.0,
		MaxSpeed:  5.5,
		world:     world,
	}
	enemy.WakeTime = 0.5
	enemy.WakeLimit = 5.0
	enemy.StunChance = 1.0
	enemy.StunTime = 0.5
	enemy.chaseTimer = rand.Float32() * 10.0

	params := enemyTypeConfigFuncs[variant](enemy)

	enemy.bloodParticles = effects.Blood(15, params.bloodColor, 0.5)
	enemy.bloodParticles.Init()
	enemy.actor.MaxHealth *= settings.CurrDifficulty().EnemyHealthMultiplier
	enemy.actor.Health, enemy.actor.TargetHealth = enemy.actor.MaxHealth, enemy.actor.MaxHealth

	enemy.SpriteRender = comps.NewSpriteRender(params.texture)
	enemy.AnimPlayer = comps.NewAnimationPlayer(params.defaultAnim, false)

	return
}

func (enemy *Enemy) Finalize() {
	enemy.bloodParticles.Finalize()
}

func (enemy *Enemy) Update(deltaTime float32) {
	enemy.AnimPlayer.Update(deltaTime)
	enemy.actor.Update(deltaTime)

	bloodTransform := enemy.Body().Transform
	bloodTransform.TranslateV(enemy.bloodOffset)
	enemy.bloodParticles.Update(deltaTime, &bloodTransform)

	enemyPos := enemy.Body().Transform.Position()
	enemyDir := enemy.actor.FacingVec()
	if enemy.voice.IsValid() {
		enemy.voice.SetPositionV(enemyPos)
	}

	// Check if the player is in view and not obstructed
	enemy.canSeeTarget = false
	enemy.canHearTarget = false
	var vecToTarget mgl32.Vec3
	if enemy.targetHandle.IsNil() {
		enemy.targetHandle = enemy.world.CurrentPlayer.Handle
	}
	if targetActor, ok := scene.Get[HasActor](enemy.targetHandle); ok && enemy.world.IsOnPlayerCamera() {
		vecToTarget = targetActor.Body().Transform.Position().Sub(enemyPos)
		enemy.distToTarget = vecToTarget.Len()
		if enemy.distToTarget != 0.0 {
			enemy.dirToTarget = vecToTarget.Normalize()
		}

		inHearingRange := targetActor.Actor().noisyTimer > 0 && enemy.world.Hud.SelectedWeapon() != nil && enemy.distToTarget < enemy.world.Hud.SelectedWeapon().NoiseLevel()
		inFieldOfView := math2.Acos(enemy.dirToTarget.Dot(enemyDir)) < ENEMY_FOV_RADS/2.0
		if enemy.distToTarget < ENEMY_WAKE_PROXIMITY {
			enemy.canSeeTarget = true
		} else if inHearingRange || inFieldOfView {
			res, _ := enemy.world.Raycast(enemyPos, enemy.dirToTarget, COL_LAYER_MAP, enemy.distToTarget, nil)
			if !res.Hit && enemy.distToTarget < ENEMY_NOTICE_PROXIMITY {
				enemy.canSeeTarget = true
				enemy.canHearTarget = true
			}
		}
	} else if enemy.state != &enemy.dieState {
		enemy.wakeTimer = 0.0
		enemy.changeState(&enemy.idleState)
	}

	if enemy.canHearTarget {
		enemy.wakeTimer = enemy.WakeLimit
	} else if !enemy.canSeeTarget {
		enemy.wakeTimer = max(0.0, enemy.wakeTimer-deltaTime)
	} else {
		enemy.wakeTimer = min(enemy.WakeLimit, enemy.wakeTimer+deltaTime)
	}

	if enemy.actor.Health <= 0.0 && enemy.state != &enemy.reviveState {
		enemy.changeState(&enemy.dieState)
	}

	enemy.spriteAngle = enemy.actor.YawAngle

	// Default state updates
	switch enemy.state {
	case &enemy.idleState:
		if enemy.wakeTimer >= enemy.WakeTime {
			enemy.changeState(&enemy.chaseState)
		}
		enemy.actor.inputForward = 0.0
		enemy.actor.inputStrafe = 0.0
	case &enemy.chaseState:
		if enemy.wakeTimer <= 0.0 && !enemy.canSeeTarget {
			enemy.changeState(&enemy.idleState)
		}
	case &enemy.stunState:
		enemy.actor.inputForward, enemy.actor.inputStrafe = 0.0, 0.0
		if enemy.stateTimer > enemy.StunTime {
			enemy.wakeTimer = enemy.WakeLimit
			enemy.changeState(&enemy.chaseState)
		}
	case &enemy.dieState:
		enemy.actor.inputForward, enemy.actor.inputStrafe = 0.0, 0.0
		radius := enemy.Body().Shape.(collision.Sphere).Radius()
		if enemy.bloodOffset.Y() > -radius {
			enemy.bloodOffset = enemy.bloodOffset.Sub(mgl32.Vec3{0.0, deltaTime, 0.0})
		} else {
			enemy.bloodOffset = mgl32.Vec3{0.0, -radius, 0.0}
		}
	case &enemy.reviveState:
		enemy.actor.inputForward, enemy.actor.inputStrafe = 0.0, 0.0
		if enemy.AnimPlayer.IsAtEnd() {
			enemy.actor.Health = enemy.actor.TargetHealth
			enemy.changeState(&enemy.chaseState)
		}
	}

	// Call custom defined state updates
	if enemy.state != nil && enemy.state.updateFunc != nil {
		enemy.state.updateFunc(enemy, deltaTime)
	}

	enemy.stateTimer += deltaTime
}

func (enemy *Enemy) Render(context *render.Context) {
	enemy.SpriteRender.Render(&enemy.Body().Transform, &enemy.AnimPlayer, context, enemy.spriteAngle)
	enemy.bloodParticles.Render(&enemy.Body().Transform, context)
}

func (enemy *Enemy) ProcessSignal(signal any) {
}

func (enemy *Enemy) OnPlayerVictory() {
	enemy.WakeTime = math2.Inf32()
	enemy.WakeLimit = 0.0
	enemy.wakeTimer = 0.0
	if enemy.actor.Health > 0 {
		enemy.changeState(&enemy.idleState)
	}
}

func (enemy *Enemy) changeState(newState *enemyState) {
	if newState == enemy.state {
		return
	}

	oldState := enemy.state

	if oldState.leaveFunc != nil {
		oldState.leaveFunc(enemy, newState)
	} else if oldState == &enemy.dieState {
		// Ensure nobody's standing on top of the enemy that is getting revived.
		actorsIter := enemy.world.IterActorsInSphere(enemy.Body().Transform.Position(), enemy.Body().Shape.(collision.Sphere).Radius(), enemy)
		for {
			actor, _ := actorsIter.Next()
			if actor == nil {
				break
			}
			if actor.Actor().Health > 0 {
				return
			}
		}

		enemy.world.Hud.EnemiesKilled--
		enemy.actor.body.Layer = ENEMY_COL_LAYERS
		enemy.actor.body.Filter = COL_FILTER_FOR_ACTORS
	}
	if leaveSound := oldState.leaveSound; leaveSound.IsValid() {
		enemy.voice.Stop()
		enemy.voice = leaveSound.PlayAttenuatedV(enemy.actor.Position())
	}

	if newState.anim.Frames != nil {
		enemy.AnimPlayer.ChangeAnimation(newState.anim)
		if newState.stopAnim {
			enemy.AnimPlayer.Stop()
		} else {
			enemy.AnimPlayer.PlayFromStart()
		}
	}
	if newState.enterSound.IsValid() {
		enemy.voice.Stop()
		enemy.voice = newState.enterSound.PlayAttenuatedV(enemy.actor.Position())
	}

	// Initialize new state
	if newState.enterFunc != nil {
		newState.enterFunc(enemy, enemy.state)
	} else if newState == &enemy.dieState {
		enemy.world.Hud.EnemiesKilled++
		enemy.actor.body.Layer = COL_LAYER_NONE
		enemy.actor.body.Filter = COL_LAYER_MAP
		enemy.bloodParticles.EmissionTimer = newState.anim.Duration()
	}

	enemy.stateTimer = 0.0
	enemy.previousState = enemy.state
	enemy.state = newState
}

func (enemy *Enemy) OnDamage(sourceEntity any, damage float32) bool {
	if enemy.state == &enemy.dieState {
		return false
	}

	enemy.bloodParticles.EmissionTimer = 0.1
	enemy.actor.Health -= damage
	if enemy.actor.Health <= 0.0 {
		enemy.changeState(&enemy.dieState)
	} else if enemy.state != &enemy.stunState {
		sourceStunChance := float32(1.0)
		if proj, ok := sourceEntity.(*Projectile); ok {
			sourceStunChance = proj.StunChance
		}
		if rand.Float32() < enemy.StunChance*sourceStunChance {
			enemy.changeState(&enemy.stunState)
		} else {
			enemy.wakeTimer = enemy.WakeLimit
			if enemy.state == &enemy.idleState {
				enemy.changeState(&enemy.chaseState)
			}
		}
	}
	return true
}

func (enemy *Enemy) faceTarget() {
	enemy.actor.YawAngle = math2.Atan2(-enemy.dirToTarget.X(), -enemy.dirToTarget.Z())
}

func (enemy *Enemy) chase(
	deltaTime float32,
	chaseStraightTime float32, // Number of seconds enemy chases in a straight line before turning.
	chaseStrafeTime float32, // Number of seconds enemy chases diagonally.
) {
	totalChaseTime := chaseStraightTime + chaseStrafeTime
	enemy.actor.inputForward = 1.0
	enemy.chaseTimer += deltaTime
	enemy.faceTarget()
	if enemy.chaseTimer < chaseStraightTime {
		// First, walk forward for a bit
		enemy.chaseStrafeDir = 0.0
		enemy.spriteAngle = enemy.actor.YawAngle
	} else if enemy.chaseTimer < totalChaseTime {
		// Then turn in a random direction.
		if enemy.chaseStrafeDir == 0.0 {
			enemy.chaseStrafeDir = ([2]float32{-0.7, 0.7})[rand.Intn(2)]
		}
		enemy.spriteAngle = enemy.actor.YawAngle - (math2.Signum(enemy.chaseStrafeDir) * math.Pi / 2.0)

		// Cancel the turn if we are facing a wall
		hit, _ := enemy.world.Raycast(
			enemy.actor.Position(),
			mgl32.Vec3{-math2.Sin(enemy.spriteAngle), 0.0, -math2.Cos(enemy.spriteAngle)},
			COL_LAYER_MAP|COL_LAYER_ACTORS|COL_LAYER_INVISIBLE,
			WRAITH_MELEE_RANGE,
			enemy,
		)
		if hit.Hit {
			enemy.spriteAngle = enemy.actor.YawAngle
			enemy.chaseStrafeDir = 0.0
			enemy.chaseTimer = 0.0
		}
	} else {
		for enemy.chaseTimer > totalChaseTime {
			enemy.chaseTimer -= totalChaseTime
		}
	}
	enemy.actor.inputStrafe = enemy.chaseStrafeDir
}

// This will move the enemy in 4 directions relative to the player, only sometimes closing in on her position.
// Useful for ranged enemies.
func (enemy *Enemy) stalk(
	deltaTime float32,
	moveTime float32,
) {
	enemy.actor.inputForward = 1.0
	enemy.chaseTimer += deltaTime

	if enemy.chaseTimer >= moveTime {
		switch rand.Intn(4) {
		case 0:
			enemy.actor.YawAngle = math2.Atan2(-enemy.dirToTarget.X(), -enemy.dirToTarget.Z())
		case 1:
			enemy.actor.YawAngle = math2.Atan2(enemy.dirToTarget.X(), enemy.dirToTarget.Z())
		case 2:
			enemy.actor.YawAngle = math2.Atan2(-enemy.dirToTarget.Z(), enemy.dirToTarget.X())
		case 3:
			enemy.actor.YawAngle = math2.Atan2(enemy.dirToTarget.Z(), -enemy.dirToTarget.X())
		}
		enemy.chaseTimer = 0.0
	} else {
		// Cancel the movement if we are approaching an obstacle
		hit, _ := enemy.world.Raycast(
			enemy.actor.Position(),
			enemy.actor.FacingVec(),
			COL_LAYER_MAP|COL_LAYER_ACTORS|COL_LAYER_INVISIBLE,
			WRAITH_MELEE_RANGE,
			enemy,
		)
		if hit.Hit {
			enemy.chaseTimer = moveTime
		}
	}
}
