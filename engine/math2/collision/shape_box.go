package collision

import (
	"github.com/go-gl/mathgl/mgl32"
	"tophatdemon.com/total-invasion-ii/engine/math2"
)

type Box struct {
	shape
}

var _ Shape = (*Box)(nil)

func NewBox(extents math2.Box) Box {
	return Box{
		shape: shape{
			extents: extents,
		},
	}
}

func (box Box) String() string {
	return "Box"
}

func (box Box) Extents() math2.Box {
	return box.extents
}

func (box Box) Raycast(rayOrigin, rayDir, shapeOffset mgl32.Vec3, maxDist float32) RaycastResult {
	var res RaycastResult = RayBoxCollision(rayOrigin, rayDir, box.extents.Translate(shapeOffset))
	if res.Distance <= maxDist {
		return res
	} else {
		return RaycastResult{}
	}
}

func (box Box) ResolveCollision(myPosition, myMovement, theirPosition mgl32.Vec3, theirShape Shape) Result {
	nextPosition := myPosition.Add(myMovement)
	switch otherShape := theirShape.(type) {
	case Sphere:
		return ResolveSphereBox(nextPosition, myPosition, otherShape, box)
	default:
		panic("collision resolution not implemented for boxes and " + otherShape.String())
	}
}
