package ecomps

import (
	"github.com/go-gl/mathgl/mgl32"
)

type Camera struct {
	projection mgl32.Mat4
}

func NewCamera(fovDegrees, aspectRatio, nearDist, farDist float32) Camera {
	return Camera{
		projection: mgl32.Perspective(mgl32.DegToRad(fovDegrees), aspectRatio, nearDist, farDist),
	}
}

func (c *Camera) GetProjectionMatrix() mgl32.Mat4 {
	return c.projection
}
