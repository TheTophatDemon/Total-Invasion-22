package ents

import (
	"github.com/go-gl/mathgl/mgl32"
	"tophatdemon.com/total-invasion-ii/engine/assets"
	"tophatdemon.com/total-invasion-ii/engine/assets/shaders"
	"tophatdemon.com/total-invasion-ii/engine/math2/collision"
	"tophatdemon.com/total-invasion-ii/engine/render"
	"tophatdemon.com/total-invasion-ii/engine/world/comps"
)

type Enemy struct {
	Actor
	MeshRender comps.MeshRender
	AnimPlayer comps.AnimationPlayer
}

func NewEnemy(position, angles mgl32.Vec3) Enemy {
	wraithTexture := assets.GetTexture("assets/textures/sprites/wraith.png")
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
			MaxSpeed:  7.0,
			RestrictY: position.Y(),
		},
		MeshRender: comps.NewMeshRender(
			assets.QuadMesh,
			shaders.SpriteShader,
			wraithTexture,
		),
		AnimPlayer: comps.NewAnimationPlayer(
			wraithTexture.GetAnimation(0),
			true,
		),
	}
}

func (enemy *Enemy) Update(deltaTime float32) {
	enemy.inputForward = 1.0
	enemy.YawAngle += deltaTime * 2.0
	enemy.Body.Transform.SetRotation(0.0, enemy.YawAngle, 0.0)
	enemy.AnimPlayer.Update(deltaTime)
	enemy.Actor.Update(deltaTime)
}

func (enemy *Enemy) Render(context *render.Context) {
	enemy.MeshRender.Render(&enemy.Body.Transform, &enemy.AnimPlayer, context)
}
