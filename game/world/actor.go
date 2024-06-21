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
}

func (a *Actor) Update(deltaTime float32) {
	input := mgl32.Vec3{a.inputStrafe, 0.0, -a.inputForward}
	if input.LenSqr() != 0.0 {
		input = input.Normalize()
	}
	moveDir := mgl32.TransformCoordinate(input, a.body.Transform.Matrix().Mat3().Mat4())

	// Apply acceleration
	a.body.Velocity = a.body.Velocity.Add(moveDir.Mul(a.AccelRate * deltaTime))
	// Apply friction
	if speed := a.body.Velocity.Len(); speed > mgl32.Epsilon {
		frictionVec := a.body.Velocity.Mul(-min(speed, a.Friction*deltaTime) / speed)
		a.body.Velocity = a.body.Velocity.Add(frictionVec)
	}
	// Limit speed
	if speed := a.body.Velocity.Len(); speed > a.MaxSpeed && a.MaxSpeed > mgl32.Epsilon {
		a.body.Velocity = a.body.Velocity.Mul(a.MaxSpeed / speed)
	}
}

func (a *Actor) SetYaw(newYaw float32) {
	a.YawAngle = newYaw
}

func (a *Actor) FacingVec() mgl32.Vec3 {
	return mgl32.Vec3{
		math2.Sin(a.YawAngle),
		0.0,
		-math2.Cos(a.YawAngle),
	}
}

func (a *Actor) Body() *comps.Body {
	return &a.body
}
