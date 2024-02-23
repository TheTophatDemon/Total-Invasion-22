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
	SpriteRender comps.SpriteRender
	AnimPlayer   comps.AnimationPlayer
	actor        Actor
	timer        float32
}

var _ HasActor = (*Enemy)(nil)
var _ comps.HasBody = (*Enemy)(nil)

func (e *Enemy) Actor() *Actor {
	return &e.actor
}

func (e *Enemy) Body() *comps.Body {
	return &e.actor.body
}

func SpawnEnemy(st *scene.Storage[Enemy], position, angles mgl32.Vec3) (id scene.Id[Enemy], wr *Enemy, err error) {
	id, wr, err = st.New()
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

	wr.actor = Actor{
		body: comps.Body{
			Transform: comps.TransformFromTranslationAnglesScale(
				mgl32.Vec3(position).Add(mgl32.Vec3{0.0, -0.1, 0.0}), angles, mgl32.Vec3{0.9, 0.9, 0.9},
			),
			Shape:     collision.NewSphere(0.7),
			Pushiness: 10,
			NoClip:    false,
		},
		YawAngle:  mgl32.DegToRad(angles[1]),
		AccelRate: 80.0,
		Friction:  20.0,
		MaxSpeed:  5.0,
	}
	wr.SpriteRender = comps.NewSpriteRender(wraithTexture)
	wr.AnimPlayer = comps.NewAnimationPlayer(anim, true)

	return
}

func (enemy *Enemy) Update(deltaTime float32) {
	enemy.timer += deltaTime
	enemy.actor.inputForward = math2.Sin(enemy.timer)
	// enemy.YawAngle += deltaTime * 2.0
	// enemy.Body.Transform.SetRotation(0.0, enemy.YawAngle, 0.0)
	enemy.AnimPlayer.Update(deltaTime)
	enemy.actor.Update(deltaTime)
}

func (enemy *Enemy) Render(context *render.Context) {
	enemy.SpriteRender.Render(&enemy.Body().Transform, &enemy.AnimPlayer, context, enemy.actor.YawAngle)
}

func (enemy *Enemy) ProcessSignal(s Signal, params any) {
}
