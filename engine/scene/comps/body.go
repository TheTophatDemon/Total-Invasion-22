package comps

import (
	"github.com/go-gl/mathgl/mgl32"
	"tophatdemon.com/total-invasion-ii/engine/math2/collision"
	"tophatdemon.com/total-invasion-ii/engine/scene"
)

type Body struct {
	Transform Transform
	Velocity  mgl32.Vec3
	Shape     collision.Shape
	Filter    collision.Mask // Determines which collision layers this body will respond to collisions with
	Layer     collision.Mask // The collision layer(s) that this body resides on
	LockY     bool           // When true, the body will not move on the Y axis in response to collisions.
}

func (body *Body) Body() *Body {
	return body
}

func (body *Body) ResolveBodyCollisions(deltaTime float32, bodies []scene.Handle) mgl32.Vec3 {
	movement := body.Velocity.Mul(deltaTime)

	for _, handle := range bodies {
		if collidingEnt, ok := scene.Get[HasBody](handle); ok {
			otherBody := collidingEnt.Body()
			if otherBody == nil || body == otherBody || body.Filter&otherBody.Layer == 0 {
				continue
			}

			// Bounding box check
			bbox := body.Shape.Extents().Translate(body.Transform.pos)
			if !bbox.Intersects(otherBody.Shape.Extents().Translate(otherBody.Transform.pos)) {
				continue
			}

			movement = movement.Add(body.Shape.PushOut(body.Transform.pos.Add(movement), otherBody.Transform.pos, otherBody.Shape))
		}
	}

	if body.LockY {
		// Restrict movement to the XZ plane
		movement[1] = 0.0
	}

	return movement
}

func (body *Body) OnLayer(layer collision.Mask) bool {
	return body.Layer&layer != 0
}
