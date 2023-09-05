package comps

import (
	"log"

	"github.com/go-gl/mathgl/mgl32"
	"tophatdemon.com/total-invasion-ii/engine/math2"
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

type HasBody interface {
	Body() *Body
}

type Body struct {
	Transform Transform
	Velocity  mgl32.Vec3
	Shape     CollisionShape
	Extents   mgl32.Vec3       // Describes the half-size of the shape on each axis relative to its origin.
	Triangles []math2.Triangle // Refers to the triangles in a mesh collision shape.
	Pushiness int              // When two bodies collide, the one with higher Pushiness will not be moved.
	NoClip    bool             // Allows the body to pass through all other bodies.
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
	if !math2.BoxIntersect(b.Transform.Position(), b.Extents, otherBody.Transform.Position(), otherBody.Extents) {
		return
	}

	// Resolve based on shape
	if b.Shape == COL_SHAPE_SPHERE {
		switch otherBody.Shape {
		case COL_SHAPE_SPHERE:
			b.ResolveCollisionSphere(otherBody.Transform.Position(), otherBody.Extents[0])
		case COL_SHAPE_MESH:
			b.ResolveCollisionTriangles(otherBody.Transform.Position(), otherBody.Triangles)
		default:
			log.Printf("collision is not implemented between shapes %s and %s.\n", b.Shape, otherBody.Shape)
		}
	} else {
		log.Printf("collision is not implemented for shape %s.\n", b.Shape)
	}
}

func (b *Body) ResolveCollisionSphere(spherePos mgl32.Vec3, sphereRadius float32) {
	diff := b.Transform.Position().Sub(spherePos)
	dist := diff.Len()
	// Resolve sphere vs. sphere
	if dist < b.Extents[0]+sphereRadius && dist != 0.0 {
		b.Transform.TranslateV(diff.Normalize().Mul(b.Extents[0] + sphereRadius - dist))
	}
}

func (b *Body) ResolveCollisionTriangles(trianglesOffset mgl32.Vec3, triangles []math2.Triangle) {
	panic("not implemented")
}
