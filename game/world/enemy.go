package world

import (
	"github.com/go-gl/mathgl/mgl32"
	"tophatdemon.com/total-invasion-ii/engine/assets/cache"
	"tophatdemon.com/total-invasion-ii/engine/assets/textures"
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
	enemy.bloodParticles = comps.NewParticleRender(cache.QuadMesh, bloodTexture, nil, 25)

	return
}

func (enemy *Enemy) Update(deltaTime float32) {
	enemy.timer += deltaTime
	enemy.actor.inputForward = math2.Sin(enemy.timer)
	enemy.AnimPlayer.Update(deltaTime)
	enemy.actor.Update(deltaTime)
	enemy.bloodParticles.Update(deltaTime)
}

func (enemy *Enemy) Render(context *render.Context) {
	enemy.SpriteRender.Render(&enemy.Body().Transform, &enemy.AnimPlayer, context, enemy.actor.YawAngle)
	enemy.bloodParticles.Render(&enemy.Body().Transform, context)
}

func (enemy *Enemy) ProcessSignal(signal Signal, params any) {
}
