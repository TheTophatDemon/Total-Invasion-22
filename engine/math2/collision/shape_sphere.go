package collision

import (
	"github.com/go-gl/mathgl/mgl32"
	"tophatdemon.com/total-invasion-ii/engine/math2"
)

type Sphere struct {
	shape
	radius     float32
	continuous bool
}

var _ Shape = (*Sphere)(nil)

func NewSphere(radius float32) Sphere {
	return Sphere{
		shape: shape{
			extents: math2.BoxFromRadius(radius),
		},
		radius: radius,
	}
}

func NewContinuousSphere(radius float32) Sphere {
	sphere := NewSphere(radius)
	sphere.continuous = true
	return sphere
}

func (sphere Sphere) String() string {
	return "Sphere"
}

func (sphere Sphere) Extents() math2.Box {
	return sphere.extents
}

func (sphere Sphere) Radius() float32 {
	return sphere.radius
}

func (sphere Sphere) Raycast(rayOrigin, rayDir, shapeOffset mgl32.Vec3, maxDist float32) RaycastResult {
	var res RaycastResult = RaySphereCollision(rayOrigin, rayDir, shapeOffset, sphere.radius)
	if res.Distance <= maxDist {
		return res
	} else {
		return RaycastResult{}
	}
}

func (sphere Sphere) ResolveCollision(myPosition, myMovement, theirPosition mgl32.Vec3, theirShape Shape) Result {
	myNextPosition := myPosition.Add(myMovement)
	if myMovement.LenSqr() == 0.0 {
		return sphere.ResolveCollisionDiscrete(myNextPosition, theirPosition, theirShape)
	}

	boundingCenter := myPosition.Add(myMovement.Mul(0.5))
	boundingRadius := myMovement.Len() * 0.5

	switch otherShape := theirShape.(type) {
	case Sphere:

	case Box:
		return ResolveSphereBox(myPosition, theirPosition, sphere, otherShape)
	case Mesh:
		// Check for triangle hits in the center first, then the edges.
		// This prevents triangle edges from stopping smooth movement along neighboring triangles.
		var res1, res2 Result
		var firstHitPosition, originalPosition, overallMovement, overallNormal mgl32.Vec3
		var overallDistance float32

		originalPosition = myPosition
		res1 = ResolveSphereTriangles(myPosition, theirPosition, sphere, otherShape, TRI_PART_CENTER)
		if res1.Hit {
			firstHitPosition = res1.Position
			myPosition = myPosition.Add(res1.Normal.Mul(res1.Penetration))
		}
		res2 = ResolveSphereTriangles(myPosition, theirPosition, sphere, otherShape, TRI_PART_ALL)
		if res2.Hit {
			if !res1.Hit {
				firstHitPosition = res2.Position
			}
			myPosition = myPosition.Add(res2.Normal.Mul(res2.Penetration))
		}
		overallMovement = myPosition.Sub(originalPosition)
		overallDistance = overallMovement.Len()
		if overallDistance > 0.0 {
			overallNormal = overallMovement.Mul(1.0 / overallDistance)
		}
		return Result{
			Hit:         res1.Hit || res2.Hit,
			Position:    firstHitPosition,
			Normal:      overallNormal,
			Penetration: overallDistance,
		}
	case Grid:
		return otherShape.ResolveOtherBodysCollision(theirPosition, myPosition, myMovement, sphere)
	}
	return Result{}
}

func (sphere Sphere) ResolveCollisionDiscrete(myNextPosition, theirPosition mgl32.Vec3, theirShape Shape) Result {
	switch otherShape := theirShape.(type) {
	case Sphere:
		return ResolveSphereSphere(myNextPosition, theirPosition, sphere, otherShape)
	case Box:
		return ResolveSphereBox(myNextPosition, theirPosition, sphere, otherShape)
	case Mesh:
		// Check for triangle hits in the center first, then the edges.
		// This prevents triangle edges from stopping smooth movement along neighboring triangles.
		var res1, res2 Result
		var firstHitPosition, originalPosition, overallMovement, overallNormal mgl32.Vec3
		var overallDistance float32

		originalPosition = myNextPosition
		res1 = ResolveSphereTriangles(myNextPosition, theirPosition, sphere, otherShape, TRI_PART_CENTER)
		if res1.Hit {
			firstHitPosition = res1.Position
			myNextPosition = myNextPosition.Add(res1.Normal.Mul(res1.Penetration))
		}
		res2 = ResolveSphereTriangles(myNextPosition, theirPosition, sphere, otherShape, TRI_PART_ALL)
		if res2.Hit {
			if !res1.Hit {
				firstHitPosition = res2.Position
			}
			myNextPosition = myNextPosition.Add(res2.Normal.Mul(res2.Penetration))
		}
		overallMovement = myNextPosition.Sub(originalPosition)
		overallDistance = overallMovement.Len()
		if overallDistance > 0.0 {
			overallNormal = overallMovement.Mul(1.0 / overallDistance)
		}
		return Result{
			Hit:         res1.Hit || res2.Hit,
			Position:    firstHitPosition,
			Normal:      overallNormal,
			Penetration: overallDistance,
		}
	case Grid:
		return otherShape.ResolveOtherBodysCollision(theirPosition, myNextPosition, mgl32.Vec3{}, sphere)
	}
	return Result{}
}
