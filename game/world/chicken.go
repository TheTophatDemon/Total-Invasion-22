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
	"tophatdemon.com/total-invasion-ii/engine/scene"
	"tophatdemon.com/total-invasion-ii/engine/scene/comps"
	"tophatdemon.com/total-invasion-ii/engine/tdaudio"
)

const (
	SfxChickenBok  = "assets/sounds/chicken/chicken_bok.wav"
	SfxChickenPain = "assets/sounds/chicken/chicken_pain.wav"
)

type Chicken struct {
	SpriteRender               comps.SpriteRender
	AnimPlayer                 comps.AnimationPlayer
	bloodParticles             comps.ParticleRender
	actor                      Actor
	voice                      tdaudio.VoiceId
	walkAnim, flyAnim, dieAnim textures.Animation
	id                         scene.Id[*Chicken]
	decomposeTimer             float32 // Time in seconds before the chicken's corpse disappears.
}

var _ HasActor = (*Chicken)(nil)

func (chk *Chicken) Actor() *Actor {
	return &chk.actor
}

func (chk *Chicken) Body() *comps.Body {
	return &chk.actor.body
}

func SpawnChicken(world *World, position, angles mgl32.Vec3) (id scene.Id[*Chicken], chk *Chicken, err error) {
	id, chk, err = world.Chickens.New()
	if err != nil {
		return
	}

	chk.id = id

	chk.bloodParticles = BloodParticles(5, color.Red, 0.3)
	chk.bloodParticles.Init()

	tex := cache.GetTexture("assets/textures/sprites/chicken.png")
	chk.walkAnim, _ = tex.GetAnimation("walk;front")
	chk.flyAnim, _ = tex.GetAnimation("fly;front")
	chk.dieAnim, _ = tex.GetAnimation("die;front")

	chk.actor = Actor{
		body: comps.Body{
			Position: position,
			Shape:    collision.NewBoxShape(0.5, 0.5, 0.5),
			Layer:    ColLayerActors | ColLayerNPCs,
		},
		collisionFilter: ColLayerMap | ColLayerActors,
		YawAngle:        mgl32.DegToRad(angles[1]),
		AccelRate:       80.0,
		Friction:        20.0,
		MaxSpeed:        2.5,
		GravityAccel:    80.0,
		MaxFallSpeed:    15.0,
	}
	chk.SpriteRender = comps.NewSpriteRender(tex, nil, &mgl32.Vec2{0.5, 0.5})

	chk.AnimPlayer = comps.NewAnimationPlayer(chk.walkAnim, false)

	chk.actor.Health = 45.0
	chk.actor.MaxHealth, chk.actor.TargetHealth = chk.actor.Health, chk.actor.Health

	chk.voice = cache.GetSfx(SfxChickenBok).PlayAttenuatedV(position)

	chk.decomposeTimer = 120.0

	return
}

func (chk *Chicken) Finalize() {
	chk.bloodParticles.Finalize()
}

func (chk *Chicken) Update(deltaTime float32) {
	if chk == nil {
		return
	}
	body := chk.Body()
	chk.AnimPlayer.Update(deltaTime)
	chk.actor.Update(deltaTime)
	chk.bloodParticles.Update(deltaTime, body.Position)

	chkPos := body.Position
	chkDir := chk.actor.FacingVec()
	if chk.voice.IsValid() {
		chk.voice.SetPositionV(chkPos)
	}

	if chk.actor.Health > 0 {
		chk.actor.inputForward = 1.0
		chk.actor.inputStrafe = 0.0
		if chk.actor.onGround {
			chk.AnimPlayer.SwapAnimation(chk.walkAnim)
		} else {
			chk.AnimPlayer.SwapAnimation(chk.flyAnim)
		}
		chk.AnimPlayer.Play()

		// Cast forward to see if there is a wall in front
		hit, closestBody := gWorld.Raycast(chkPos, chkDir, chk.actor.collisionFilter, 1.0, &chk.actor.body)
		if hit.Hit && !closestBody.IsNil() {
			// Turn around if we're about to hit a wall
			chk.actor.YawAngle += math.Pi/2.0 + rand.Float32()*math.Pi/2.0
		}
	} else {
		chk.decomposeTimer -= deltaTime
		if chk.decomposeTimer <= 0.0 {
			gWorld.QueueRemoval(chk.id.Handle)
		}

		chk.actor.inputForward = 0.0
		chk.actor.inputStrafe = 0.0

		if !chk.AnimPlayer.IsPlayingAnim(chk.dieAnim) {
			chk.AnimPlayer.ChangeAnimation(chk.dieAnim)
			chk.AnimPlayer.PlayFromStart()
		} else if chk.AnimPlayer.IsAtEnd() {
			chk.bloodParticles.EmissionTimer = 0.0
		} else {
			chk.bloodParticles.EmissionTimer = 0.5
		}

		if math2.Vec3WithY(body.Velocity, 0.0).ApproxEqual(mgl32.Vec3{}) {
			body.Velocity = mgl32.Vec3{}
			chk.actor.GravityAccel = 0.0
			body.Noclip = true
		}
	}
}

func (chk *Chicken) Render(context *render.Context) {
	chk.SpriteRender.Render(chk.Body().Position, &chk.AnimPlayer, context, chk.actor.YawAngle)
	chk.bloodParticles.Render(chk.Body().Position, context)
}

func (chk *Chicken) ProcessSignal(signal any) {
}

func (chk *Chicken) OnDamage(sourceEntity any, damage float32) bool {
	if chk.actor.Health <= 0 {
		return false
	}
	chk.bloodParticles.EmissionTimer = 0.1
	chk.actor.Health -= damage
	if chk.actor.Health <= 0 {
		chk.voice.Stop()
		chk.voice = cache.GetSfx(SfxChickenPain).PlayAttenuatedV(chk.Body().Position)
		// Spawn an item sometimes
		switch v := rand.Float32(); {
		case v < 0.1:
			SpawnStimpack(chk.actor.Position())
		case v < 0.3:
			SpawnEggCarton(chk.actor.Position())
		}
	} else if !chk.voice.IsPlaying() && rand.Float32() < 0.25 {
		chk.voice = cache.GetSfx(SfxChickenBok).PlayAttenuatedV(chk.Body().Position)
	}
	return true
}
