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
)

type EnemyState uint8

const (
	ENEMY_STATE_IDLE EnemyState = iota
	ENEMY_STATE_CHASE
)

type Enemy struct {
	SpriteRender          comps.SpriteRender
	AnimPlayer            comps.AnimationPlayer
	WakeTime              float32 // Number of seconds player must be in sight before enemy begins to pursue.
	WakeLimit             float32 // Maximum number of seconds after losing sight of player before giving up.
	ChaseStraightTime     float32 // Number of seconds enemy chases in a straight line before turning.
	ChaseStrafeTime       float32 // Number of seconds enemy chases diagonally.
	bloodParticles        comps.ParticleRender
	actor                 Actor
	world                 *World
	walkAnim, attackAnim  textures.Animation
	wakeTimer, chaseTimer float32
	chaseStrafeDir        float32 // 1.0 to strafe right, -1.0 to strafe left while chasing player.
	spriteAngle           float32 // Yaw angle on the Y axis determining where the sprite faces. Sometimes corresponds with actor.YawAngle
	wakeSound             *audio.Sfx
	state                 EnemyState
	voice                 audio.VoiceId
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
	enemy.attackAnim, _ = wraithTexture.GetAnimation("walk;front")
	enemy.wakeSound, _ = cache.GetSfx("assets/sounds/enemy/wraith/wraith_greeting.wav")

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
		MaxSpeed:  5.0,
	}
	enemy.SpriteRender = comps.NewSpriteRender(wraithTexture)
	enemy.AnimPlayer = comps.NewAnimationPlayer(enemy.walkAnim, false)
	enemy.WakeTime = 0.5
	enemy.WakeLimit = 5.0
	enemy.ChaseStraightTime = 3.0
	enemy.ChaseStrafeTime = 1.0
	enemy.chaseTimer = rand.Float32() * enemy.ChaseStrafeTime

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
	enemy.bloodParticles.Update(deltaTime, &enemy.Body().Transform)

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

	enemy.spriteAngle = enemy.actor.YawAngle

	switch enemy.state {
	case ENEMY_STATE_IDLE:
		if enemy.wakeTimer >= enemy.WakeTime {
			enemy.changeState(ENEMY_STATE_CHASE)
		}
		enemy.actor.inputForward = 0.0
		enemy.actor.inputStrafe = 0.0
	case ENEMY_STATE_CHASE:
		if enemy.wakeTimer <= 0.0 && !canSeePlayer {
			enemy.changeState(ENEMY_STATE_IDLE)
		}
		enemy.actor.inputForward = 1.0
		enemy.chaseTimer += deltaTime
		enemy.actor.YawAngle = math2.Atan2(-vecToPlayer.X(), -vecToPlayer.Z())
		if enemy.chaseTimer < enemy.ChaseStraightTime {
			enemy.chaseStrafeDir = 0.0
			enemy.spriteAngle = enemy.actor.YawAngle
		} else if enemy.chaseTimer < enemy.ChaseStraightTime+enemy.ChaseStrafeTime {
			if enemy.chaseStrafeDir == 0.0 {
				enemy.chaseStrafeDir = ([2]float32{-0.7, 0.7})[rand.Intn(2)]
			}
			enemy.spriteAngle = enemy.actor.YawAngle - (math2.Signum(enemy.chaseStrafeDir) * math.Pi / 2.0)
		} else {
			enemy.chaseTimer = 0.0
		}
		enemy.actor.inputStrafe = enemy.chaseStrafeDir
	}
}

func (enemy *Enemy) Render(context *render.Context) {
	enemy.SpriteRender.Render(&enemy.Body().Transform, &enemy.AnimPlayer, context, enemy.spriteAngle)
	enemy.bloodParticles.Render(&enemy.Body().Transform, context)
}

func (enemy *Enemy) ProcessSignal(signal Signal, params any) {
}

func (enemy *Enemy) changeState(newState EnemyState) {
	// Initialize new state
	switch newState {
	case ENEMY_STATE_IDLE:
		enemy.AnimPlayer.ChangeAnimation(enemy.walkAnim)
		enemy.AnimPlayer.Stop()
	case ENEMY_STATE_CHASE:
		enemy.voice = enemy.wakeSound.Play()
		enemy.AnimPlayer.ChangeAnimation(enemy.walkAnim)
		enemy.AnimPlayer.Play()
	}
	enemy.state = newState
}
