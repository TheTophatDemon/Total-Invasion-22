package collision

import (
	"github.com/go-gl/mathgl/mgl32"
	"tophatdemon.com/total-invasion-ii/engine/math2"
)

type Cylinder struct {
	shape
	radius, halfHeight float32
}

var _ Shape = (*Cylinder)(nil)

func NewCylinder(radius, height float32) Cylinder {
	halfHeight := height / 2.0
	return Cylinder{
		shape: shape{
			extents: math2.Box{
				Max: mgl32.Vec3{radius, halfHeight, radius},
				Min: mgl32.Vec3{-radius, -halfHeight, -radius},
			},
		},
		radius:     radius,
		halfHeight: halfHeight,
	}
}

func (cylinder Cylinder) String() string {
	return "Cylinder"
}

func (cylinder Cylinder) Extents() math2.Box {
	return cylinder.extents
}

func (cylinder Cylinder) Radius() float32 {
	return cylinder.radius
}

func (cylinder Cylinder) Height() float32 {
	return cylinder.halfHeight * 2.0
}

func (cylinder Cylinder) Raycast(rayOrigin, rayDir, shapeOffset mgl32.Vec3, maxDist float32) RaycastResult {
	//TODO
	return RaycastResult{}
}
