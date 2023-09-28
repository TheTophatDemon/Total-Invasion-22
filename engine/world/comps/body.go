package comps

import (
	"fmt"
	"log"

	"github.com/go-gl/mathgl/mgl32"
	"tophatdemon.com/total-invasion-ii/engine/assets"
	"tophatdemon.com/total-invasion-ii/engine/math2"
	"tophatdemon.com/total-invasion-ii/engine/math2/collision"
)

type CollisionShape uint8

const (
	COL_SHAPE_SPHERE CollisionShape = iota
	COL_SHAPE_MESH
)

func (cs CollisionShape) String() string {
	switch cs {
	case COL_SHAPE_SPHERE:
		return "Sphere"
	case COL_SHAPE_MESH:
		return "Mesh"
	}
	return "Unknown"
}

var (
	errBodyNotSphere = fmt.Errorf("body should be a sphere")
)

type HasBody interface {
	Body() *Body
}

type Body struct {
	Transform Transform
	Velocity  mgl32.Vec3
	Shape     collision.Shape
	Pushiness int  // When two bodies collide, the one with higher Pushiness will not be moved.
	NoClip    bool // Allows the body to pass through all other bodies.
}

func (b *Body) Body() *Body {
	return b
}

func (b *Body) Update(deltaTime float32) {
	b.Transform.Translate(b.Velocity[0]*deltaTime, b.Velocity[1]*deltaTime, b.Velocity[2]*deltaTime)
}

// Change the position of this body so that it doesn't collide with the other body.
func (b *Body) ResolveCollision(otherBody *Body) {
	if otherBody == nil || b == otherBody || b.Pushiness > otherBody.Pushiness || b.NoClip || otherBody.NoClip {
		return
	}

	// Bounding box check
	if !b.Shape.Extents().Translate(b.Transform.Position()).Intersects(otherBody.Shape.Extents().Translate(otherBody.Transform.Position())) {
		return
	}

	// Resolve based on shape
	if _, ok := b.Shape.Radius(); ok {
		radius, isSphere := otherBody.Shape.Radius()
		mesh, isMesh := otherBody.Shape.Mesh()
		isBox := (otherBody.Shape.Kind() == collision.SHAPE_KIND_BOX)
		switch {
		case isSphere:
			b.ResolveCollisionSphereSphere(otherBody.Transform.Position(), radius)
		case isMesh:
			b.ResolveCollisionSphereTriangles(otherBody.Transform.Position(), mesh, nil, collision.TRIHIT_ALL)
		case isBox:
			b.ResolveCollisionSphereBox(otherBody.Transform.Position(), otherBody.Shape.Extents(), true)
		default:
			log.Printf("collision is not implemented between shapes %v and %v.\n", b.Shape, otherBody.Shape)
		}
	} else {
		log.Printf("collision is not implemented for shape %v.\n", b.Shape)
	}
}

func (b *Body) ResolveCollisionSphereSphere(spherePos mgl32.Vec3, otherRadius float32) error {
	radius, isSphere := b.Shape.Radius()
	if !isSphere {
		return errBodyNotSphere
	}
	diff := b.Transform.Position().Sub(spherePos)
	dist := diff.Len()
	// Resolve sphere vs. sphere
	if dist < radius+otherRadius && dist != 0.0 {
		b.Transform.TranslateV(diff.Normalize().Mul(radius + otherRadius - dist))
	}

	return nil
}

func (b *Body) ResolveCollisionSphereBox(boxOffset mgl32.Vec3, box math2.Box, hitCorners bool) error {
	radius, isSphere := b.Shape.Radius()
	if !isSphere {
		return errBodyNotSphere
	}
	pos := b.Transform.Position()

	box = box.Translate(boxOffset)
	projectedPoint := math2.Vec3Max(math2.Vec3Min(pos, box.Max), box.Min)
	diff := pos.Sub(projectedPoint)
	distSq := diff.LenSqr()

	// Ignore corner collisions if specified
	if !hitCorners && math2.Abs(diff.X()) > mgl32.Epsilon && math2.Abs(diff.Y()) > mgl32.Epsilon {
		return nil
	}

	if distSq > 0.0 && distSq < radius*radius {
		// When the sphere's center touches the edge of the box, push in the direction of the edge.
		b.Transform.TranslateV(diff.Normalize().Mul(radius - math2.Sqrt(distSq)))
	}

	return nil
}

func (b *Body) ResolveCollisionSphereTriangles(trianglesOffset mgl32.Vec3, mesh *assets.Mesh, triangleIndices []int, filter collision.TriangleHit) error {
	radius, isSphere := b.Shape.Radius()
	if !isSphere {
		return errBodyNotSphere
	}
	if mesh == nil {
		return fmt.Errorf("mesh cannot be nil")
	}
	if filter == collision.TRIHIT_NONE {
		return nil
	}

	var tc int
	if triangleIndices != nil {
		tc = len(triangleIndices)
	} else {
		tc = len(mesh.Triangles())
	}
	for tt := 0; tt < tc; tt += 1 {
		var t int
		if triangleIndices != nil {
			t = triangleIndices[tt]
		} else {
			t = tt
		}
		triangle := mesh.Triangles()[t]
		// Add offset
		for i := 0; i < len(triangle); i += 1 {
			triangle[i] = triangle[i].Add(trianglesOffset)
		}

		hit, col := collision.SphereTriangleCollision(b.Transform.Position(), radius, triangle)
		if int(hit)&int(filter) > 0 {
			b.Transform.TranslateV(col.Normal.Mul(col.Penetration + mgl32.Epsilon))
		}
	}
	return nil
}
