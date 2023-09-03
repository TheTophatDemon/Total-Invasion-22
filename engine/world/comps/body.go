package comps

import (
	"github.com/go-gl/mathgl/mgl32"
)

type CollisionShape uint8

const (
	COL_SHAPE_SPHERE CollisionShape = iota
	COL_SHAPE_BOX
)

type HasBody interface {
	Body() *Body
}

type Body struct {
	Velocity   mgl32.Vec3
	Shape      CollisionShape
	Extents    mgl32.Vec3 // Describes the size of the collision shape. For spheres, the X axis is the radius. For boxes, it describes the half-dimensions.
	Unpushable bool       // Prevents the body from moving in response to other colliding bodies.
	NoClip     bool       // Allows the body to pass through all other bodies.
}

func (b *Body) Body() *Body {
	return b
}

func (b *Body) Update(transform *Transform, deltaTime float32) {
	transform.Translate(b.Velocity[0]*deltaTime, b.Velocity[1]*deltaTime, b.Velocity[2]*deltaTime)
}

// Change the velocity of this body so that it doesn't collide with the other body.
func (b *Body) ResolveCollision(myPos, otherPos mgl32.Vec3, otherBody *Body, deltaTime float32) {
	if otherBody == nil || b == otherBody || b.Unpushable || b.NoClip || otherBody.NoClip {
		return
	}
	myNewPos := myPos.Add(b.Velocity.Mul(deltaTime))
	otherNewPos := otherPos.Add(otherBody.Velocity.Mul(deltaTime))
	diff := myNewPos.Sub(otherNewPos)
	dist := diff.Len()
	if b.Shape == COL_SHAPE_SPHERE && otherBody.Shape == COL_SHAPE_SPHERE {
		// Resolve sphere vs. sphere
		if dist < b.Extents[0]+otherBody.Extents[0] && dist != 0.0 {
			b.Velocity = b.Velocity.Add(diff.Normalize().Mul(b.Extents[0] + otherBody.Extents[0] - dist))
		}
	} else if b.Shape == COL_SHAPE_BOX && otherBody.Shape == COL_SHAPE_SPHERE {
		// Resolve box vs. sphere
		panic("not implemented")
	} else if b.Shape == COL_SHAPE_SPHERE && otherBody.Shape == COL_SHAPE_BOX {
		// Resolve sphere vs. box
		panic("not implemented")
	} else if b.Shape == COL_SHAPE_BOX && otherBody.Shape == COL_SHAPE_BOX {
		// Resolve box vs. box
		panic("not implemented")
	}
}
