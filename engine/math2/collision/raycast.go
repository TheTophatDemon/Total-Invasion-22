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
