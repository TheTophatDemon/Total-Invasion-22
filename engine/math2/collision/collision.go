package collision

import (
	"math"

	"github.com/go-gl/mathgl/mgl32"
	"tophatdemon.com/total-invasion-ii/engine/math2"
)

type Result struct {
	Position, Normal mgl32.Vec3
	Penetration      float32
}

type TriangleHit uint8

const (
	TRIHIT_NONE TriangleHit = 1 << iota
	TRIHIT_CENTER
	TRIHIT_EDGE
	TRIHIT_ALL = TRIHIT_CENTER + TRIHIT_EDGE
)

func SphereTriangleCollision(spherePos mgl32.Vec3, sphereRadius float32, triangle math2.Triangle) (TriangleHit, Result) {
	// Check if the sphere intersects the triangle's plane.
	plane := triangle.Plane()
	distToPlane := plane.Normal.Dot(spherePos.Sub(triangle[0]))
	if distToPlane < 0 || distToPlane < -sphereRadius || distToPlane > sphereRadius {
		// Does not intersect plane of the triangle
		return TRIHIT_NONE, Result{}
	}

	// Check if the projected sphere center is within the bounds of the triangle.
	centerProj := spherePos.Sub(plane.Normal.Mul(distToPlane))
	c0 := centerProj.Sub(triangle[0]).Cross(triangle[1].Sub(triangle[0]))
	c1 := centerProj.Sub(triangle[1]).Cross(triangle[2].Sub(triangle[1]))
	c2 := centerProj.Sub(triangle[2]).Cross(triangle[0].Sub(triangle[2]))
	if c0.Dot(plane.Normal) <= 0.0 && c1.Dot(plane.Normal) <= 0.0 && c2.Dot(plane.Normal) <= 0.0 {
		// Center of the sphere is inside of the triangle's edges.
		return TRIHIT_CENTER, Result{
			Position:    centerProj,
			Normal:      plane.Normal,
			Penetration: sphereRadius - distToPlane,
		}
	}

	// Find the closest point on the closest edge of the triangle.
	var minEdge struct {
		closest, diff mgl32.Vec3
		distSqr       float32
	}
	minEdge.distSqr = math.MaxFloat32
	for e := 0; e < 3; e += 1 {
		closest := math2.ClosestPointOnLine(triangle[e], triangle[(e+1)%3], spherePos)
		diff := spherePos.Sub(closest)
		distSqr := diff.LenSqr()
		if distSqr < minEdge.distSqr {
			minEdge.closest = closest
			minEdge.diff = diff
			minEdge.distSqr = distSqr
		}
	}

	if minEdge.distSqr < sphereRadius*sphereRadius {
		// Sphere intersects with closest edge.
		dist := math2.Sqrt(minEdge.distSqr)
		if dist == 0.0 {
			return TRIHIT_NONE, Result{}
		}

		pushNormal := minEdge.diff.Mul(1.0 / dist)
		return TRIHIT_EDGE, Result{
			Position:    minEdge.closest,
			Normal:      pushNormal,
			Penetration: sphereRadius - dist,
		}
	}

	return TRIHIT_NONE, Result{}
}
