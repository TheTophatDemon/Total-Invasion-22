package collision

import (
	"math"
	"slices"

	"github.com/go-gl/mathgl/mgl32"
	"tophatdemon.com/total-invasion-ii/engine/assets/geom"
	"tophatdemon.com/total-invasion-ii/engine/failure"
	"tophatdemon.com/total-invasion-ii/engine/math2"
)

const MaxPointCount = 8

// A generic collision shape representing a 2D shape on the XZ axis extended upwards.
type (
	Shape struct {
		extents    math2.Box // Represents the calculated bounding box for the shape, relative to the center. Also represents the shape's height.
		points     [MaxPointCount]mgl32.Vec2
		pointCount int
	}
	SegmentIter struct {
		Shape         Shape
		PointIndex    int
		ShapePosition mgl32.Vec2
	}
	Segment struct {
		Points [2]mgl32.Vec2
		Normal mgl32.Vec2
	}
)

func NewShape(height float32, points ...mgl32.Vec2) Shape {
	var shape Shape

	if len(points) > MaxPointCount {
		failure.LogErrWithLocation("warning: shape has %v points but the maximum number is %v; it will be truncated", len(points), MaxPointCount)
	}
	shape.pointCount = min(len(points), MaxPointCount)
	copy(shape.points[:shape.pointCount], points)

	// Calculate bounding box over points
	shape.extents.Min[1] = -height / 2.0
	shape.extents.Max[1] = height / 2.0
	shape.extents.Max[0] = -math.MaxFloat32
	shape.extents.Max[2] = -math.MaxFloat32
	shape.extents.Min[0] = math.MaxFloat32
	shape.extents.Min[2] = math.MaxFloat32

	for _, point := range shape.points[:shape.pointCount] {
		shape.extents.Max[0] = max(shape.extents.Max[0], point[0])
		shape.extents.Max[2] = max(shape.extents.Max[2], point[1])
		shape.extents.Min[0] = min(shape.extents.Min[0], point[0])
		shape.extents.Min[2] = min(shape.extents.Min[2], point[1])
	}

	return shape
}

func NewBoxShape(halfExtentX, halfExtentY, halfExtentZ float32) Shape {
	return NewShape(halfExtentY*2.0, []mgl32.Vec2{
		{-halfExtentX, -halfExtentZ},
		{+halfExtentX, -halfExtentZ},
		{+halfExtentX, +halfExtentZ},
		{-halfExtentX, +halfExtentZ},
	}...)
}

func NewShapeHull(height float32, points ...mgl32.Vec2) Shape {
	if len(points) <= 2 {
		return Shape{}
	}

	// Andrew's algorithm
	// Real Time Collision Detection 3.9.1

	// Sort points from left to right
	slices.SortFunc(points, func(a, b mgl32.Vec2) int {
		if a[0] > b[0] {
			return 1
		} else if a[0] < b[0] {
			return -1
		} else if a[1] > b[1] {
			return 1
		} else if a[1] < b[1] {
			return -1
		}
		return 0
	})

	// Construct upper hull
	hullPoints := make([]mgl32.Vec2, 2, len(points))
	hullPoints[0] = points[0]
	hullPoints[1] = points[1]
	for _, point := range points[2:] {
		for back := range len(hullPoints) {
			segIndex := len(hullPoints) - back - 2
			if segIndex < 0 {
				hullPoints = hullPoints[:1]
				hullPoints = append(hullPoints, point)
				break
			}
			segStart := hullPoints[segIndex]
			segEnd := hullPoints[segIndex+1]
			if segStart.ApproxEqual(segEnd) || segStart.ApproxEqual(point) {
				// Segment degenerates into point. Ignore!
				continue
			}
			previousSegment := segEnd.Sub(segStart)
			dp := point.Sub(segStart).Dot(mgl32.Vec2{previousSegment[1], -previousSegment[0]})
			if dp <= 0.0 && !point.ApproxEqual(segEnd) {
				hullPoints = hullPoints[:segIndex+2]
				hullPoints = append(hullPoints, point)
				break
			}
		}
	}

	lowerStartIdx := len(hullPoints)

	// Construct lower hull
	for p := len(points) - 1; p >= 0; p -= 1 {
		point := points[p]
		for back := range len(hullPoints) - lowerStartIdx + 1 {
			segIndex := len(hullPoints) - back - 2
			segStart := hullPoints[segIndex]
			segEnd := hullPoints[segIndex+1]
			if segStart.ApproxEqual(segEnd) || segStart.ApproxEqual(point) {
				// Segment degenerates into point. Ignore!
				continue
			}
			previousSegment := segEnd.Sub(segStart)
			dp := point.Sub(segStart).Dot(mgl32.Vec2{previousSegment[1], -previousSegment[0]})
			if point.ApproxEqual(segEnd) {
				break
			}
			if dp <= 0.0 {
				hullPoints = hullPoints[:segIndex+2]
				hullPoints = append(hullPoints, point)
				break
			}
		}
	}
	// Remove redundant point at the end
	hullPoints = hullPoints[:len(hullPoints)-1]

	return NewShape(height, hullPoints...)
}

func NewShapeFromMesh(mesh *geom.Mesh, matrix mgl32.Mat4) Shape {
	if mesh == nil {
		return Shape{}
	}

	// Smush all of the vertices into a convex hull
	hullVerts := make([]mgl32.Vec2, 0, len(mesh.Verts().Pos))
	minY := float32(math.MaxFloat32)
	maxY := -minY

meshLoop:
	for _, pos := range mesh.Verts().Pos {
		pos = mgl32.TransformCoordinate(pos, matrix)
		minY = min(minY, pos[1])
		maxY = max(maxY, pos[1])
		pos2d := mgl32.Vec2{pos[0], pos[2]}

		// Skip over duplicate vertices
		for _, vert := range hullVerts {
			if vert.ApproxEqual(pos2d) {
				continue meshLoop
			}
		}

		hullVerts = append(hullVerts, pos2d)
	}

	return NewShapeHull(maxY-minY, hullVerts...)
}

func (shape Shape) Extents() math2.Box {
	return shape.extents
}

func (shape Shape) Radius() float32 {
	return shape.extents.LongestDimension() / 2.0
}

func (shape Shape) Points() []mgl32.Vec2 {
	return shape.points[:shape.pointCount]
}

func (shape Shape) IsNil() bool {
	return shape.pointCount == 0
}

func (shape Shape) Segments(myPosition mgl32.Vec2) SegmentIter {
	return SegmentIter{Shape: shape, ShapePosition: myPosition}
}

func (iter *SegmentIter) Next() (Segment, bool) {
	if iter == nil || iter.PointIndex >= iter.Shape.pointCount {
		return Segment{}, false
	}

	pointA := iter.Shape.points[iter.PointIndex].Add(iter.ShapePosition)
	pointB := iter.Shape.points[(iter.PointIndex+1)%iter.Shape.pointCount].Add(iter.ShapePosition)

	iter.PointIndex += 1

	if pointA.ApproxEqual(pointB) {
		// Degenerated into a point, somethin' ain't right here.
		return Segment{}, false
	}

	return Segment{
		Points: [2]mgl32.Vec2{pointA, pointB},
		Normal: mgl32.Vec2{pointB[1] - pointA[1], pointA[0] - pointB[0]}.Normalize(),
	}, true
}

func (shape Shape) ContainsPoint(myPosition, testPoint mgl32.Vec2) bool {
	segIter := shape.Segments(myPosition)
	for seg, ok := segIter.Next(); ok; seg, ok = segIter.Next() {
		if testPoint.Sub(seg.Points[0]).Dot(seg.Normal) > 0.0 {
			return false
		}
	}
	return true
}

func (shape Shape) Touches(myPosition, theirPosition mgl32.Vec3, theirShape Shape) bool {
	// Check vertical intersection
	if theirPosition[1]+theirShape.extents.Max[1] < myPosition[1]+shape.extents.Min[1] {
		return false
	}
	if theirPosition[1]+theirShape.extents.Min[1] > myPosition[1]+shape.extents.Max[1] {
		return false
	}

	myPos2D := mgl32.Vec2{myPosition[0], myPosition[2]}
	theirPos2D := mgl32.Vec2{theirPosition[0], theirPosition[2]}

	for _, point := range theirShape.Points() {
		point = point.Add(theirPos2D)
		if shape.ContainsPoint(myPos2D, point) {
			return true
		}
	}

	for _, point := range shape.Points() {
		point = point.Add(myPos2D)
		if theirShape.ContainsPoint(theirPos2D, point) {
			return true
		}
	}

	return false
}

func (shape Shape) Raycast(myPosition, rayOrigin, rayDir mgl32.Vec3, maxDist float32) (result Result) {
	// Real Time Collision Detection 5.3.8
	result.Distance = maxDist

	tFirst := float32(0.0)
	tFirstNormal := mgl32.Vec3{}
	tLast := float32(1.0)

	var planesArr [MaxPointCount + 2]math2.Plane
	planes := planesArr[0:0]

	segs := shape.Segments(mgl32.Vec2{myPosition[0], myPosition[2]})
	for seg, ok := segs.Next(); ok; seg, ok = segs.Next() {
		planes = append(planes, math2.PlaneFromPointAndNormal(
			mgl32.Vec3{seg.Points[0][0], myPosition[1], seg.Points[0][1]},
			mgl32.Vec3{seg.Normal[0], 0.0, seg.Normal[1]},
		))
	}
	planes = append(planes,
		math2.PlaneFromPointAndNormal(mgl32.Vec3{0.0, myPosition[1] + shape.extents.Max[1], 0.0}, mgl32.Vec3{0.0, 1.0, 0.0}))
	planes = append(planes,
		math2.PlaneFromPointAndNormal(mgl32.Vec3{0.0, myPosition[1] + shape.extents.Min[1], 0.0}, mgl32.Vec3{0.0, -1.0, 0.0}))

	for _, plane := range planes {
		denom := plane.Normal.Dot(rayDir.Mul(maxDist))
		dist := -plane.SignedDistance(rayOrigin)
		if denom == 0.0 {
			// Ray is a parallel to this plane, abort if it lies outside
			if dist < 0.0 {
				return
			}
		} else {
			t := dist / denom
			if denom < 0.0 {
				// When entering halfspace, update tfirst if t is larger
				if t > tFirst {
					tFirst = t
					tFirstNormal = plane.Normal
				}
			} else {
				tLast = min(tLast, t)
			}
			if tFirst > tLast {
				// When the interval becomes empty there is no intersection
				return
			}
		}
	}

	result.Hit = true
	result.Distance = tFirst
	result.Normal = tFirstNormal
	result.Position = rayOrigin.Add(rayDir.Mul(tFirst))
	return
}

func (shape Shape) Shapecast(myPosition, theirPosition, theirMovement mgl32.Vec3, theirShape Shape) Result {
	//TODO
	return Result{}
}

// Returns the vector needed to push this shape out of theirShape when they are colliding.
func (shape Shape) PushOutOf(myPosition, theirPosition mgl32.Vec3, theirShape Shape) (hit bool, pushOut mgl32.Vec3) {
	// Check vertical intersection
	var verticalInterval float32
	if interval := (theirPosition[1] + theirShape.extents.Max[1]) - (myPosition[1] + shape.extents.Min[1]); interval < 0.0 {
		return false, mgl32.Vec3{}
	} else {
		verticalInterval = interval
	}
	if interval := (theirPosition[1] + theirShape.extents.Min[1]) - (myPosition[1] + shape.extents.Max[1]); interval > 0.0 {
		return false, mgl32.Vec3{}
	} else if -interval < verticalInterval {
		verticalInterval = interval
	}

	var smallestInterval float32 = math.MaxFloat32
	var closestNormal mgl32.Vec2

	myPos2d := mgl32.Vec2{myPosition[0], myPosition[2]}
	theirPos2d := mgl32.Vec2{theirPosition[0], theirPosition[2]}

	for _, iter := range [...]SegmentIter{shape.Segments(myPos2d), theirShape.Segments(theirPos2d)} {
		for seg, ok := iter.Next(); ok; seg, ok = iter.Next() {
			var aMin float32 = math.MaxFloat32
			var aMax float32 = -math.MaxFloat32
			for _, point := range theirShape.Points() {
				proj := point.Add(theirPos2d).Dot(seg.Normal)
				aMin = min(aMin, proj)
				aMax = max(aMax, proj)
			}

			var bMin float32 = math.MaxFloat32
			var bMax float32 = -math.MaxFloat32
			for _, point := range shape.Points() {
				proj := point.Add(myPos2d).Dot(seg.Normal)
				bMin = min(bMin, proj)
				bMax = max(bMax, proj)
			}

			if bMin < aMax && bMax >= aMin {
				var interval float32
				if bMax >= aMax {
					interval = aMax - bMin
				} else {
					interval = aMin - bMax
				}
				if math.Abs(float64(interval)) < math.Abs(float64(smallestInterval)) {
					smallestInterval = interval
					closestNormal = seg.Normal
				}
				continue
			}
			// Separating axis found
			return
		}
	}

	hit = true
	if math2.Abs(smallestInterval) < math2.Abs(verticalInterval) {
		pushOut2d := closestNormal.Mul(smallestInterval)
		pushOut = mgl32.Vec3{pushOut2d[0], 0.0, pushOut2d[1]}
	} else {
		pushOut = mgl32.Vec3{0.0, verticalInterval, 0.0}
	}
	return
}
