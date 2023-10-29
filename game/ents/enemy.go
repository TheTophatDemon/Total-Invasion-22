package ents

import (
	"github.com/go-gl/mathgl/mgl32"
	"tophatdemon.com/total-invasion-ii/engine/assets/cache"
	"tophatdemon.com/total-invasion-ii/engine/assets/textures"
	"tophatdemon.com/total-invasion-ii/engine/math2"
	"tophatdemon.com/total-invasion-ii/engine/math2/collision"
	"tophatdemon.com/total-invasion-ii/engine/render"
	"tophatdemon.com/total-invasion-ii/engine/world/comps"
)

type Enemy struct {
	Actor
	SpriteRender comps.SpriteRender
	AnimPlayer   comps.AnimationPlayer
	timer        float32
}

func NewEnemy(position, angles mgl32.Vec3) Enemy {
	wraithTexture := cache.GetTexture("assets/textures/sprites/wraith.png")
	var anim textures.Animation
	if int(position.Len()*1000.0)%2 == 0 {
		anim, _ = wraithTexture.GetAnimation("attack;front")
	} else {
		anim, _ = wraithTexture.GetAnimation("walk;front")
	}
	return Enemy{
		Actor: Actor{
			Body: comps.Body{
				Transform: comps.TransformFromTranslationAngles(
					position, angles,
				),
				Shape:     collision.ShapeSphere(0.7),
				Pushiness: 10,
				NoClip:    false,
			},
			YawAngle:  mgl32.DegToRad(angles[1]),
			AccelRate: 80.0,
			Friction:  20.0,
			MaxSpeed:  5.0,
			RestrictY: position.Y(),
		},
		SpriteRender: comps.NewSpriteRender(wraithTexture),
		AnimPlayer:   comps.NewAnimationPlayer(anim, true),
	}
}

func (enemy *Enemy) Update(deltaTime float32) {
	enemy.timer += deltaTime
	enemy.inputForward = math2.Sin(enemy.timer)
	// enemy.YawAngle += deltaTime * 2.0
	// enemy.Body.Transform.SetRotation(0.0, enemy.YawAngle, 0.0)
	enemy.AnimPlayer.Update(deltaTime)
	enemy.Actor.Update(deltaTime)
}

func (enemy *Enemy) Render(context *render.Context) {
	enemy.SpriteRender.Render(&enemy.Body.Transform, &enemy.AnimPlayer, context, enemy.YawAngle)
}
