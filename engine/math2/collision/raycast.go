package collision

import (
	"math"

	"github.com/go-gl/mathgl/mgl32"
	"tophatdemon.com/total-invasion-ii/engine/math2"
)

type RaycastResult struct {
	Hit              bool
	Position, Normal mgl32.Vec3
	Distance         float32
}

func RayPlaneCollision(rayOrigin, rayDir mgl32.Vec3, plane math2.Plane, twoSided bool) RaycastResult {
	if denom := rayDir.Dot(plane.Normal); math2.Abs(denom) > 0.0001 && (twoSided || denom < 0) {
		t := plane.Normal.Mul(plane.Dist).Sub(rayOrigin).Dot(plane.Normal) / denom
		if t > 0.0001 {
			return RaycastResult{
				Hit:      true,
				Position: rayOrigin.Add(rayDir.Mul(t)),
				Normal:   plane.Normal.Mul(-math2.Signum(denom)),
				Distance: t,
			}
		}
	}
	return RaycastResult{}
}

func RayBoxCollision(rayOrigin, rayDir mgl32.Vec3, box math2.Box) RaycastResult {
	// Tests the ray against an axis aligned plane.
	testAxisPlane := func(axis int, planeOffset float32) (bool, mgl32.Vec3, float32) {
		if rayDir[axis] == 0.0 {
			return false, mgl32.Vec3{}, 0.0
		}
		t := (planeOffset - rayOrigin[axis]) / rayDir[axis]

		var proj mgl32.Vec3
		proj[axis] = planeOffset
		otherAxis := (axis + 2) % 3
		proj[otherAxis] = rayOrigin[otherAxis] + rayDir[otherAxis]*t
		otherOtherAxis := (axis + 1) % 3
		proj[otherOtherAxis] = rayOrigin[otherOtherAxis] + rayDir[otherOtherAxis]*t

		if proj[otherAxis] >= box.Min[otherAxis] && proj[otherAxis] < box.Max[otherAxis] && proj[otherOtherAxis] >= box.Min[otherOtherAxis] && proj[otherOtherAxis] < box.Max[otherOtherAxis] {
			return true, proj, proj.Sub(rayOrigin).LenSqr()
		}

		return false, mgl32.Vec3{}, 0.0
	}

	var result RaycastResult
	var nearestPlaneDist float32 = math.MaxFloat32

	// Left side
	if hit, proj, distSqr := testAxisPlane(0, box.Min.X()); hit && distSqr < nearestPlaneDist {
		nearestPlaneDist = distSqr
		result = RaycastResult{
			Hit:      true,
			Position: proj,
			Normal:   mgl32.Vec3{-1.0, 0.0, 0.0},
			Distance: math2.Sqrt(distSqr),
		}
	}

	// Right side
	if hit, proj, distSqr := testAxisPlane(0, box.Max.X()); hit && distSqr < nearestPlaneDist {
		nearestPlaneDist = distSqr
		result = RaycastResult{
			Hit:      true,
			Position: proj,
			Normal:   mgl32.Vec3{1.0, 0.0, 0.0},
			Distance: math2.Sqrt(distSqr),
		}
	}

	// Bottom
	if hit, proj, distSqr := testAxisPlane(1, box.Min.Y()); hit && distSqr < nearestPlaneDist {
		nearestPlaneDist = distSqr
		result = RaycastResult{
			Hit:      true,
			Position: proj,
			Normal:   mgl32.Vec3{0.0, -1.0, 0.0},
			Distance: math2.Sqrt(distSqr),
		}
	}

	// Top
	if hit, proj, distSqr := testAxisPlane(1, box.Max.Y()); hit && distSqr < nearestPlaneDist {
		nearestPlaneDist = distSqr
		result = RaycastResult{
			Hit:      true,
			Position: proj,
			Normal:   mgl32.Vec3{0.0, 1.0, 0.0},
			Distance: math2.Sqrt(distSqr),
		}
	}

	// Front
	if hit, proj, distSqr := testAxisPlane(2, box.Min.Z()); hit && distSqr < nearestPlaneDist {
		nearestPlaneDist = distSqr
		result = RaycastResult{
			Hit:      true,
			Position: proj,
			Normal:   mgl32.Vec3{0.0, 0.0, -1.0},
			Distance: math2.Sqrt(distSqr),
		}
	}

	// Back
	if hit, proj, distSqr := testAxisPlane(2, box.Max.Z()); hit && distSqr < nearestPlaneDist {
		nearestPlaneDist = distSqr
		result = RaycastResult{
			Hit:      true,
			Position: proj,
			Normal:   mgl32.Vec3{0.0, 0.0, 1.0},
			Distance: math2.Sqrt(distSqr),
		}
	}

	return result
}

func RaySphereCollision(rayOrigin, rayDir, spherePos mgl32.Vec3, sphereRadius float32) RaycastResult {
	// Yes, this is the same algorithm that Raylib uses...
	diff := spherePos.Sub(rayOrigin)
	dist := diff.Len()
	dp := diff.Dot(rayDir)
	d := sphereRadius*sphereRadius - (dist*dist - dp*dp)
	if d < 0.0 {
		return RaycastResult{}
	}

	if dist < sphereRadius {
		// Ray is inside of the sphere
		collisionDist := dp + math2.Sqrt(d)
		pos := rayOrigin.Add(rayDir.Mul(collisionDist))
		return RaycastResult{
			Hit:      true,
			Position: pos,
			Normal:   pos.Sub(spherePos).Normalize(), // Normal points inwards
			Distance: collisionDist,
		}
	} else {
		// Ray is outside of the sphere
		collisionDist := dp - math2.Sqrt(d)
		pos := rayOrigin.Add(rayDir.Mul(collisionDist))
		return RaycastResult{
			Hit:      true,
			Position: pos,
			Normal:   pos.Sub(spherePos).Normalize(), // Normal points outwards
			Distance: collisionDist,
		}
	}
}

func RayCylinderCollision(rayOrigin, rayDir, cylinderPos mgl32.Vec3, cylinderRadius, cylinderHalfHeight float32) RaycastResult {
	var endNormal mgl32.Vec3
	var endDifference float32
	if endDifference = cylinderPos[1] + cylinderHalfHeight - rayOrigin[1]; endDifference < 0.0 {
		// Ray starts from above cylinder
		if rayDir[1] >= 0.0 {
			// Pointing away
			return RaycastResult{}
		}

		endNormal[1] = 1.0
	} else if endDifference = cylinderPos[1] - cylinderHalfHeight - rayOrigin[1]; endDifference > 0.0 {
		// Ray starts from below cylinder
		if rayDir[1] <= 0.0 {
			// Pointing away
			return RaycastResult{}
		}
		endNormal[1] = -1.0
	}
	if endDifference != 0.0 && endNormal != (mgl32.Vec3{}) {
		// Find point of contact with plane on top/bottom of cylinder
		t := endDifference / rayDir[1]
		// X and Z distance from center of cylinder
		dx := (rayOrigin[0] + rayDir[0]*t) - cylinderPos[0]
		dz := (rayOrigin[2] + rayDir[2]*t) - cylinderPos[2]
		if (dx*dx)+(dz*dz) < cylinderRadius*cylinderRadius {
			// Ray hits plane within the circle on top/bottom of the cylinder.
			return RaycastResult{
				Hit:      true,
				Position: mgl32.Vec3{dx + cylinderPos[0], endDifference + rayOrigin[1], dz + cylinderPos[2]},
				Normal:   endNormal,
				Distance: t,
			}
		}
	}

	// See if the ray hits the circle the cylinder projects onto the XZ plane.
	diff := mgl32.Vec2{cylinderPos[0] - rayOrigin[0], cylinderPos[2] - rayOrigin[2]}
	dist := diff.Len()
	dp := diff.Dot(mgl32.Vec2{rayDir[0], rayDir[2]})
	d := cylinderRadius*cylinderRadius - (dist*dist - dp*dp)
	if d < 0.0 {
		return RaycastResult{}
	}

	// The ray reaches the cylinder on the XZ plane, but now we must check the height.

	return RaycastResult{}
}

func RayTriangleCollision(rayOrigin, rayDir mgl32.Vec3, tri math2.Triangle) RaycastResult {
	// Uses the Trumbore intersection algorithm
	// https://en.wikipedia.org/wiki/M%C3%B6ller%E2%80%93Trumbore_intersection_algorithm

	edge1 := tri[1].Sub(tri[0])
	edge2 := tri[2].Sub(tri[0])
	triNormal := rayDir.Cross(edge2)
	det := edge1.Dot(triNormal)

	if det > -mgl32.Epsilon && det < mgl32.Epsilon {
		return RaycastResult{}
	}

	invDet := 1.0 / det
	s := rayOrigin.Sub(tri[0])
	u := invDet * s.Dot(triNormal)

	if u < 0.0 || u > 1.0 {
		return RaycastResult{}
	}

	sEdge1Cross := s.Cross(edge1)
	v := invDet * rayDir.Dot(sEdge1Cross)

	if v < 0.0 || u+v > 1.0 {
		return RaycastResult{}
	}

	t := invDet * edge2.Dot(sEdge1Cross)

	if t > mgl32.Epsilon {
		return RaycastResult{
			Hit:      true,
			Position: rayOrigin.Add(rayDir.Mul(t)),
			Normal:   triNormal,
			Distance: t,
		}
	}

	return RaycastResult{}
}
