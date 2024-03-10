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

func (b Box) Extents() math2.Box {
	return b.extents
}

func (b Box) Raycast(rayOrigin, rayDir, shapeOffset mgl32.Vec3, maxDist float32) RaycastResult {
	var res RaycastResult = RayBoxCollision(rayOrigin, rayDir, b.extents.Translate(shapeOffset))
	if res.Distance <= maxDist {
		return res
	} else {
		return RaycastResult{}
	}
}

func (b Box) ResolveCollision(myPosition, theirPosition mgl32.Vec3, theirShape Shape) Result {
	panic("collision resolution not implemented for boxes")
}
