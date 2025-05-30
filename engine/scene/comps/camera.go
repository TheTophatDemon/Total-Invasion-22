package comps

import (
	"github.com/go-gl/mathgl/mgl32"
)

type Camera struct {
	Transform
	projection mgl32.Mat4
}

func NewCamera(fovDegrees, aspectRatio, nearDist, farDist float32, transform Transform) Camera {
	return Camera{
		projection: mgl32.Perspective(mgl32.DegToRad(fovDegrees), aspectRatio, nearDist, farDist),
		Transform:  transform,
	}
}

func (c *Camera) ProjectionMatrix() mgl32.Mat4 {
	return c.projection
}
