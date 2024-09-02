package world

import (
	"math"
	"math/rand"

	"github.com/go-gl/mathgl/mgl32"
	"tophatdemon.com/total-invasion-ii/engine/assets/cache"
	"tophatdemon.com/total-invasion-ii/engine/assets/textures"
	"tophatdemon.com/total-invasion-ii/engine/color"
	"tophatdemon.com/total-invasion-ii/engine/math2"
	"tophatdemon.com/total-invasion-ii/engine/math2/collision"
	"tophatdemon.com/total-invasion-ii/engine/render"
	"tophatdemon.com/total-invasion-ii/engine/scene/comps"
	"tophatdemon.com/total-invasion-ii/engine/tdaudio"
)

const (
	ENEMY_FOV_RADS       = math.Pi
	ENEMY_WAKE_PROXIMITY = 1.7
)

type EnemyState uint8

const (
	ENEMY_STATE_IDLE EnemyState = iota
	ENEMY_STATE_CHASE
	ENEMY_STATE_STUN
	ENEMY_STATE_ATTACK
	ENEMY_STATE_DIE
	ENEMY_STATE_COUNT
)

type enemyState struct {
	anim                   textures.Animation
	stopAnim               bool
	enterSound, leaveSound tdaudio.SoundId
	updateFunc             func(deltaTime float32)
	enterFunc              func(oldState EnemyState)
	leaveFunc              func(newState EnemyState)
}

type Enemy struct {
	SpriteRender                      comps.SpriteRender
	AnimPlayer                        comps.AnimationPlayer
	WakeTime                          float32 // Number of seconds player must be in sight before enemy begins to pursue.
	WakeLimit                         float32 // Maximum number of seconds after losing sight of player before giving up.
	StunTime                          float32 // Number of seconds the enemy stays in the 'stunned' state after getting hurt.
	StunChance                        float32 // The probability from 0 to 1 of the enemy getting stunned when hurt.
	bloodParticles                    comps.ParticleRender
	bloodOffset                       mgl32.Vec3
	actor                             Actor
	world                             *World
	states                            [ENEMY_STATE_COUNT]enemyState
	wakeTimer, chaseTimer, stateTimer float32
	chaseStrafeDir                    float32 // 1.0 to strafe right, -1.0 to strafe left while chasing player.
	spriteAngle                       float32 // Yaw angle on the Y axis determining where the sprite faces. Sometimes corresponds with actor.YawAngle
	state, previousState              EnemyState
	voice                             tdaudio.VoiceId

	// Player or target tracking variables
	dirToTarget  mgl32.Vec3
	distToTarget float32
	canSeeTarget bool
}

var _ HasActor = (*Enemy)(nil)
var _ comps.HasBody = (*Enemy)(nil)

func (enemy *Enemy) Actor() *Actor {
	return &enemy.actor
}

func (enemy *Enemy) Body() *comps.Body {
	return &enemy.actor.body
}

func (enemy *Enemy) initBlood(maxCount int, color color.Color, radius float32) {
	bloodTexture := cache.GetTexture("assets/textures/sprites/blood.png")
	bloodAnim, _ := bloodTexture.GetAnimation("default")
	enemy.bloodParticles = comps.ParticleRender{
		Mesh:             cache.QuadMesh,
		Texture:          bloodTexture,
		SpawnRate:        0.01,
		SpawnRadius:      radius,
		VisibilityRadius: 5.0,
		EmissionTimer:    0.0,
		MaxCount:         maxCount,
		SpawnFunc: func(index int, form *comps.ParticleForm, info *comps.ParticleInfo) {
			form.Color = color.Vector()
			s := rand.Float32()*0.10 + 0.15
			form.Size = mgl32.Vec2{s, s}
			info.Velocity = info.Velocity.Mul(rand.Float32()*5 + 1.0)
			info.Acceleration = mgl32.Vec3{0.0, -20.0, 0.0}
			info.Lifetime = 1.0
			info.AnimPlayer = comps.NewAnimationPlayer(bloodAnim, false)
			info.AnimPlayer.MoveToRandomFrame()
		},
		UpdateFunc: func(deltaTime float32, form *comps.ParticleForm, info *comps.ParticleInfo) {
			const SHRINK_RATE = 0.75
			form.Size[0] -= deltaTime * SHRINK_RATE
			form.Size[1] -= deltaTime * SHRINK_RATE
			if form.Size[0] <= 0.1 {
				form.Size = mgl32.Vec2{}
				info.Lifetime = 0.0
			}
		},
	}
	enemy.bloodParticles.Init()
}

func (enemy *Enemy) initDefaults(world *World) {
	enemy.world = world
	enemy.state = ENEMY_STATE_IDLE
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
	var vecToTarget mgl32.Vec3
	if player, ok := enemy.world.CurrentPlayer.Get(); ok {
		vecToTarget = player.Body().Transform.Position().Sub(enemyPos)
		enemy.distToTarget = vecToTarget.Len()
		if enemy.distToTarget != 0.0 {
			enemy.dirToTarget = vecToTarget.Normalize()
		}

		if enemy.distToTarget < ENEMY_WAKE_PROXIMITY {
			enemy.canSeeTarget = true
		} else if angle := math2.Acos(enemy.dirToTarget.Dot(enemyDir)); angle < ENEMY_FOV_RADS/2.0 {
			res, handle := enemy.world.Raycast(enemyPos, enemy.dirToTarget, COL_LAYER_PLAYERS|COL_LAYER_MAP, 25.0, enemy)
			if handle.Equals(player.id.Handle) && res.Hit {
				enemy.canSeeTarget = true
			}
		}
	}

	if !enemy.canSeeTarget {
		enemy.wakeTimer = max(0.0, enemy.wakeTimer-deltaTime)
	} else {
		enemy.wakeTimer = min(enemy.WakeLimit, enemy.wakeTimer+deltaTime)
	}

	if enemy.actor.Health <= 0.0 {
		enemy.changeState(ENEMY_STATE_DIE)
	}

	enemy.spriteAngle = enemy.actor.YawAngle

	// Default state updates
	switch enemy.state {
	case ENEMY_STATE_IDLE:
		if enemy.wakeTimer >= enemy.WakeTime {
			enemy.changeState(ENEMY_STATE_CHASE)
		}
		enemy.actor.inputForward = 0.0
		enemy.actor.inputStrafe = 0.0
	case ENEMY_STATE_CHASE:
		if enemy.wakeTimer <= 0.0 && !enemy.canSeeTarget {
			enemy.changeState(ENEMY_STATE_IDLE)
		}
	case ENEMY_STATE_STUN:
		enemy.actor.inputForward, enemy.actor.inputStrafe = 0.0, 0.0
		if enemy.stateTimer > enemy.StunTime {
			enemy.wakeTimer = enemy.WakeLimit
			enemy.changeState(ENEMY_STATE_CHASE)
		}
	case ENEMY_STATE_DIE:
		enemy.actor.inputForward, enemy.actor.inputStrafe = 0.0, 0.0
		radius := enemy.Body().Shape.(collision.Sphere).Radius()
		if enemy.bloodOffset.Y() > -radius {
			enemy.bloodOffset = enemy.bloodOffset.Sub(mgl32.Vec3{0.0, deltaTime, 0.0})
		} else {
			enemy.bloodOffset = mgl32.Vec3{0.0, -radius, 0.0}
		}
	}

	// Call custom defined state updates
	if updateFunc := enemy.states[enemy.state].updateFunc; updateFunc != nil {
		updateFunc(deltaTime)
	}

	enemy.stateTimer += deltaTime
}

func (enemy *Enemy) Render(context *render.Context) {
	enemy.SpriteRender.Render(&enemy.Body().Transform, &enemy.AnimPlayer, context, enemy.spriteAngle)
	enemy.bloodParticles.Render(&enemy.Body().Transform, context)
}

func (enemy *Enemy) ProcessSignal(signal any) {
}

func (enemy *Enemy) changeState(newStateId EnemyState) {
	if newStateId == enemy.state {
		return
	}

	newState := enemy.states[newStateId]
	oldState := enemy.states[enemy.state]

	if leaveSound := oldState.leaveSound; leaveSound.IsValid() {
		enemy.voice.Stop()
		enemy.voice = leaveSound.PlayAttenuatedV(enemy.actor.Position())
	}
	if oldState.leaveFunc != nil {
		oldState.leaveFunc(newStateId)
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
		newState.enterFunc(enemy.state)
	} else if newStateId == ENEMY_STATE_DIE {
		enemy.actor.body.Layer = COL_LAYER_NONE
		enemy.actor.body.Filter = COL_LAYER_MAP
		enemy.bloodParticles.EmissionTimer = newState.anim.Duration()
	}

	enemy.stateTimer = 0.0
	enemy.previousState = enemy.state
	enemy.state = newStateId
}

func (enemy *Enemy) OnDamage(sourceEntity any, damage float32) bool {
	if enemy.state == ENEMY_STATE_DIE {
		return false
	}
	if proj, ok := sourceEntity.(*Projectile); ok && proj.Body().OnLayer(COL_LAYER_PROJECTILES) {
		enemy.bloodParticles.EmissionTimer = 0.1
		enemy.actor.Health -= damage
		if enemy.state != ENEMY_STATE_STUN {
			if rand.Float32() < enemy.StunChance*proj.StunChance {
				enemy.changeState(ENEMY_STATE_STUN)
			} else {
				enemy.wakeTimer = enemy.WakeLimit
				if enemy.state == ENEMY_STATE_IDLE {
					enemy.changeState(ENEMY_STATE_CHASE)
				}
			}
		}
	}
	return true
}

func (enemy *Enemy) chase(
	deltaTime float32,
	chaseStraightTime float32, // Number of seconds enemy chases in a straight line before turning.
	chaseStrafeTime float32, // Number of seconds enemy chases diagonally.
) {
	totalChaseTime := chaseStraightTime + chaseStrafeTime
	enemy.actor.inputForward = 1.0
	enemy.chaseTimer += deltaTime
	enemy.actor.YawAngle = math2.Atan2(-enemy.dirToTarget.X(), -enemy.dirToTarget.Z())
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
			COL_LAYER_MAP,
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
