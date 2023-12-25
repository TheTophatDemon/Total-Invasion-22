package ents

import (
	"github.com/go-gl/mathgl/mgl32"
	"tophatdemon.com/total-invasion-ii/engine/color"
	"tophatdemon.com/total-invasion-ii/engine/input"
	"tophatdemon.com/total-invasion-ii/engine/math2"
	"tophatdemon.com/total-invasion-ii/engine/math2/collision"
	"tophatdemon.com/total-invasion-ii/engine/world/comps"

	"tophatdemon.com/total-invasion-ii/game/settings"
)

const (
	USE_DIST float32 = 4.0
)

type Player struct {
	Camera                                   comps.Camera
	RunSpeed, WalkSpeed                      float32
	StandFriction, WalkFriction, RunFriction float32
	actor                                    Actor
	world                                    WorldOps
}

var _ HasActor = (*Player)(nil)
var _ comps.HasBody = (*Player)(nil)

func (p *Player) Actor() *Actor {
	return &p.actor
}

func (p *Player) Body() *comps.Body {
	return &p.actor.body
}

func NewPlayer(position, angles mgl32.Vec3, world WorldOps) Player {
	return Player{
		actor: Actor{
			body: comps.Body{
				Transform: comps.TransformFromTranslationAngles(
					position, angles,
				),
				Shape:     collision.NewSphere(0.7),
				Pushiness: 10,
				NoClip:    false,
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
		world:         world,
	}
}

func (player *Player) Update(deltaTime float32) {
	if input.IsActionPressed(settings.ACTION_FORWARD) {
		player.actor.inputForward = 1.0
	} else if input.IsActionPressed(settings.ACTION_BACK) {
		player.actor.inputForward = -1.0
	} else {
		player.actor.inputForward = 0.0
	}

	if input.IsActionPressed(settings.ACTION_RIGHT) {
		player.actor.inputStrafe = 1.0
	} else if input.IsActionPressed(settings.ACTION_LEFT) {
		player.actor.inputStrafe = -1.0
	} else {
		player.actor.inputStrafe = 0.0
	}

	if input.IsActionJustPressed(settings.ACTION_NOCLIP) {
		player.Body().NoClip = !player.Body().NoClip
		message := "No-Clip "
		if player.Body().NoClip {
			message += "Activated"
		} else {
			message += "Deactivated"
		}
		player.world.ShowMessage(message, 4.0, 100, color.Red)
	}

	if input.IsActionPressed(settings.ACTION_USE) {
		rayOrigin := player.Body().Transform.Position()
		rayDir := mgl32.TransformNormal(math2.Vec3Forward(), player.Body().Transform.Matrix())
		hit, closestBody := player.world.Raycast(rayOrigin, rayDir, true, USE_DIST, player)
		if hit.Hit && closestBody != nil {
			if usable, isUsable := closestBody.(Usable); isUsable {
				usable.OnUse(player)
			}
		}
	}

	if input.IsActionPressed(settings.ACTION_SLOW) {
		player.actor.MaxSpeed = player.WalkSpeed
	} else {
		player.actor.MaxSpeed = player.RunSpeed
	}

	if math2.Abs(player.actor.inputForward) > mgl32.Epsilon || math2.Abs(player.actor.inputStrafe) > mgl32.Epsilon {
		if player.actor.MaxSpeed == player.WalkSpeed {
			player.actor.Friction = player.WalkFriction
		} else {
			player.actor.Friction = player.RunFriction
		}
	} else {
		player.actor.Friction = player.StandFriction
	}

	player.actor.YawAngle -= input.ActionAxis(settings.ACTION_LOOK_HORZ)
	player.Body().Transform.SetRotation(0.0, player.actor.YawAngle, 0.0)

	player.actor.Update(deltaTime)
}
