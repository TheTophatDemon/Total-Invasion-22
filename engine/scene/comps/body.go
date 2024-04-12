package comps

import (
	"log"

	"github.com/go-gl/mathgl/mgl32"
	"tophatdemon.com/total-invasion-ii/engine/math2/collision"
	"tophatdemon.com/total-invasion-ii/engine/scene"
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
	LockY          bool                          // When true, the body will not move on the Y axis in response to collisions.
}

func (body *Body) Body() *Body {
	return body
}

func (body *Body) MoveAndCollide(deltaTime float32, bodiesIterator func() (HasBody, scene.Handle)) {
	before := body.Transform.Position()

	movement := body.Velocity.Mul(deltaTime)
	if body.Velocity.LenSqr() != 0.0 {
		for collidingEnt, _ := bodiesIterator(); collidingEnt != nil; collidingEnt, _ = bodiesIterator() {
			if collidingEnt.Body() != body {
				body.ResolveCollision(movement, collidingEnt.Body())
			}
		}
	}

	body.Transform.TranslateV(movement)

	if body.LockY {
		// Restrict movement to the XZ plane
		after := body.Transform.Position()
		body.Transform.SetPosition(mgl32.Vec3{after.X(), before.Y(), after.Z()})
	}
}

// Change the position of this body so that it doesn't collide with the other body.
func (body *Body) ResolveCollision(movement mgl32.Vec3, otherBody *Body) {
	if otherBody == nil || body == otherBody || (body.Filter&otherBody.Layer == 0 && body.OnIntersect == nil) {
		return
	}

	nextPosition := body.Transform.Position().Add(movement)

	// Bounding box check
	bbox := body.Shape.Extents().Translate(nextPosition)
	if body.SweepCollision {
		bbox = bbox.Union(body.Shape.Extents().Translate(body.Transform.Position()))
	}
	if !bbox.Intersects(otherBody.Shape.Extents().Translate(otherBody.Transform.Position())) {
		return
	}

	var resolveFraction float32 = 0.0
	if body.Filter&otherBody.Layer != 0 {
		if body.Layer&otherBody.Filter != 0 {
			// When both bodies collide with each other, have both of them resolve it half way
			resolveFraction = 0.5
		} else {
			// Otherwise, the body that collides takes full precedence
			resolveFraction = 1.0
		}
	}

	if _, isSphere := body.Shape.(collision.Sphere); !isSphere {
		log.Printf("collision is not implemented for shape %v.\n", body.Shape)
	}

	var shapePosition mgl32.Vec3 = nextPosition
	if body.SweepCollision {
		movementLength := movement.Len()
		if movementLength != 0.0 {
			movementDir := movement.Mul(1.0 / movementLength)
			rayHit := otherBody.Shape.Raycast(body.Transform.Position(), movementDir, otherBody.Transform.Position(), movementLength)
			if rayHit.Hit {
				shapePosition = rayHit.Position
			}
		}
	}

	res := body.Shape.ResolveCollision(shapePosition, otherBody.Transform.Position(), otherBody.Shape)
	if res.Hit {
		if body.Filter&otherBody.Layer != 0 {
			body.Transform.TranslateV(res.Normal.Mul(res.Penetration).Mul(resolveFraction))
		}
		if body.OnIntersect != nil {
			body.OnIntersect(otherBody, res)
		}
	}
}
