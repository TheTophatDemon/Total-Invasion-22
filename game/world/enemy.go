package world

import (
	"math"
	"math/rand"

	"github.com/go-gl/mathgl/mgl32"
	"tophatdemon.com/total-invasion-ii/engine/assets/audio"
	"tophatdemon.com/total-invasion-ii/engine/assets/cache"
	"tophatdemon.com/total-invasion-ii/engine/assets/textures"
	"tophatdemon.com/total-invasion-ii/engine/color"
	"tophatdemon.com/total-invasion-ii/engine/math2"
	"tophatdemon.com/total-invasion-ii/engine/math2/collision"
	"tophatdemon.com/total-invasion-ii/engine/render"
	"tophatdemon.com/total-invasion-ii/engine/scene"
	"tophatdemon.com/total-invasion-ii/engine/scene/comps"
)

const (
	ENEMY_FOV_RADS       = math.Pi
	ENEMY_WAKE_PROXIMITY = 1.7
	WRAITH_MELEE_RANGE   = 2.0
)

type EnemyState uint8

const (
	ENEMY_STATE_IDLE EnemyState = iota
	ENEMY_STATE_CHASE
	ENEMY_STATE_STUN
	ENEMY_STATE_ATTACK
	ENEMY_STATE_DIE
)

const (
	SFX_WRAITH_WAKE = "assets/sounds/enemy/wraith/wraith_greeting.wav"
	SFX_WRAITH_HURT = "assets/sounds/enemy/wraith/wraith_hurt.wav"
	SFX_WRAITH_DIE  = "assets/sounds/enemy/wraith/wraith_die.wav"
)

type Enemy struct {
	SpriteRender                            comps.SpriteRender
	AnimPlayer                              comps.AnimationPlayer
	WakeTime                                float32 // Number of seconds player must be in sight before enemy begins to pursue.
	WakeLimit                               float32 // Maximum number of seconds after losing sight of player before giving up.
	StunTime                                float32 // Number of seconds the enemy stays in the 'stunned' state after getting hurt.
	StunChance                              float32 // The probability from 0 to 1 of the enemy getting stunned when hurt.
	bloodParticles                          comps.ParticleRender
	bloodOffset                             mgl32.Vec3
	actor                                   Actor
	world                                   *World
	walkAnim, attackAnim, stunAnim, dieAnim textures.Animation
	wakeTimer, chaseTimer, stateTimer       float32
	chaseStrafeDir                          float32 // 1.0 to strafe right, -1.0 to strafe left while chasing player.
	spriteAngle                             float32 // Yaw angle on the Y axis determining where the sprite faces. Sometimes corresponds with actor.YawAngle
	state                                   EnemyState
	voice                                   audio.VoiceId
}

var _ HasActor = (*Enemy)(nil)
var _ comps.HasBody = (*Enemy)(nil)

func (enemy *Enemy) Actor() *Actor {
	return &enemy.actor
}

func (enemy *Enemy) Body() *comps.Body {
	return &enemy.actor.body
}

func SpawnEnemy(storage *scene.Storage[Enemy], position, angles mgl32.Vec3, world *World) (id scene.Id[*Enemy], enemy *Enemy, err error) {
	id, enemy, err = storage.New()
	if err != nil {
		return
	}

	enemy.world = world

	wraithTexture := cache.GetTexture("assets/textures/sprites/wraith.png")
	enemy.walkAnim, _ = wraithTexture.GetAnimation("walk;front")
	enemy.attackAnim, _ = wraithTexture.GetAnimation("attack;front")
	enemy.stunAnim, _ = wraithTexture.GetAnimation("hurt;front")
	enemy.dieAnim, _ = wraithTexture.GetAnimation("die;front")
	// Preload sounds
	cache.GetSfx(SFX_WRAITH_WAKE)
	cache.GetSfx(SFX_WRAITH_HURT)
	cache.GetSfx(SFX_WRAITH_DIE)

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
	enemy.AnimPlayer = comps.NewAnimationPlayer(enemy.walkAnim, false)
	enemy.WakeTime = 0.5
	enemy.WakeLimit = 5.0
	enemy.StunChance = 1.0
	enemy.StunTime = 0.5
	enemy.chaseTimer = rand.Float32() * 10.0
	enemy.actor.Health = 100.0

	enemy.state = ENEMY_STATE_IDLE

	bloodTexture := cache.GetTexture("assets/textures/sprites/blood.png")
	bloodAnim, _ := bloodTexture.GetAnimation("default")
	enemy.bloodParticles = comps.ParticleRender{
		Mesh:             cache.QuadMesh,
		Texture:          bloodTexture,
		SpawnRate:        0.01,
		SpawnRadius:      0.5,
		VisibilityRadius: 5.0,
		EmissionTimer:    0.0,
		SpawnFunc: func(index int, form *comps.ParticleForm, info *comps.ParticleInfo) {
			form.Color = color.Red.Vector()
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
	enemy.bloodParticles.Init(25)

	return
}

func (enemy *Enemy) Update(deltaTime float32) {
	enemy.AnimPlayer.Update(deltaTime)
	enemy.actor.Update(deltaTime)

	bloodTransform := enemy.Body().Transform
	bloodTransform.TranslateV(enemy.bloodOffset)
	enemy.bloodParticles.Update(deltaTime, &bloodTransform)

	enemyPos := enemy.Body().Transform.Position()
	enemyDir := enemy.actor.FacingVec()

	// Check if the player is in view and not obstructed
	canSeePlayer := false
	var vecToPlayer, dirToPlayer mgl32.Vec3
	var distToPlayer float32
	if player, ok := enemy.world.CurrentPlayer.Get(); ok {
		enemy.voice.Attenuate(enemyPos, player.Body().Transform.Matrix())

		vecToPlayer = player.Body().Transform.Position().Sub(enemyPos)
		distToPlayer = vecToPlayer.Len()
		if distToPlayer != 0.0 {
			dirToPlayer = vecToPlayer.Normalize()
		}

		if distToPlayer < ENEMY_WAKE_PROXIMITY {
			canSeePlayer = true
		} else if angle := math2.Acos(dirToPlayer.Dot(enemyDir)); angle < ENEMY_FOV_RADS/2.0 {
			res, handle := enemy.world.Raycast(enemyPos, dirToPlayer, COL_LAYER_PLAYERS|COL_LAYER_MAP, 25.0, enemy)
			if handle.Equals(player.id.Handle) && res.Hit {
				canSeePlayer = true
			}
		}
	}

	if !canSeePlayer {
		enemy.wakeTimer = max(0.0, enemy.wakeTimer-deltaTime)
	} else {
		enemy.wakeTimer = min(enemy.WakeLimit, enemy.wakeTimer+deltaTime)
	}

	if enemy.actor.Health <= 0.0 {
		enemy.changeState(ENEMY_STATE_DIE)
	}

	enemy.spriteAngle = enemy.actor.YawAngle

	switch enemy.state {
	case ENEMY_STATE_IDLE:
		if enemy.wakeTimer >= enemy.WakeTime {
			enemy.changeState(ENEMY_STATE_CHASE)
		}
		enemy.actor.inputForward = 0.0
		enemy.actor.inputStrafe = 0.0
	case ENEMY_STATE_CHASE:
		enemy.wraithChase(deltaTime, 3.0, 1.0, vecToPlayer)
		if enemy.wakeTimer <= 0.0 && !canSeePlayer {
			enemy.changeState(ENEMY_STATE_IDLE)
		}
		if distToPlayer < WRAITH_MELEE_RANGE {
			enemy.changeState(ENEMY_STATE_ATTACK)
		}
	case ENEMY_STATE_STUN:
		enemy.actor.inputForward, enemy.actor.inputStrafe = 0.0, 0.0
		if enemy.stateTimer > enemy.StunTime {
			enemy.wakeTimer = enemy.WakeLimit
			enemy.changeState(ENEMY_STATE_CHASE)
		}
	case ENEMY_STATE_ATTACK:
		enemy.actor.inputForward, enemy.actor.inputStrafe = 0.0, 0.0
		if enemy.stateTimer > 0.5 {
			if distToPlayer >= WRAITH_MELEE_RANGE {
				enemy.changeState(ENEMY_STATE_CHASE)
			} else {
				// TODO: Send damage
			}
			enemy.stateTimer = 0.0
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
	enemy.stateTimer += deltaTime
}

func (enemy *Enemy) Render(context *render.Context) {
	enemy.SpriteRender.Render(&enemy.Body().Transform, &enemy.AnimPlayer, context, enemy.spriteAngle)
	enemy.bloodParticles.Render(&enemy.Body().Transform, context)
}

func (enemy *Enemy) ProcessSignal(signal Signal, params any) {
}

func (enemy *Enemy) changeState(newState EnemyState) {
	if newState == enemy.state {
		return
	}
	// Initialize new state
	switch newState {
	case ENEMY_STATE_IDLE:
		enemy.AnimPlayer.ChangeAnimation(enemy.walkAnim)
		enemy.AnimPlayer.Stop()
	case ENEMY_STATE_CHASE:
		if enemy.state == ENEMY_STATE_IDLE {
			enemy.voice = cache.GetSfx(SFX_WRAITH_WAKE).Play()
		}
		enemy.AnimPlayer.ChangeAnimation(enemy.walkAnim)
		enemy.AnimPlayer.Play()
	case ENEMY_STATE_STUN:
		enemy.voice = cache.GetSfx(SFX_WRAITH_HURT).Play()
		enemy.AnimPlayer.ChangeAnimation(enemy.stunAnim)
		enemy.AnimPlayer.PlayFromStart()
	case ENEMY_STATE_ATTACK:
		enemy.AnimPlayer.ChangeAnimation(enemy.attackAnim)
		enemy.AnimPlayer.PlayFromStart()
	case ENEMY_STATE_DIE:
		enemy.voice = cache.GetSfx(SFX_WRAITH_DIE).Play()
		enemy.AnimPlayer.ChangeAnimation(enemy.dieAnim)
		enemy.AnimPlayer.PlayFromStart()
		enemy.actor.body.Layer = COL_LAYER_NONE
		enemy.actor.body.Filter = COL_LAYER_NONE
		enemy.bloodParticles.EmissionTimer = enemy.dieAnim.Duration()
	}
	enemy.stateTimer = 0.0
	enemy.state = newState
}

func (enemy *Enemy) OnDamage(sourceEntity any, damage float32) {
	if enemy.state == ENEMY_STATE_DIE {
		return
	}
	if proj, ok := sourceEntity.(*Projectile); ok && proj.Body().OnLayer(COL_LAYER_PROJECTILES) {
		enemy.bloodParticles.EmissionTimer = 0.1
		enemy.actor.Health -= damage
		if enemy.state != ENEMY_STATE_STUN {
			if rand.Float32() < enemy.StunChance*proj.StunChance {
				enemy.changeState(ENEMY_STATE_STUN)
			} else {
				enemy.wakeTimer = enemy.WakeLimit
				enemy.changeState(ENEMY_STATE_CHASE)
			}
		}
	}
}

func (enemy *Enemy) wraithChase(
	deltaTime float32,
	chaseStraightTime float32, // Number of seconds enemy chases in a straight line before turning.
	chaseStrafeTime float32, // Number of seconds enemy chases diagonally.
	vecToPlayer mgl32.Vec3,
) {
	totalChaseTime := chaseStraightTime + chaseStrafeTime
	enemy.actor.inputForward = 1.0
	enemy.chaseTimer += deltaTime
	enemy.actor.YawAngle = math2.Atan2(-vecToPlayer.X(), -vecToPlayer.Z())
	if enemy.chaseTimer < chaseStraightTime {
		enemy.chaseStrafeDir = 0.0
		enemy.spriteAngle = enemy.actor.YawAngle
	} else if enemy.chaseTimer < totalChaseTime {
		if enemy.chaseStrafeDir == 0.0 {
			enemy.chaseStrafeDir = ([2]float32{-0.7, 0.7})[rand.Intn(2)]
		}
		enemy.spriteAngle = enemy.actor.YawAngle - (math2.Signum(enemy.chaseStrafeDir) * math.Pi / 2.0)
	} else {
		for enemy.chaseTimer > totalChaseTime {
			enemy.chaseTimer -= totalChaseTime
		}
	}
	enemy.actor.inputStrafe = enemy.chaseStrafeDir
}
