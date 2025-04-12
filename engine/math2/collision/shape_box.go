package collision

import (
	"github.com/go-gl/mathgl/mgl32"
	"tophatdemon.com/total-invasion-ii/engine/math2"
)

type Box struct {
	shape
}

var _ MovingShape = (*Box)(nil)

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
		return ResolveSphereBox(theirPosition, nextPosition, otherShape, box)
	case Box:
		//TODO: Box to box collision
	case Mesh:
		//TODO: Box to mesh collision
	case Cylinder:
		//TODO: Box to cylinder collision
	case Grid:
		return otherShape.ResolveOtherBodysCollision(theirPosition, myPosition, myMovement, box)
	default:
		panic("collision resolution not implemented for boxes and " + otherShape.String())
	}
	return Result{}
}

func (box Box) Touches(myPosition, theirPosition mgl32.Vec3, theirShape Shape) bool {
	switch otherShape := theirShape.(type) {
	case Sphere:
		return SphereTouchesBox(theirPosition, otherShape.radius, box.extents.Translate(myPosition))
	case Box:
		return box.extents.Translate(myPosition).Intersects(otherShape.extents.Translate(theirPosition))
	case Mesh:
		//TODO: Box to mesh touching
	case Grid:
		return otherShape.OtherBodyTouches(theirPosition, myPosition, box)
	}
	return false
}
