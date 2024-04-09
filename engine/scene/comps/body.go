package comps

import (
	"log"

	"github.com/go-gl/mathgl/mgl32"
	"tophatdemon.com/total-invasion-ii/engine/math2/collision"
)

type HasBody interface {
	Body() *Body
}

type Body struct {
	Transform      Transform
	Velocity       mgl32.Vec3
	Shape          collision.Shape
	Filter         collision.Mask                // Determines which collision layers this body will respond to collisions with
	Layer          collision.Mask                // The collision layer(s) that this body resides on
	OnIntersect    func(*Body, collision.Result) // Called when the body intersects another body that passes the filter. Can be nil.
	SweepCollision bool                          // When true, the body has continous collision, which is more performance intensive but also more accurate.
}

func (b *Body) Body() *Body {
	return b
}

func (b *Body) Update(deltaTime float32) {
	b.Transform.Translate(b.Velocity[0]*deltaTime, b.Velocity[1]*deltaTime, b.Velocity[2]*deltaTime)
}

// Change the position of this body so that it doesn't collide with the other body.
func (b *Body) ResolveCollision(otherBody *Body) {
	if otherBody == nil || b == otherBody || (b.Filter&otherBody.Layer == 0 && b.OnIntersect == nil) {
		return
	}

	// Bounding box check
	if !b.Shape.Extents().Translate(b.Transform.Position()).Intersects(otherBody.Shape.Extents().Translate(otherBody.Transform.Position())) {
		return
	}

	var resolveFraction float32 = 0.0
	if b.Filter&otherBody.Layer != 0 {
		if b.Layer&otherBody.Filter != 0 {
			// When both bodies collide with each other, have both of them resolve it half way
			resolveFraction = 0.5
		} else {
			// Otherwise, the body that collides takes full precedence
			resolveFraction = 1.0
		}
	}

	if _, isSphere := b.Shape.(collision.Sphere); !isSphere {
		log.Printf("collision is not implemented for shape %v.\n", b.Shape)
	}

	res := b.Shape.ResolveCollision(b.Transform.Position(), otherBody.Transform.Position(), otherBody.Shape)
	if res.Hit {
		if b.Filter&otherBody.Layer != 0 {
			b.Transform.TranslateV(res.Normal.Mul(res.Penetration).Mul(resolveFraction))
		}
		if b.OnIntersect != nil {
			b.OnIntersect(otherBody, res)
		}
	}
}
