package math2

import (
	"math"

	"github.com/go-gl/mathgl/mgl32"
)

const (
	CORNER_BLB = iota // Back left bottom
	CORNER_BRB        // Back right bottom
	CORNER_BRT        // Back right top
	CORNER_FRT        // Front right top
	CORNER_FLB        // Front left bottom
	CORNER_FLT        // Front left top
	CORNER_BLT        // Back left top
	CORNER_FRB        // Front right bottom
)

const (
	PLANE_NEAR = iota
	PLANE_FAR
	PLANE_LEFT
	PLANE_RIGHT
	PLANE_TOP
	PLANE_BOTTOM
)

type (
	Rect struct {
		X, Y, Width, Height float32
	}

	Triangle [3]mgl32.Vec3

	Plane struct {
		Normal mgl32.Vec3
		Dist   float32
	}

	Frustum struct {
		Planes  [6]Plane
		Corners [8]mgl32.Vec3
	}

	Box struct {
		Min, Max mgl32.Vec3
	}
)

// Generate a rectangle that wraps around all of the given points (there must be at least 2).
func RectFromPoints(point0, point1 mgl32.Vec2, points ...mgl32.Vec2) Rect {
	minX, minY := min(point0.X(), point1.X()), min(point0.Y(), point1.Y())
	maxX, maxY := max(point0.X(), point1.X()), max(point0.Y(), point1.Y())
	for _, p := range points {
		minX = min(minX, p.X())
		minY = min(minY, p.Y())
		maxX = max(maxX, p.X())
		maxY = max(maxY, p.Y())
	}
	return Rect{
		X:      minX,
		Y:      minY,
		Width:  maxX - minX,
		Height: maxY - minY,
	}
}

func (r Rect) Center() (float32, float32) {
	return (r.X + r.Width/2.0), (r.Y + r.Height/2.0)
}

func (r Rect) Vec4() mgl32.Vec4 {
	return mgl32.Vec4{r.X, r.Y, r.Width, r.Height}
}

func (tri Triangle) Plane() Plane {
	normal := tri[1].Sub(tri[0]).Cross(tri[2].Sub(tri[0])).Normalize()
	return Plane{
		Normal: normal,
		Dist:   -normal.Dot(tri[0]),
	}
}

func (tri Triangle) OffsetBy(offset mgl32.Vec3) Triangle {
	return Triangle{
		tri[0].Add(offset),
		tri[1].Add(offset),
		tri[2].Add(offset),
	}
}

func PlaneFromPointAndNormal(point, normal mgl32.Vec3) Plane {
	return Plane{
		Normal: normal,
		Dist:   normal.Dot(point),
	}
}

func FrustumFromMatrices(invViewProj mgl32.Mat4) (frustum Frustum) {
	corners := [...]mgl32.Vec3{
		CORNER_BLB: {-1.0, -1.0, -1.0},
		CORNER_BRB: {1.0, -1.0, -1.0},
		CORNER_BRT: {1.0, 1.0, -1.0},
		CORNER_FRT: {1.0, 1.0, 1.0},
		CORNER_FLB: {-1.0, -1.0, 1.0},
		CORNER_FLT: {-1.0, 1.0, 1.0},
		CORNER_BLT: {-1.0, 1.0, -1.0},
		CORNER_FRB: {1.0, -1.0, 1.0},
	}
	for i := range corners {
		corners[i] = mgl32.TransformCoordinate(corners[i], invViewProj)
	}
	frustum.Planes[PLANE_NEAR] = PlaneFromPointAndNormal(corners[CORNER_BLB], corners[CORNER_BRB].Sub(corners[CORNER_BLB]).Cross(corners[CORNER_BLT].Sub(corners[CORNER_BLB])).Normalize())
	frustum.Planes[PLANE_FAR] = PlaneFromPointAndNormal(corners[CORNER_FRT], corners[CORNER_FRB].Sub(corners[CORNER_FRT]).Cross(corners[CORNER_FLT].Sub(corners[CORNER_FRT])).Normalize())
	frustum.Planes[PLANE_LEFT] = PlaneFromPointAndNormal(corners[CORNER_BLB], corners[CORNER_BLT].Sub(corners[CORNER_BLB]).Cross(corners[CORNER_FLB].Sub(corners[CORNER_BLB])).Normalize())
	frustum.Planes[PLANE_RIGHT] = PlaneFromPointAndNormal(corners[CORNER_BRB], corners[CORNER_BRT].Sub(corners[CORNER_FRT]).Cross(corners[CORNER_FRB].Sub(corners[CORNER_FRT])).Normalize())
	frustum.Planes[PLANE_TOP] = PlaneFromPointAndNormal(corners[CORNER_FLT], corners[CORNER_BLT].Sub(corners[CORNER_FLT]).Cross(corners[CORNER_FRT].Sub(corners[CORNER_FLT])).Normalize())
	frustum.Planes[PLANE_BOTTOM] = PlaneFromPointAndNormal(corners[CORNER_FLB], corners[CORNER_BLB].Sub(corners[CORNER_BRB]).Cross(corners[CORNER_FRB].Sub(corners[CORNER_BRB])).Normalize())
	frustum.Corners = corners
	return
}

func (p Plane) SignedDistance(point mgl32.Vec3) float32 {
	return point.Dot(p.Normal) - p.Dist
}

func (p Plane) IsPointInFront(point mgl32.Vec3) bool {
	return p.SignedDistance(point) > 0
}

func (frustum Frustum) ContainsPoint(point mgl32.Vec3) bool {
	for i := range frustum.Planes {
		if frustum.Planes[i].IsPointInFront(point) {
			return false
		}
	}
	return true
}

// Returns true if the given box intersects the frustum.
func (frustum Frustum) IntersectsBox(box Box) bool {
	for i := range frustum.Planes {
		if !box.IntersectsPlane(frustum.Planes[i]) {
			return false
		}
	}
	return true
}

func (frustum Frustum) IntersectsSphere(point mgl32.Vec3, radius float32) bool {
	for i := range frustum.Planes {
		if frustum.Planes[i].SignedDistance(point) > radius {
			return false
		}
	}
	return true
}

func BoxFromExtents(halfWidth, halfHeight, halfLength float32) Box {
	return Box{
		Max: mgl32.Vec3{halfWidth, halfHeight, halfLength},
		Min: mgl32.Vec3{-halfWidth, -halfHeight, -halfLength},
	}
}

func BoxFromRadius(radius float32) Box {
	return Box{
		Max: mgl32.Vec3{radius, radius, radius},
		Min: mgl32.Vec3{-radius, -radius, -radius},
	}
}

func BoxFromPoints(points ...mgl32.Vec3) Box {
	min := mgl32.Vec3{math.MaxFloat32, math.MaxFloat32, math.MaxFloat32}
	max := mgl32.Vec3{-math.MaxFloat32, -math.MaxFloat32, -math.MaxFloat32}
	for i := range points {
		min = Vec3Min(min, points[i])
		max = Vec3Max(max, points[i])
	}
	return Box{
		Min: min,
		Max: max,
	}
}

// Returns true if two boxes intersect.
func (box Box) Intersects(other Box) bool {
	return box.Max[0] > other.Min[0] && box.Max[1] > other.Min[1] && box.Max[2] > other.Min[2] &&
		other.Max[0] > box.Min[0] && other.Max[1] > box.Min[1] && other.Max[2] > box.Min[2]
}

func (box Box) IntersectsPlane(plane Plane) bool {
	center := box.Center()
	extents := box.Max.Sub(center)
	r := extents.X()*Abs(plane.Normal.X()) +
		extents.Y()*Abs(plane.Normal.Y()) +
		extents.Z()*Abs(plane.Normal.Z())

	return plane.SignedDistance(center) <= r
}

// Returns vectors representing all 8 corners of the box.
func (box Box) Corners() [8]mgl32.Vec3 {
	return [8]mgl32.Vec3{
		CORNER_FLB: box.Min,
		CORNER_BRT: box.Max,
		CORNER_BLT: {box.Min.X(), box.Max.Y(), box.Max.Z()},
		CORNER_BLB: {box.Min.X(), box.Min.Y(), box.Max.Z()},
		CORNER_FRB: {box.Max.X(), box.Min.Y(), box.Min.Z()},
		CORNER_FRT: {box.Max.X(), box.Max.Y(), box.Min.Z()},
		CORNER_BRB: {box.Max.X(), box.Min.Y(), box.Max.Z()},
		CORNER_FLT: {box.Min.X(), box.Max.Y(), box.Min.Z()},
	}
}

func (box Box) ContainsPoint(point mgl32.Vec3) bool {
	return true &&
		point.X() > box.Min.X() &&
		point.Y() > box.Min.Y() &&
		point.Z() > box.Min.Z() &&
		point.X() < box.Max.X() &&
		point.Y() < box.Max.Y() &&
		point.Z() < box.Max.Z() &&
		true
}

func (box Box) Translate(offset mgl32.Vec3) Box {
	return Box{
		Min: box.Min.Add(offset),
		Max: box.Max.Add(offset),
	}
}

func (box Box) Size() mgl32.Vec3 {
	return box.Max.Sub(box.Min)
}

func (box Box) LongestDimension() float32 {
	var dims [3]float32 = box.Size()
	return max(dims[0], dims[1], dims[2])
}

func (husband Box) Union(wife Box) (child Box) {
	child = Box{
		Min: Vec3Min(husband.Min, wife.Min),
		Max: Vec3Max(husband.Max, wife.Max),
	}
	return
}

func (box Box) Center() mgl32.Vec3 {
	return box.Min.Add(box.Size().Mul(0.5))
}

func ClosestPointOnLine(lineStart, lineEnd, point mgl32.Vec3) mgl32.Vec3 {
	lineDir := lineEnd.Sub(lineStart)
	t := point.Sub(lineStart).Dot(lineDir) / lineDir.Dot(lineDir)
	return lineStart.Add(lineDir.Mul(Clamp(t, 0.0, 1.0)))
}
