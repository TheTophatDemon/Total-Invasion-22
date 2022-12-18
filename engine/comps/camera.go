package comps

import (
	"github.com/go-gl/mathgl/mgl32"
	"tophatdemon.com/total-invasion-ii/engine/scene"
)

type Camera struct {
	projection mgl32.Mat4
}

func (c *Camera) Update(ent *scene.Entity, deltaTime float32) {
}

func NewCamera(fovDegrees, aspectRatio, nearDist, farDist float32) *Camera {
	return &Camera{
		projection: mgl32.Perspective(mgl32.DegToRad(fovDegrees), aspectRatio, nearDist, farDist),
	}
}

func (c *Camera) GetProjectionMatrix() mgl32.Mat4 {
	return c.projection
}
