package collision

import (
	"fmt"
	"unsafe"

	"github.com/go-gl/mathgl/mgl32"
	"tophatdemon.com/total-invasion-ii/engine/assets/geom"
	"tophatdemon.com/total-invasion-ii/engine/math2"
	"tophatdemon.com/total-invasion-ii/engine/math2/collision/c2"
)

type ShapeType c2.ShapeType

const (
	ShapeNone ShapeType = iota
	ShapeCylinder
	ShapeBox
	ShapeConvex = 4
)

// A generic collision shape representing a 2D shape on the XZ axis extended upwards.
// Combines all of the C2 shape types into a single data structure and supplies a height variable.
type Shape struct {
	extents math2.Box // Represents the calculated bounding box for the shape. Will also contain the circle's radius.
	type2d  c2.ShapeType
	poly    c2.Poly
}

func (sh Shape) String() string {
	switch sh.type2d {
	case c2.TypeNone:
		return "None"
	case c2.TypeCircle:
		return "Cylinder"
	case c2.TypeAABB:
		return "Box"
	case c2.TypePoly:
		return "Convex"
	default:
		return "Unknown"
	}
}

func (sh Shape) Type() ShapeType {
	return ShapeType(sh.type2d)
}

func (sh Shape) Extents() math2.Box {
	return sh.extents
}

func (sh Shape) Radius() float32 {
	return sh.extents.Max[0]
}

func NewCylinder(radius, height float32) Shape {
	return Shape{
		type2d:  c2.TypeCircle,
		extents: math2.BoxFromRadius(radius),
	}
}

func NewBox(extents math2.Box) Shape {
	return Shape{
		type2d:  c2.TypeAABB,
		extents: extents,
	}
}

func NewConvex(mesh *geom.Mesh) Shape {
	shape := Shape{
		type2d: c2.TypePoly,
	}
	if mesh == nil {
		return shape
	}

	shape.extents = mesh.BoundingBox()

	// TODO: Smush all of the vertices into a convex hull

	return shape
}

func (shape Shape) Inflate(amount float32) Shape {
	switch shape.type2d {
	case c2.TypeCircle, c2.TypeAABB:
		shape.extents.Max = shape.extents.Max.Add(mgl32.Vec3{amount, amount, amount})
		shape.extents.Min = shape.extents.Min.Sub(mgl32.Vec3{amount, amount, amount})
	case c2.TypePoly:
		polyPtr := unsafe.Pointer(&shape.poly)
		c2.Inflate(polyPtr, shape.type2d, amount)
	}
	return shape
}

func (shape Shape) Touches(myPosition, theirPosition mgl32.Vec3, theirShape Shape) bool {
	aPtr, aTrans := shape.c2Ptr(myPosition)
	bPtr, bTrans := theirShape.c2Ptr(theirPosition)
	return c2.Touches(aPtr, shape.type2d, &aTrans, bPtr, theirShape.type2d, &bTrans) &&
		theirShape.extents.Max[1] >= shape.extents.Min[1] &&
		theirShape.extents.Min[1] < shape.extents.Max[1]
}

func (shape Shape) Raycast(myPosition, rayOrigin, rayDir mgl32.Vec3, maxDist float32) Result {
	shapePtr, shapeTrans := shape.c2Ptr(myPosition)

	hit, c2Result := c2.CastRay(c2.Ray{
		Pos:      mgl32.Vec2{rayOrigin[0], rayOrigin[2]},
		Dir:      mgl32.Vec2{rayDir[0], rayDir[2]},
		Distance: maxDist,
	}, shapePtr, shape.type2d, &shapeTrans)

	if hit {
		hitCrossSectionAt := rayOrigin.Add(rayDir.Mul(c2Result.Distance))
		if hitCrossSectionAt[1] >= shape.extents.Min[1] && hitCrossSectionAt[1] < shape.extents.Max[1] {
			return Result{
				Hit:      true,
				Position: hitCrossSectionAt,
				Normal:   mgl32.Vec3{c2Result.Normal[0], 0.0, c2Result.Normal[1]},
				Distance: c2Result.Distance,
			}
		}
		// TODO: Check for upper and lower plane
	}
	return Result{
		Distance: maxDist,
	}
}

func (shape Shape) Shapecast(myPosition, theirPosition, theirMovement mgl32.Vec3, theirShape Shape) Result {
	maxLen := theirMovement.Len()
	var planeT float32 = -math2.Inf32()
	var planePos mgl32.Vec3
	yDir := theirMovement[1] / maxLen
	if theirPosition[1]+theirShape.extents.Min[1] > myPosition[1]+shape.extents.Max[1] {
		// They start from above us
		if theirMovement[1] >= 0.0 {
			fmt.Println("A")
			// They are moving away vertically
			return Result{Distance: maxLen}
		}
		// They are moving towards the top plane. Find the time of intersection
		planeT = (myPosition[1] + shape.extents.Max[1] - (theirPosition[1] + theirShape.extents.Min[1])) / yDir
		planePos = theirPosition.Add(theirMovement.Mul(planeT / maxLen))
	} else if theirPosition[1]+theirShape.extents.Max[1] < myPosition[1]+shape.extents.Min[1] {
		// They start from below us
		if theirMovement[1] <= 0.0 {
			fmt.Println("B")
			// They are moving away vertically
			return Result{Distance: maxLen}
		}
		// They are moving towards the bottom plane. Find the time of intersection
		planeT = (myPosition[1] + shape.extents.Min[1] - (theirPosition[1] + theirShape.extents.Max[1])) / yDir
		planePos = theirPosition.Add(theirMovement.Mul(planeT / maxLen))
	}

	aPtr, aTrans := shape.c2Ptr(myPosition)
	bPtr, bTrans := theirShape.c2Ptr(theirPosition)
	sweepResult := c2.TOI(
		bPtr, theirShape.type2d, &bTrans, mgl32.Vec2{theirMovement[0], theirMovement[2]},
		aPtr, shape.type2d, &aTrans, mgl32.Vec2{}, true,
	)
	if sweepResult.Hit {
		sweepDist := sweepResult.Toi * maxLen
		if sweepDist < planeT {
			fmt.Println("C")
			// They hit the cross section and then one of the caps.
			if shape.Touches(myPosition, planePos, theirShape) {
				return Result{
					Hit:      true,
					Position: planePos,
					Normal:   mgl32.Vec3{0.0, -math2.Signum(yDir), 0.0},
					Distance: planeT,
				}
			}
		} else {
			fmt.Println("D")
			// They passed through the top or bottom plane and hit the cross section
			sweepPos := theirPosition.Add(theirMovement.Mul(sweepDist / maxLen))
			intersectsY := sweepPos[1]+theirShape.extents.Min[1] <= myPosition[1]+shape.extents.Max[1] &&
				sweepPos[1]+theirShape.extents.Max[1] >= myPosition[1]+shape.extents.Min[1]
			if intersectsY {
				return Result{
					Hit:      true,
					Position: sweepPos,
					Normal:   mgl32.Vec3{sweepResult.Normal[0], 0.0, sweepResult.Normal[1]},
					Distance: sweepDist,
				}
			}
		}
	}
	fmt.Println("E")
	return Result{
		Distance: maxLen,
	}
}

// Returns the vector needed to push theirShape (roughly) out of this shape when they are colliding.
func (shape Shape) PushOut(myPosition, theirPosition mgl32.Vec3, theirShape Shape) mgl32.Vec3 {
	intersectsY := myPosition[1]+shape.extents.Max[1] >= theirPosition[1]+theirShape.extents.Min[1] &&
		myPosition[1]+shape.extents.Min[1] < theirPosition[1]+theirShape.extents.Max[1]
	if !intersectsY {
		return mgl32.Vec3{}
	}

	aPtr, aTrans := shape.c2Ptr(myPosition)
	bPtr, bTrans := theirShape.c2Ptr(theirPosition)
	res := c2.Collide(aPtr, shape.type2d, &aTrans, bPtr, theirShape.type2d, &bTrans)
	if res.Count > 0 {
		avg := res.Depths[0]
		if res.Count > 1 {
			avg += res.Depths[1]
			avg /= 2.0
		}
		pushVec := res.Normal.Mul(-0.25 * avg)
		return mgl32.Vec3{pushVec[0], 0.0, pushVec[1]}
	}
	return mgl32.Vec3{}
}

func (shape Shape) IsNil() bool {
	return shape.type2d == c2.TypeNone
}

func (shape Shape) c2Ptr(position mgl32.Vec3) (unsafe.Pointer, c2.XForm) {
	switch shape.type2d {
	case c2.TypeCircle:
		circle := c2.Circle{
			Pos:    mgl32.Vec2{position[0], position[2]},
			Radius: shape.extents.Max[0],
		}
		return unsafe.Pointer(&circle), c2.XForm{}
	case c2.TypeAABB:
		aabb := c2.AABB{
			Min: mgl32.Vec2{shape.extents.Min[0] + position[0], shape.extents.Min[2] + position[2]},
			Max: mgl32.Vec2{shape.extents.Max[0] + position[0], shape.extents.Max[2] + position[2]},
		}
		return unsafe.Pointer(&aabb), c2.XForm{}
	case c2.TypePoly:
		return unsafe.Pointer(&shape.poly), c2.XForm{
			Pos: mgl32.Vec2{position[0], position[2]},
			Rot: c2.Rot{Cos: 1.0},
		}
	}
	return nil, c2.XForm{}
}
