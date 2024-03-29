package collision

import (
	"math"

	"github.com/go-gl/mathgl/mgl32"
	"tophatdemon.com/total-invasion-ii/engine/math2"
)

type Result struct {
	Hit              bool
	Position, Normal mgl32.Vec3
	Penetration      float32
}

// Represents a bit mask that filters what things will collide with what.
type Mask uint64

const (
	MASK_EMPTY Mask = iota
	MASK_ALL        = 0xFFFFFFFFFFFFFFFF
)

// Represents the portion of a triangle a collision should involve, whether it's the edges or the center.
type TriParts uint8

const (
	TRI_PART_NONE   TriParts = 0
	TRI_PART_CENTER TriParts = 1 << (iota - 1)
	TRI_PART_EDGE
	TRI_PART_ALL = TRI_PART_CENTER + TRI_PART_EDGE
)

func SphereTriangleCollision(spherePos mgl32.Vec3, sphereRadius float32, triangle math2.Triangle, trianglesOffset mgl32.Vec3) (TriParts, Result) {
	spherePos = spherePos.Sub(trianglesOffset)

	// Check if the sphere intersects the triangle's plane.
	plane := triangle.Plane()
	distToPlane := plane.Normal.Dot(spherePos.Sub(triangle[0]))
	if distToPlane < 0 || distToPlane > sphereRadius {
		// Does not intersect plane of the triangle
		return TRI_PART_NONE, Result{}
	}

	// Check if the projected sphere center is within the bounds of the triangle.
	centerProj := spherePos.Sub(plane.Normal.Mul(distToPlane))
	c0 := centerProj.Sub(triangle[0]).Cross(triangle[1].Sub(triangle[0]))
	c1 := centerProj.Sub(triangle[1]).Cross(triangle[2].Sub(triangle[1]))
	c2 := centerProj.Sub(triangle[2]).Cross(triangle[0].Sub(triangle[2]))
	if c0.Dot(plane.Normal) <= 0.0 && c1.Dot(plane.Normal) <= 0.0 && c2.Dot(plane.Normal) <= 0.0 {
		// Center of the sphere is inside of the triangle's edges.
		return TRI_PART_CENTER, Result{
			Hit:         true,
			Position:    centerProj.Add(trianglesOffset),
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
	for e := range 3 {
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
			return TRI_PART_NONE, Result{}
		}

		pushNormal := minEdge.diff.Mul(1.0 / dist)
		return TRI_PART_EDGE, Result{
			Hit:         true,
			Position:    minEdge.closest.Add(trianglesOffset),
			Normal:      pushNormal,
			Penetration: sphereRadius - dist,
		}
	}

	return TRI_PART_NONE, Result{}
}

func SphereTouchesSphere(pos1 mgl32.Vec3, radius1 float32, pos2 mgl32.Vec3, radius2 float32) bool {
	return pos1.Sub(pos2).LenSqr() < (radius1+radius2)*(radius1+radius2)
}

func ResolveSphereSphere(sphere1Pos, sphere2Pos mgl32.Vec3, sphere1, sphere2 Sphere) (result Result) {
	diff := sphere1Pos.Sub(sphere2Pos)
	distSq := diff.LenSqr()

	if distSq < (sphere1.Radius()+sphere2.Radius())*(sphere1.Radius()+sphere2.Radius()) && distSq != 0.0 {
		result.Hit = true
		dist := math2.Sqrt(distSq)
		result.Normal = diff.Mul(1.0 / dist)
		result.Penetration = sphere1.Radius() + sphere2.Radius() - dist
		result.Position = sphere1Pos.Add(result.Normal.Mul(-sphere1.Radius()))
	}

	return
}

func SphereTouchesBox(spherePos mgl32.Vec3, sphereRadius float32, box math2.Box) bool {
	projectedPoint := math2.Vec3Max(math2.Vec3Min(spherePos, box.Max), box.Min)
	diff := spherePos.Sub(projectedPoint)
	distSq := diff.LenSqr()

	if distSq > 0.0 && distSq < sphereRadius*sphereRadius {
		return true
	}

	return false
}

func ResolveSphereBox(spherePos, boxPos mgl32.Vec3, sphere Sphere, box Box) (result Result) {
	worldSpaceBox := box.Extents().Translate(boxPos)
	projectedPoint := math2.Vec3Max(math2.Vec3Min(spherePos, worldSpaceBox.Max), worldSpaceBox.Min)
	diff := spherePos.Sub(projectedPoint)
	distSq := diff.LenSqr()

	if distSq > 0.0 && distSq < sphere.Radius()*sphere.Radius() {
		result.Hit = true
		dist := math2.Sqrt(distSq)
		result.Normal = diff.Mul(1.0 / dist)
		result.Penetration = sphere.Radius() - dist
		result.Position = spherePos.Add(result.Normal.Mul(-sphere.Radius()))
	} else if distSq == 0.0 {
		result.Hit = true
		diffToCenter := spherePos.Sub(boxPos)
		distToCenter := diffToCenter.Len()
		result.Normal = diffToCenter.Mul(1.0 / distToCenter)
		result.Penetration = sphere.Radius() - distToCenter
		result.Position = spherePos.Add(result.Normal.Mul(-sphere.Radius()))
	}

	return
}

func ResolveSphereTriangles(spherePos, meshPos mgl32.Vec3, sphere Sphere, mesh Mesh, filter TriParts) (result Result) {
	if filter == TRI_PART_NONE {
		return
	}
	for _, triangle := range mesh.triangles {
		hit, col := SphereTriangleCollision(spherePos, sphere.Radius(), triangle, meshPos)
		if hit&filter != 0 {
			result = col
			result.Penetration += mgl32.Epsilon
			return
		}
	}
	return
}
