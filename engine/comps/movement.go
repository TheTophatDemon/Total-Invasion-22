package comps

import (
	"github.com/go-gl/mathgl/mgl32"
	"tophatdemon.com/total-invasion-ii/engine/scene"
)

type Movement struct {
	MaxSpeed, Accel, Friction float32
	InputForward, InputStrafe float32
	YawAngle, PitchAngle      float32 //Radians
}

func (m *Movement) Update(transforms *scene.ComponentStorage[Transform], ent scene.Entity, deltaTime float32) {
	transform, ok := transforms.Get(ent)
	if !ok {
		return
	}

	strafe := m.InputStrafe * m.MaxSpeed * deltaTime
	forward := m.InputForward * m.MaxSpeed * deltaTime

	globalMove := mgl32.TransformCoordinate(mgl32.Vec3{strafe, 0.0, -forward}, transform.GetMatrix().Mat3().Mat4())
	transform.Translate(globalMove[0], globalMove[1], globalMove[2])
	transform.SetRotation(m.PitchAngle, m.YawAngle, 0.0)
}
