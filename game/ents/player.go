package ents

import (
	"github.com/go-gl/mathgl/mgl32"
	"tophatdemon.com/total-invasion-ii/engine/input"
	"tophatdemon.com/total-invasion-ii/engine/math2"
	"tophatdemon.com/total-invasion-ii/engine/world/comps"

	"tophatdemon.com/total-invasion-ii/game/settings"
)

type Player struct {
	Actor
	Camera                                   comps.Camera
	RunSpeed, WalkSpeed                      float32
	StandFriction, WalkFriction, RunFriction float32
}

func NewPlayer(position, angles mgl32.Vec3) Player {
	return Player{
		Actor: Actor{
			Transform: comps.TransformFromTranslationAngles(
				position, angles,
			),
			Body: comps.Body{
				Shape:      comps.COL_SHAPE_SPHERE,
				Extents:    mgl32.Vec3{0.7, 0.7, 0.7},
				Unpushable: false,
				NoClip:     false,
			},
			YawAngle:  mgl32.DegToRad(angles[1]),
			AccelRate: 80.0,
			Friction:  20.0,
		},
		Camera: comps.NewCamera(
			70.0, settings.WINDOW_ASPECT_RATIO, 0.1, 1000.0,
		),
		RunSpeed:      12.0,
		WalkSpeed:     7.0,
		StandFriction: 80.0,
		WalkFriction:  1.0,
		RunFriction:   20.0,
	}
}

func (player *Player) Update(deltaTime float32) {
	if input.IsActionPressed(settings.ACTION_FORWARD) {
		player.inputForward = 1.0
	} else if input.IsActionPressed(settings.ACTION_BACK) {
		player.inputForward = -1.0
	} else {
		player.inputForward = 0.0
	}

	if input.IsActionPressed(settings.ACTION_RIGHT) {
		player.inputStrafe = 1.0
	} else if input.IsActionPressed(settings.ACTION_LEFT) {
		player.inputStrafe = -1.0
	} else {
		player.inputStrafe = 0.0
	}

	if input.IsActionPressed(settings.ACTION_SLOW) {
		player.MaxSpeed = player.WalkSpeed
	} else {
		player.MaxSpeed = player.RunSpeed
	}

	if math2.Abs(player.inputForward) > mgl32.Epsilon || math2.Abs(player.inputStrafe) > mgl32.Epsilon {
		if player.MaxSpeed == player.WalkSpeed {
			player.Friction = player.WalkFriction
		} else {
			player.Friction = player.RunFriction
		}
	} else {
		player.Friction = player.StandFriction
	}

	player.YawAngle -= input.ActionAxis(settings.ACTION_LOOK_HORZ)
	player.Transform.SetRotation(0.0, player.YawAngle, 0.0)

	player.Actor.Update(deltaTime)
}
