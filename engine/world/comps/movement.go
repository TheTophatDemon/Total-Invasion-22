package comps

import (
	"github.com/go-gl/mathgl/mgl32"
)

type Movement struct {
	Velocity mgl32.Vec3
}

func (m *Movement) Update(transform *Transform, deltaTime float32) {
	// TODO: Collision detection
	transform.Translate(m.Velocity[0]*deltaTime, m.Velocity[1]*deltaTime, m.Velocity[2]*deltaTime)
}
