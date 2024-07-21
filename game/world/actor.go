package world

import (
	"github.com/go-gl/mathgl/mgl32"
	"tophatdemon.com/total-invasion-ii/engine/math2"
	"tophatdemon.com/total-invasion-ii/engine/scene/comps"
)

type Actor struct {
	body                          comps.Body
	MaxSpeed, AccelRate, Friction float32
	inputForward, inputStrafe     float32
	YawAngle                      float32 // Radians
	Health, MaxHealth             float32
}

func (actor *Actor) Update(deltaTime float32) {
	input := mgl32.Vec3{actor.inputStrafe, 0.0, -actor.inputForward}
	if input.LenSqr() != 0.0 {
		input = input.Normalize()
	}
	moveDir := mgl32.TransformCoordinate(input, mgl32.HomogRotate3DY(actor.YawAngle))

	// Apply acceleration
	actor.body.Velocity = actor.body.Velocity.Add(moveDir.Mul(actor.AccelRate * deltaTime))
	// Apply friction
	if speed := actor.body.Velocity.Len(); speed > mgl32.Epsilon {
		frictionVec := actor.body.Velocity.Mul(-min(speed, actor.Friction*deltaTime) / speed)
		actor.body.Velocity = actor.body.Velocity.Add(frictionVec)
	}
	// Limit speed
	if speed := actor.body.Velocity.Len(); speed > actor.MaxSpeed && actor.MaxSpeed > mgl32.Epsilon {
		actor.body.Velocity = actor.body.Velocity.Mul(actor.MaxSpeed / speed)
	}
}

func (actor *Actor) SetYaw(newYaw float32) {
	actor.YawAngle = newYaw
}

func (actor *Actor) FacingVec() mgl32.Vec3 {
	return mgl32.Vec3{
		-math2.Sin(actor.YawAngle),
		0.0,
		-math2.Cos(actor.YawAngle),
	}
}

func (actor *Actor) Body() *comps.Body {
	return &actor.body
}

func (actor *Actor) Position() mgl32.Vec3 {
	return actor.body.Transform.Position()
}
