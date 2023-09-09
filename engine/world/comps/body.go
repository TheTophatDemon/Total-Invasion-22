package comps

import (
	"log"

	"github.com/go-gl/mathgl/mgl32"
	"tophatdemon.com/total-invasion-ii/engine/assets"
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
	Transform     Transform
	Velocity      mgl32.Vec3
	Shape         CollisionShape
	Radius        float32      // The radius of the sphere collision shape.
	Extents       math2.Box    // Describes the body's bounding box, centered at its origin.
	CollisionMesh *assets.Mesh // Refers to the mesh used for the mesh collision shape.
	Pushiness     int          // When two bodies collide, the one with higher Pushiness will not be moved.
	NoClip        bool         // Allows the body to pass through all other bodies.
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
	if !b.Extents.Translate(b.Transform.Position()).Intersects(otherBody.Extents.Translate(otherBody.Transform.Position())) {
		return
	}

	// Resolve based on shape
	if b.Shape == COL_SHAPE_SPHERE {
		switch otherBody.Shape {
		case COL_SHAPE_SPHERE:
			b.ResolveCollisionSphere(otherBody.Transform.Position(), otherBody.Radius)
		case COL_SHAPE_MESH:
			b.ResolveCollisionTriangles(otherBody.Transform.Position(), otherBody.CollisionMesh, nil, math2.TRIHIT_ALL)
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
	if dist < b.Radius+sphereRadius && dist != 0.0 {
		b.Transform.TranslateV(diff.Normalize().Mul(b.Radius + sphereRadius - dist))
	}
}

func (b *Body) ResolveCollisionTriangles(trianglesOffset mgl32.Vec3, mesh *assets.Mesh, triangleIndices []int, filter math2.TriangleHit) {
	if filter == math2.TRIHIT_NONE || mesh == nil {
		log.Println("Warning: Invalid parameter fed to ResolveCollisionTriangles.")
		return
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

		hit, col := math2.SphereTriangleCollision(b.Transform.Position(), b.Radius, triangle)
		if int(hit)&int(filter) > 0 {
			b.Transform.TranslateV(col.Normal.Mul(col.Penetration + mgl32.Epsilon))
		}
	}
}
