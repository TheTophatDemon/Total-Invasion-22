package world

import (
	"math"
	"math/rand"

	"github.com/go-gl/mathgl/mgl32"
	"tophatdemon.com/total-invasion-ii/engine/assets/cache"
	"tophatdemon.com/total-invasion-ii/engine/assets/textures"
	"tophatdemon.com/total-invasion-ii/engine/color"
	"tophatdemon.com/total-invasion-ii/engine/math2/collision"
	"tophatdemon.com/total-invasion-ii/engine/render"
	"tophatdemon.com/total-invasion-ii/engine/scene"
	"tophatdemon.com/total-invasion-ii/engine/scene/comps"
	"tophatdemon.com/total-invasion-ii/engine/tdaudio"
	"tophatdemon.com/total-invasion-ii/game/world/effects"
)

const (
	SFX_CHICKEN_BOK  = "assets/sounds/chicken/chicken_bok.wav"
	SFX_CHICKEN_PAIN = "assets/sounds/chicken/chicken_pain.wav"
)

type Chicken struct {
	SpriteRender               comps.SpriteRender
	AnimPlayer                 comps.AnimationPlayer
	bloodParticles             comps.ParticleRender
	actor                      Actor
	voice                      tdaudio.VoiceId
	walkAnim, flyAnim, dieAnim textures.Animation
	world                      *World
}

var _ HasActor = (*Chicken)(nil)

func (chk *Chicken) Actor() *Actor {
	return &chk.actor
}

func (chk *Chicken) Body() *comps.Body {
	return &chk.actor.body
}

func SpawnChicken(storage *scene.Storage[Chicken], position, angles mgl32.Vec3, world *World) (id scene.Id[*Chicken], chk *Chicken, err error) {
	id, chk, err = storage.New()
	if err != nil {
		return
	}

	chk.world = world

	chk.bloodParticles = effects.Blood(5, color.Red, 0.3)
	chk.bloodParticles.Init()

	tex := cache.GetTexture("assets/textures/sprites/chicken.png")
	chk.walkAnim, _ = tex.GetAnimation("walk;front")
	chk.flyAnim, _ = tex.GetAnimation("fly;front")
	chk.dieAnim, _ = tex.GetAnimation("die;front")

	chk.actor = Actor{
		body: comps.Body{
			Transform: comps.TransformFromTranslationAnglesScale(
				mgl32.Vec3(position), mgl32.Vec3{}, mgl32.Vec3{0.5, 0.5, 0.5},
			),
			Shape:  collision.NewSphere(0.5),
			Layer:  COL_LAYER_ACTORS | COL_LAYER_NPCS,
			Filter: COL_FILTER_FOR_ACTORS,
			LockY:  false,
		},
		YawAngle:     mgl32.DegToRad(angles[1]),
		AccelRate:    80.0,
		Friction:     20.0,
		MaxSpeed:     2.5,
		GravityAccel: 80.0,
		MaxFallSpeed: 15.0,
		world:        world,
	}
	chk.SpriteRender = comps.NewSpriteRender(tex)
	chk.AnimPlayer = comps.NewAnimationPlayer(chk.walkAnim, false)

	chk.actor.Health = 30.0
	chk.actor.MaxHealth = chk.actor.Health

	chk.voice = cache.GetSfx(SFX_CHICKEN_BOK).PlayAttenuatedV(position)

	return
}

func (chk *Chicken) Finalize() {
	chk.bloodParticles.Finalize()
}

func (chk *Chicken) Update(deltaTime float32) {
	chk.AnimPlayer.Update(deltaTime)
	chk.actor.Update(deltaTime)
	chk.bloodParticles.Update(deltaTime, &chk.Body().Transform)

	chkPos := chk.Body().Transform.Position()
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
		hit, closestBody := chk.world.Raycast(chkPos, chkDir, COL_FILTER_FOR_ACTORS, 1.0, chk)
		if hit.Hit && !closestBody.IsNil() {
			// Turn around if we're about to hit a wall
			chk.actor.YawAngle += math.Pi/2.0 + rand.Float32()*math.Pi/2.0
		}
	} else {
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

		if chk.actor.body.Velocity.ApproxEqual(mgl32.Vec3{}) {
			chk.actor.body.Layer = COL_LAYER_NONE
			chk.actor.body.Filter = COL_LAYER_NONE
		}
	}
}

func (chk *Chicken) Render(context *render.Context) {
	chk.SpriteRender.Render(&chk.Body().Transform, &chk.AnimPlayer, context, chk.actor.YawAngle)
	chk.bloodParticles.Render(&chk.Body().Transform, context)
}

func (chk *Chicken) ProcessSignal(signal any) {
}

func (chk *Chicken) OnDamage(sourceEntity any, damage float32) bool {
	if chk.actor.Health <= 0 {
		return false
	}
	if _, ok := sourceEntity.(*Projectile); ok {
		chk.bloodParticles.EmissionTimer = 0.1
		chk.actor.Health -= damage
	}
	if chk.actor.Health <= 0 {
		chk.voice = cache.GetSfx(SFX_CHICKEN_PAIN).PlayAttenuatedV(chk.Body().Transform.Position())
	} else if !chk.voice.IsPlaying() && rand.Float32() < 0.25 {
		chk.voice = cache.GetSfx(SFX_CHICKEN_BOK).PlayAttenuatedV(chk.Body().Transform.Position())
	}
	return true
}
