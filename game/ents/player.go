package ents

import (
	"github.com/go-gl/mathgl/mgl32"
	"tophatdemon.com/total-invasion-ii/engine/input"
	"tophatdemon.com/total-invasion-ii/engine/math2"
	"tophatdemon.com/total-invasion-ii/engine/world/comps"

	"tophatdemon.com/total-invasion-ii/game/settings"
)

type Player struct {
	Transform comps.Transform
	Movement  comps.Movement
	Camera    comps.Camera
	FlyMode   bool
}

func NewPlayer(position, angles mgl32.Vec3) Player {
	return Player{
		Transform: comps.TransformFromTranslationAngles(
			position, angles,
		),
		Camera: comps.NewCamera(
			70.0, settings.WINDOW_ASPECT_RATIO, 0.1, 1000.0,
		),
		Movement: comps.Movement{
			MaxSpeed:   12.0,
			YawAngle:   mgl32.DegToRad(angles[1]),
			PitchAngle: 0.0,
		},
	}
}

func (player *Player) Update(deltaTime float32) {
	player.Movement.Update(&player.Transform, deltaTime)

	if input.IsActionPressed(settings.ACTION_FORWARD) {
		player.Movement.InputForward = 1.0
	} else if input.IsActionPressed(settings.ACTION_BACK) {
		player.Movement.InputForward = -1.0
	} else {
		player.Movement.InputForward = 0.0
	}

	if input.IsActionPressed(settings.ACTION_RIGHT) {
		player.Movement.InputStrafe = 1.0
	} else if input.IsActionPressed(settings.ACTION_LEFT) {
		player.Movement.InputStrafe = -1.0
	} else {
		player.Movement.InputStrafe = 0.0
	}

	player.Movement.YawAngle -= input.ActionAxis(settings.ACTION_LOOK_HORZ)

	if player.FlyMode {
		player.Movement.PitchAngle -= input.ActionAxis(settings.ACTION_LOOK_VERT)
		player.Movement.PitchAngle = math2.Clamp(player.Movement.PitchAngle, -math2.HALF_PI, math2.HALF_PI)
	}
}
