package ents

import (
	"github.com/go-gl/mathgl/mgl32"
	"tophatdemon.com/total-invasion-ii/engine/world/comps"
)

type Actor struct {
	Transform                     comps.Transform
	Movement                      comps.Movement
	MaxSpeed, AccelRate, Friction float32
	inputForward, inputStrafe     float32
	YawAngle                      float32 // Radians
}

func (a *Actor) Update(deltaTime float32) {
	input := mgl32.Vec3{a.inputStrafe, 0.0, -a.inputForward}
	if input.LenSqr() != 0.0 {
		input = input.Normalize()
	}
	moveDir := mgl32.TransformCoordinate(input, a.Transform.GetMatrix().Mat3().Mat4())

	// Apply acceleration
	a.Movement.Velocity = a.Movement.Velocity.Add(moveDir.Mul(a.AccelRate * deltaTime))
	// Apply friction
	if speed := a.Movement.Velocity.Len(); speed > mgl32.Epsilon {
		frictionVec := a.Movement.Velocity.Mul(-min(speed, a.Friction*deltaTime) / speed)
		a.Movement.Velocity = a.Movement.Velocity.Add(frictionVec)
	}
	// Limit speed
	if speed := a.Movement.Velocity.Len(); speed > a.MaxSpeed && a.MaxSpeed > mgl32.Epsilon {
		a.Movement.Velocity = a.Movement.Velocity.Mul(a.MaxSpeed / speed)
	}

	a.Movement.Update(&a.Transform, deltaTime)
}
