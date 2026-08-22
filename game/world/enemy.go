package world

import (
	"fmt"
	"math"
	"math/rand"

	"github.com/go-gl/mathgl/mgl32"
	"tophatdemon.com/total-invasion-ii/engine/assets/te3"
	"tophatdemon.com/total-invasion-ii/engine/assets/textures"
	"tophatdemon.com/total-invasion-ii/engine/color"
	"tophatdemon.com/total-invasion-ii/engine/containers/maybe"
	"tophatdemon.com/total-invasion-ii/engine/math2"
	"tophatdemon.com/total-invasion-ii/engine/math2/collision"
	"tophatdemon.com/total-invasion-ii/engine/render"
	"tophatdemon.com/total-invasion-ii/engine/scene"
	"tophatdemon.com/total-invasion-ii/engine/scene/comps"
	"tophatdemon.com/total-invasion-ii/engine/tdaudio"
	"tophatdemon.com/total-invasion-ii/game"
	"tophatdemon.com/total-invasion-ii/game/settings"
)

const (
	EnemyColLayers = ColLayerActors | ColLayerNPCs
)

type Enemy struct {
	SpriteRender                                   comps.SpriteRender
	AnimPlayer                                     comps.AnimationPlayer
	WakeTime                                       float32 // Number of seconds player must be in sight before enemy begins to pursue.
	WakeLimit                                      float32 // Maximum number of seconds after losing sight of player before giving up.
	StunTime                                       float32 // Number of seconds the enemy stays in the 'stunned' state after getting hurt.
	StunChance                                     float32 // The probability from 0 to 1 of the enemy getting stunned when hurt.
	bloodParticles                                 comps.ParticleRender
	actor                                          Actor
	id                                             scene.Id[*Enemy]
	wakeTimer, chaseTimer, stateTimer, attackTimer float32
	chaseStrafeDir                                 float32       // 1.0 to strafe right, -1.0 to strafe left while chasing player.
	spriteAngle                                    math2.Radians // Yaw angle on the Y axis determining where the sprite faces. Sometimes corresponds with actor.YawAngle
	idleState, chaseState, stunState               enemyState
	attackState, dodgeState, dieState, reviveState enemyState
	state, previousState                           *enemyState
	voice                                          tdaudio.VoiceId
	spawnAmmo                                      game.AmmoType // Ammo type that will drop when enemy is killed
	spawnAmmoChance                                float32       // Probability from 0 to 1
	defaultCollisionFilters                        collision.Mask
	spawnOffset                                    mgl32.Vec3 // Vector to add to enemy's position after spawning.

	// Player or target tracking variables
	targetHandle                scene.Handle
	dirToTarget                 mgl32.Vec3
	distToTarget                float32
	canSeeTarget, canHearTarget bool

	variant     game.EnemyType
	framesAlive int
}

type enemyState struct {
	anim                   textures.Animation
	stopAnim               bool // Set to true to leave the animation on its first frame without playing it.
	enterSound, leaveSound tdaudio.SoundId
	collisionFilters       maybe.T[collision.Mask]
	updateFunc             func(enemy *Enemy, deltaTime float32)
	enterFunc              func(enemy *Enemy, oldState *enemyState)
	leaveFunc              func(enemy *Enemy, newState *enemyState)
}

type enemyConfig struct {
	bloodColor  color.Color
	texture     *textures.Texture
	defaultAnim textures.Animation
	spawnOffset maybe.T[mgl32.Vec3]
}

var enemyTypeConfigFuncs = [game.EnemyTypeCount]func(enemy *Enemy) enemyConfig{
	game.EnemyTypeWraith:       configureWraith,
	game.EnemyTypeFireWraith:   configureFireWraith,
	game.EnemyTypeMotherWraith: configureMotherWraith,
	game.EnemyTypeDummkopf:     configureDummkopf,
	game.EnemyTypePrisrak:      configurePrisrak,
}

var _ HasActor = (*Enemy)(nil)
var _ comps.HasBody = (*Enemy)(nil)

func SpawnEnemyFromTE3(ent te3.Ent) (scene.Id[*Enemy], *Enemy, error) {
	var variant game.EnemyType
	switch ent.Properties["enemy"] {
	case "fire wraith":
		variant = game.EnemyTypeFireWraith
	case "mother wraith":
		variant = game.EnemyTypeMotherWraith
	case "dummkopf":
		variant = game.EnemyTypeDummkopf
	case "prisrak":
		variant = game.EnemyTypePrisrak
	default:
		variant = game.EnemyTypeWraith
	}
	id, enemy, err := SpawnEnemy(ent.Position, ent.AnglesInRadians(), variant)
	if err != nil {
		return id, enemy, err
	}

	enemy.actor.Health = ent.FloatPropertyOr("health", enemy.actor.TargetHealth)
	if enemy.actor.Health <= 0 {
		enemy.changeState(&enemy.dieState)
	}

	return id, enemy, err
}

func (enemy *Enemy) Actor() *Actor {
	return &enemy.actor
}

func (enemy *Enemy) Body() *comps.Body {
	return &enemy.actor.body
}

func SpawnEnemy(position mgl32.Vec3, angles [3]math2.Radians, variant game.EnemyType) (id scene.Id[*Enemy], enemy *Enemy, err error) {
	id, enemy, err = gWorld.Enemies.New()
	if err != nil {
		return
	}

	gWorld.Hud.VictoryScreen.EnemiesTotal++
	enemy.variant = variant
	enemy.id = id

	enemy.actor = Actor{
		body: comps.Body{
			Position: mgl32.Vec3(position),
			Shape:    collision.NewBoxShape(0.6, 0.7, 0.6),
			Layers:   EnemyColLayers,
		},
		YawAngle:  angles[1],
		AccelRate: 80.0,
		Friction:  20.0,
		MaxSpeed:  5.5,
	}
	enemy.WakeTime = 0.5
	enemy.WakeLimit = 5.0
	enemy.StunChance = 1.0
	enemy.StunTime = 0.5
	enemy.chaseTimer = rand.Float32() * 10.0
	enemy.defaultCollisionFilters = ColLayerMap | ColLayerActors | ColLayerInvisible

	params := enemyTypeConfigFuncs[variant](enemy)

	if !enemy.dieState.collisionFilters.IsSome() {
		enemy.dieState.collisionFilters = maybe.Some(ColLayerMap | ColLayerInvisible)
	}

	enemy.bloodParticles = BloodParticles(15, params.bloodColor, 0.5)
	enemy.bloodParticles.Init()
	enemy.actor.MaxHealth *= settings.CurrDifficulty().EnemyHealthMultiplier
	enemy.actor.Health, enemy.actor.TargetHealth = enemy.actor.MaxHealth, enemy.actor.MaxHealth

	enemy.SpriteRender = comps.NewSpriteRender(params.texture, nil, &mgl32.Vec2{0.9, 0.9})
	enemy.AnimPlayer = comps.NewAnimationPlayer(params.defaultAnim, false)

	enemy.spawnOffset = params.spawnOffset.Or(mgl32.Vec3{0.0, -0.1, 0.0})
	enemy.Body().TranslateV(enemy.spawnOffset)

	enemy.changeState(&enemy.idleState)

	return
}

func (enemy *Enemy) Finalize() {
	enemy.bloodParticles.Finalize()
}

func (enemy *Enemy) Update(deltaTime float32) {
	if enemy == nil {
		return
	}

	if settings.ActionKillEnemies.JustPressed() {
		// Kill all cheat
		enemy.actor.Health = 0
		gWorld.Hud.ShowMessage(settings.Localize("killAllEnemies"), 100, color.Red)
	}

	enemy.AnimPlayer.Update(deltaTime)
	enemy.actor.Update(deltaTime)

	enemy.bloodParticles.Update(deltaTime, enemy.Body().Position)

	enemyPos := enemy.Body().Position
	enemyDir := enemy.actor.FacingVec()
	if enemy.voice.IsValid() {
		enemy.voice.SetPositionV(enemyPos)
	}

	// Check if the player is in view and not obstructed
	enemy.canSeeTarget = false
	enemy.canHearTarget = false
	var vecToTarget mgl32.Vec3
	if enemy.targetHandle.IsNil() {
		enemy.targetHandle = gWorld.CurrentPlayer.Handle
	}
	if targetActor, ok := enemy.targetHandle.Get[HasActor](); ok && gWorld.IsOnPlayerCamera() {
		vecToTarget = targetActor.Body().Position.Sub(enemyPos)
		enemy.distToTarget = vecToTarget.Len()
		if enemy.distToTarget != 0.0 {
			enemy.dirToTarget = vecToTarget.Normalize()
		}

		inHearingRange := enemy.distToTarget < targetActor.Actor().NoiseLevel

		const enemyFovRads = math.Pi
		inFieldOfView := math2.Acos(enemy.dirToTarget.Dot(enemyDir)) < enemyFovRads/2.0
		const wakeProximity = 1.7
		const noticeProximity = 25.0
		if enemy.distToTarget < wakeProximity {
			enemy.canSeeTarget = true
		} else if inHearingRange || inFieldOfView {
			res, _ := gWorld.Raycast(enemyPos, enemy.dirToTarget, ColLayerMap, enemy.distToTarget, nil)
			if !res.Hit && enemy.distToTarget < noticeProximity {
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
	} else {
		enemy.framesAlive++
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
		radius := enemy.Body().Shape.Radius()
		if enemy.bloodParticles.LocalTransform.Position()[1] > -radius {
			enemy.bloodParticles.LocalTransform.Translate(0.0, -deltaTime, 0.0)
		} else {
			enemy.bloodParticles.LocalTransform.SetPosition(0.0, -radius, 0.0)
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
	enemy.SpriteRender.Render(enemy.Body().Position, &enemy.AnimPlayer, context, enemy.spriteAngle, string(settings.Current.Locale))
	enemy.bloodParticles.Render(enemy.Body().Position, context)
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

	if oldState != nil && enemy.framesAlive > 0 {
		if oldState.leaveFunc != nil {
			oldState.leaveFunc(enemy, newState)
		} else if oldState == &enemy.dieState {
			// Ensure nobody's standing on top of the enemy that is getting revived.
			actorsIter := gWorld.IterActorsInSphere(enemy.Body().Position, enemy.Body().Shape.Radius(), enemy)
			for {
				actor, _ := actorsIter.Next()
				if actor == nil {
					break
				}
				if actor.Actor().Health > 0 {
					return
				}
			}

			gWorld.Hud.VictoryScreen.EnemiesKilled--
			enemy.actor.body.RestoreLayers()
			enemy.bloodParticles.LocalTransform.SetPosition(0, 0, 0)
		}
		if leaveSound := oldState.leaveSound; leaveSound.IsValid() {
			enemy.voice.Stop()
			enemy.voice = leaveSound.PlayAttenuatedV(enemy.actor.Position())
		}
	}

	if newState != nil {
		if newState.anim.Frames != nil {
			enemy.AnimPlayer.ChangeAnimation(newState.anim)
			if newState.stopAnim {
				enemy.AnimPlayer.Stop()
			} else {
				enemy.AnimPlayer.PlayFromStart()
			}
			if enemy.framesAlive == 0 {
				enemy.AnimPlayer.MoveToFrame(-1)
			}
		}
		if newState.enterSound.IsValid() && enemy.framesAlive > 0 {
			enemy.voice.Stop()
			enemy.voice = newState.enterSound.PlayAttenuatedV(enemy.actor.Position())
		}

		// Initialize new state
		enemy.actor.collisionFilter = newState.collisionFilters.Or(enemy.defaultCollisionFilters)
		if newState.enterFunc != nil {
			newState.enterFunc(enemy, enemy.state)
		} else if newState == &enemy.dieState {
			enemy.actor.body.ExcludeLayers(collision.MaskAll)

			if enemy.framesAlive > 0 {
				gWorld.Hud.VictoryScreen.EnemiesKilled++
				enemy.bloodParticles.EmissionTimer = newState.anim.Duration()

				if enemy.spawnAmmo != game.AmmoTypeNone && rand.Float32() < enemy.spawnAmmoChance {
					SpawnAmmo(enemy.actor.Position().Add(enemy.actor.FacingVec().Mul(0.5)), enemy.spawnAmmo)
				}
			}
		}
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
	enemy.actor.YawAngle = math2.Radians(math2.Atan2(-enemy.dirToTarget.X(), -enemy.dirToTarget.Z()))
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
		enemy.spriteAngle = enemy.actor.YawAngle - math2.Radians(math2.Signum(enemy.chaseStrafeDir)*math.Pi/2.0)

		// Cancel the turn if we are facing a wall
		hit, _ := gWorld.Raycast(
			enemy.actor.Position(),
			mgl32.Vec3{float32(-math2.Sin(enemy.spriteAngle)), 0.0, float32(-math2.Cos(enemy.spriteAngle))},
			ColLayerMap|ColLayerActors|ColLayerInvisible,
			wraithMeleeRange,
			enemy.Body(),
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
			enemy.actor.YawAngle = math2.Radians(math2.Atan2(-enemy.dirToTarget.X(), -enemy.dirToTarget.Z()))
		case 1:
			enemy.actor.YawAngle = math2.Radians(math2.Atan2(enemy.dirToTarget.X(), enemy.dirToTarget.Z()))
		case 2:
			enemy.actor.YawAngle = math2.Radians(math2.Atan2(-enemy.dirToTarget.Z(), enemy.dirToTarget.X()))
		case 3:
			enemy.actor.YawAngle = math2.Radians(math2.Atan2(enemy.dirToTarget.Z(), -enemy.dirToTarget.X()))
		}
		enemy.chaseTimer = 0.0
	} else {
		// Cancel the movement if we are approaching an obstacle
		hit, _ := gWorld.Raycast(
			enemy.actor.Position(),
			enemy.actor.FacingVec(),
			enemy.actor.collisionFilter,
			wraithMeleeRange,
			enemy.Body(),
		)
		if hit.Hit {
			enemy.chaseTimer = moveTime
		}
	}
}

func (enemy *Enemy) Save() te3.Ent {
	return te3.Ent{
		Angles:   [3]math2.Degrees{0, math2.ToDegrees(enemy.actor.YawAngle), 0.0},
		Position: enemy.actor.Position().Sub(enemy.spawnOffset),
		Texture:  "assets/textures/sprites/wraith.png",
		Radius:   0.7,
		Display:  te3.ENT_DISPLAY_SPRITE,
		Color:    [3]uint8{255, 255, 255},
		Properties: map[string]string{
			"type":   "enemy",
			"enemy":  enemy.variant.String(),
			"health": fmt.Sprintf("%.2f", enemy.actor.Health),
		},
	}
}
