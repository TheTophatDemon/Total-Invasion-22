/*
Custom bindings to the CUTE_C2 library
*/
package c2

/*
#define CUTE_C2_IMPLEMENTATION
#include "./cute_c2.h"
*/
import "C"
import (
	"fmt"
	"strings"
	"unsafe"

	"github.com/go-gl/mathgl/mgl32"
)

type ShapeType C.int

const (
	TypeNone ShapeType = iota
	TypeCircle
	TypeAABB
	// Capsule is not supported
	TypePoly = 4
)

type (
	Circle struct {
		Pos    mgl32.Vec2
		Radius float32
	}
	AABB struct {
		Min, Max mgl32.Vec2
	}
	Poly struct {
		Count C.int
		Verts [MaxPolygonVerts]mgl32.Vec2
		Norms [MaxPolygonVerts]mgl32.Vec2
	}
	Rot struct {
		Cos, Sin float32
	}
	XForm struct {
		Pos mgl32.Vec2
		Rot Rot
	}
	TOIResult struct {
		Hit         bool
		Toi         float32
		Normal, Pos mgl32.Vec2
		Iterations  int
	}
	Manifold struct {
		Count         C.int
		Depths        [2]float32
		ContactPoints [2]mgl32.Vec2
		Normal        mgl32.Vec2
	}
	Raycast struct {
		Distance float32    // Time of impact
		Normal   mgl32.Vec2 // Normal of surface at impact (unit length)
	}
	Ray struct {
		Pos, Dir mgl32.Vec2
		Distance float32
	}
)

const (
	MaxPolygonVerts = 8
	MaxShapeSize    = max(unsafe.Sizeof(Circle{}), unsafe.Sizeof(AABB{}), unsafe.Sizeof(Poly{}))
)

func (c Circle) Type() ShapeType {
	return TypeCircle
}

func (a AABB) Type() ShapeType {
	return TypeAABB
}

func (p *Poly) Type() ShapeType {
	return TypePoly
}

func (res TOIResult) String() string {
	var sb strings.Builder
	sb.WriteString("{\n")
	fmt.Fprintf(&sb, "\thit: %t\n", res.Hit)
	fmt.Fprintf(&sb, "\tpos: (%v, %v)\n", res.Pos.X(), res.Pos.Y())
	fmt.Fprintf(&sb, "\tnormal: (%v, %v)\n", res.Normal.X(), res.Normal.Y())
	fmt.Fprintf(&sb, "\ttoi: %v\n", res.Toi)
	fmt.Fprintf(&sb, "\titerations: %v\n", res.Iterations)
	sb.WriteString("}")
	return sb.String()
}

func TOI(
	shapeA unsafe.Pointer, typeA ShapeType, transA *XForm, velA mgl32.Vec2,
	shapeB unsafe.Pointer, typeB ShapeType, transB *XForm, velB mgl32.Vec2,
	useRadius bool,
) TOIResult {
	if shapeA == nil || shapeB == nil {
		return TOIResult{}
	}
	cUseRadius := C.int(0)
	if useRadius {
		cUseRadius = C.int(1)
	}

	cResult := C.c2TOI(
		shapeA, C.C2_TYPE(typeA), (*C.c2x)(unsafe.Pointer(transA)), fromMglVec(velA),
		shapeB, C.C2_TYPE(typeB), (*C.c2x)(unsafe.Pointer(transB)), fromMglVec(velB),
		cUseRadius,
	)
	var result TOIResult
	result.Hit = (cResult.hit == C.int(1))
	result.Pos = toMglVec(cResult.p)
	result.Normal = toMglVec(cResult.n)
	result.Toi = float32(cResult.toi)
	result.Iterations = int(cResult.iterations)
	return result
}

func Inflate(shape unsafe.Pointer, shapeType ShapeType, skinFactor float32) {
	if shape == nil {
		return
	}
	C.c2Inflate(shape, C.C2_TYPE(shapeType), C.float(skinFactor))
}

func Collide(
	shapeA unsafe.Pointer, typeA ShapeType, transA *XForm,
	shapeB unsafe.Pointer, typeB ShapeType, transB *XForm,
) Manifold {
	manifold := Manifold{}
	C.c2Collide(shapeA,
		(*C.c2x)(unsafe.Pointer(transA)),
		C.C2_TYPE(typeA),
		shapeB,
		(*C.c2x)(unsafe.Pointer(transB)),
		C.C2_TYPE(typeB),
		(*C.c2Manifold)(unsafe.Pointer(&manifold)))
	return manifold
}

func Touches(
	shapeA unsafe.Pointer, typeA ShapeType, transA *XForm,
	shapeB unsafe.Pointer, typeB ShapeType, transB *XForm,
) bool {
	res := C.c2Collided(
		shapeA, (*C.c2x)(unsafe.Pointer(transA)), C.C2_TYPE(typeA),
		shapeB, (*C.c2x)(unsafe.Pointer(transB)), C.C2_TYPE(typeB),
	)
	return res == C.int(1)
}

func CastRay(ray Ray, shape unsafe.Pointer, shapeType ShapeType, shapeTrans *XForm) (bool, Raycast) {
	var cast Raycast
	hit := C.c2CastRay(convRay(ray), shape, (*C.c2x)(unsafe.Pointer(shapeTrans)), C.C2_TYPE(shapeType), (*C.c2Raycast)(unsafe.Pointer(&cast)))
	if hit == C.int(1) {
		return true, cast
	}
	return false, Raycast{}
}

func MakePoly(p *Poly) {
	if p == nil {
		return
	}

	C.c2MakePoly((*C.c2Poly)(unsafe.Pointer(p)))
}

func fromMglVec(vec mgl32.Vec2) C.c2v {
	return C.c2v{x: C.float(vec[0]), y: C.float(vec[1])}
}

func toMglVec(vec C.c2v) mgl32.Vec2 {
	return mgl32.Vec2{float32(vec.x), float32(vec.y)}
}

func convRay(ray Ray) C.c2Ray {
	return C.c2Ray{
		p: fromMglVec(ray.Pos),
		d: fromMglVec(ray.Dir),
		t: C.float(ray.Distance),
	}
}
