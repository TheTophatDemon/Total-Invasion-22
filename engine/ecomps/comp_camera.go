package ecomps

import (
	"github.com/go-gl/mathgl/mgl32"
	"tophatdemon.com/total-invasion-ii/engine/ecs"
)

type Camera struct {
	projection mgl32.Mat4
}

func AddCamera(ent ecs.Entity, fovDegrees, aspectRatio, nearDist, farDist float32) {
	Cameras.Assign(ent, Camera{
		projection: mgl32.Perspective(mgl32.DegToRad(fovDegrees), aspectRatio, nearDist, farDist),
	})
}

func (c *Camera) GetProjectionMatrix() mgl32.Mat4 {
	return c.projection
}
