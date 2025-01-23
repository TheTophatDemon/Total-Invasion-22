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

	// Called when the body intersects another body, include those that don't pass the collision filter. Can be nil.
	OnIntersect func(
		collidingEntity HasBody,
		collision collision.Result,
		deltaTime float32,
	)
}

func (body *Body) Body() *Body {
	return body
}

func (body *Body) MoveAndCollide(deltaTime float32, bodiesIter func() (HasBody, scene.Handle)) {
	before := body.Transform.Position()

	movement := body.Velocity.Mul(deltaTime)
	for collidingEnt, _ := bodiesIter(); collidingEnt != nil; collidingEnt, _ = bodiesIter() {
		body.ResolveCollision(movement, collidingEnt, deltaTime)
	}

	body.Transform.TranslateV(movement)

	if body.LockY {
		// Restrict movement to the XZ plane
		after := body.Transform.Position()
		body.Transform.SetPosition(mgl32.Vec3{after.X(), before.Y(), after.Z()})
	}
}

// Change the position of this body so that it doesn't collide with the other body.
func (body *Body) ResolveCollision(movement mgl32.Vec3, otherEnt HasBody, deltaTime float32) {
	otherBody := otherEnt.Body()
	if otherBody == nil || body == otherBody || (body.Filter&otherBody.Layer == 0 && body.OnIntersect == nil) {
		return
	}

	movingShape, isMovingShape := body.Shape.(collision.MovingShape)
	if !isMovingShape {
		return
	}

	nextPosition := body.Transform.Position().Add(movement)

	// Bounding box check
	bbox := movingShape.Extents().Translate(nextPosition)
	if !bbox.Intersects(otherBody.Shape.Extents().Translate(otherBody.Transform.Position())) {
		return
	}

	res := movingShape.ResolveCollision(body.Transform.Position(), movement, otherBody.Transform.Position(), otherBody.Shape)
	if res.Hit {
		if body.Filter&otherBody.Layer != 0 {
			body.Transform.TranslateV(res.Normal.Mul(res.Penetration))
		}
		if body.OnIntersect != nil {
			body.OnIntersect(otherEnt, res, deltaTime)
		}
	}
}

func (body *Body) OnLayer(layer collision.Mask) bool {
	return body.Layer&layer != 0
}
