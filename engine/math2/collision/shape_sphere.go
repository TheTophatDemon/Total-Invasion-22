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

var _ MovingShape = (*Sphere)(nil)

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
	if !sphere.continuous {
		myNextPosition := myPosition.Add(myMovement)
		return sphere.ResolveCollisionDiscrete(myNextPosition, theirPosition, theirShape)
	}
	return sphere.ResolveCollisionContinuous(myPosition, myMovement, theirPosition, theirShape)
}

func (sphere Sphere) ResolveCollisionContinuous(myPosition, myMovement, theirPosition mgl32.Vec3, theirShape Shape) Result {
	myNextPosition := myPosition.Add(myMovement)
	if myMovement.LenSqr() < sphere.radius*sphere.radius*0.25 {
		// When the sample point is less than half the sphere's distance away from the last one,
		// Evaluate the collision for real
		return sphere.ResolveCollisionDiscrete(myNextPosition, theirPosition, theirShape)
	}

	halfMovement := myMovement.Mul(0.5)
	boundingCenter := myPosition.Add(halfMovement)
	boundingSphere := NewSphere((myMovement.Len() * 0.5) + sphere.radius)

	if boundingSphere.Touches(boundingCenter, theirPosition, theirShape) {
		if res := sphere.ResolveCollisionContinuous(myPosition, halfMovement, theirPosition, theirShape); res.Hit {
			return res
		}
		return sphere.ResolveCollisionContinuous(boundingCenter, halfMovement, theirPosition, theirShape)
	}

	return Result{}
}

func (sphere Sphere) ResolveCollisionDiscrete(myNextPosition, theirPosition mgl32.Vec3, theirShape Shape) Result {
	switch otherShape := theirShape.(type) {
	case Sphere:
		return ResolveSphereSphere(myNextPosition, theirPosition, sphere, otherShape)
	case Box:
		return ResolveSphereBox(myNextPosition, theirPosition, sphere, otherShape)
	case Cylinder:
		return ResolveSphereCylinder(myNextPosition, theirPosition, sphere, otherShape)
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

func (sphere Sphere) Touches(myPosition, theirPosition mgl32.Vec3, theirShape Shape) bool {
	switch otherShape := theirShape.(type) {
	case Sphere:
		return SphereTouchesSphere(myPosition, sphere.radius, theirPosition, otherShape.radius)
	case Box:
		return SphereTouchesBox(myPosition, sphere.radius, otherShape.Extents().Translate(theirPosition))
	case Cylinder:
		return SphereTouchesCylinder(myPosition, sphere.radius, theirPosition, otherShape.radius, otherShape.halfHeight)
	case Mesh:
		for _, triangle := range otherShape.triangles {
			if hit, _ := SphereTriangleCollision(myPosition, sphere.Radius(), triangle, theirPosition); hit != TRI_PART_NONE {
				return true
			}
		}
	case Grid:
		if otherShape.OtherBodyTouches(theirPosition, myPosition, sphere) {
			return true
		}
	default:
		panic("Sphere.Touches must be implemented for shape " + otherShape.String())
	}
	return false
}
