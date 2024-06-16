package world

import (
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
)

type Enemy struct {
	SpriteRender   comps.SpriteRender
	AnimPlayer     comps.AnimationPlayer
	bloodParticles comps.ParticleRender
	actor          Actor
	timer          float32
}

var _ HasActor = (*Enemy)(nil)
var _ comps.HasBody = (*Enemy)(nil)

func (enemy *Enemy) Actor() *Actor {
	return &enemy.actor
}

func (enemy *Enemy) Body() *comps.Body {
	return &enemy.actor.body
}

func SpawnEnemy(storage *scene.Storage[Enemy], position, angles mgl32.Vec3) (id scene.Id[*Enemy], enemy *Enemy, err error) {
	id, enemy, err = storage.New()
	if err != nil {
		return
	}

	wraithTexture := cache.GetTexture("assets/textures/sprites/wraith.png")
	var anim textures.Animation
	if int(position.Len()*1000.0)%2 == 0 {
		anim, _ = wraithTexture.GetAnimation("attack;front")
	} else {
		anim, _ = wraithTexture.GetAnimation("walk;front")
	}

	enemy.actor = Actor{
		body: comps.Body{
			Transform: comps.TransformFromTranslationAnglesScale(
				mgl32.Vec3(position).Add(mgl32.Vec3{0.0, -0.1, 0.0}), angles, mgl32.Vec3{0.9, 0.9, 0.9},
			),
			Shape:  collision.NewSphere(0.7),
			Layer:  COL_LAYER_ACTORS,
			Filter: COL_FILTER_FOR_ACTORS,
			LockY:  true,
		},
		YawAngle:  mgl32.DegToRad(angles[1]),
		AccelRate: 80.0,
		Friction:  20.0,
		MaxSpeed:  5.0,
	}
	enemy.SpriteRender = comps.NewSpriteRender(wraithTexture)
	enemy.AnimPlayer = comps.NewAnimationPlayer(anim, true)

	bloodTexture := cache.GetTexture("assets/textures/sprites/blood.png")
	bloodAnim, _ := bloodTexture.GetAnimation("default")
	enemy.bloodParticles = comps.ParticleRender{
		Mesh:             cache.QuadMesh,
		Texture:          bloodTexture,
		SpawnRate:        0.01,
		SpawnRadius:      0.5,
		VisibilityRadius: 5.0,
		EmissionTimer:    math2.Inf32(),
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
	enemy.timer += deltaTime
	enemy.actor.inputForward = math2.Sin(enemy.timer)
	enemy.AnimPlayer.Update(deltaTime)
	enemy.actor.Update(deltaTime)
	enemy.bloodParticles.Update(deltaTime, &enemy.Body().Transform)
}

func (enemy *Enemy) Render(context *render.Context) {
	enemy.SpriteRender.Render(&enemy.Body().Transform, &enemy.AnimPlayer, context, enemy.actor.YawAngle)
	enemy.bloodParticles.Render(&enemy.Body().Transform, context)
}

func (enemy *Enemy) ProcessSignal(signal Signal, params any) {
}
