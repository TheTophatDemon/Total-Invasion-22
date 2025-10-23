package world

import (
	"github.com/go-gl/mathgl/mgl32"
	"tophatdemon.com/total-invasion-ii/engine/math2"
	"tophatdemon.com/total-invasion-ii/engine/math2/collision"
	"tophatdemon.com/total-invasion-ii/engine/scene/comps"
)

const KNOCKBACK_FRICTION = 40.0

type Actor struct {
	MaxSpeed, AccelRate, Friction   float32
	GravityAccel                    float32
	MaxFallSpeed                    float32
	YawAngle                        float32 // Radians
	Health, MaxHealth, TargetHealth float32 // Health cannot be more than MaxHealth but will gradually drop to TargetHealth if overhealed.
	body                            comps.Body
	inputForward, inputStrafe       float32
	onGround                        bool
	knockbackForce                  mgl32.Vec3
	noisyTimer                      float32 // While this timer is > 0, enemies will be able to 'hear' the actor
	collisionFilter                 collision.Mask
}

func (actor *Actor) Update(deltaTime float32) {
	// Diminish noise level
	actor.noisyTimer = max(0.0, actor.noisyTimer-deltaTime)

	doGravity := actor.GravityAccel != 0.0
	if doGravity {
		distToBottom := (actor.body.Shape.Extents().Max.Y()) + 0.01
		downCast, _ := gWorld.Raycast(actor.body.Position, mgl32.Vec3{0.0, -1.0, 0.0}, actor.collisionFilter, distToBottom, &actor.body)
		actor.onGround = downCast.Hit
	}

	input := mgl32.Vec3{actor.inputStrafe, 0.0, -actor.inputForward}
	if input.LenSqr() != 0.0 {
		input = input.Normalize()
	}
	moveDir := mgl32.TransformCoordinate(input, mgl32.HomogRotate3DY(actor.YawAngle))

	// Apply acceleration
	actor.body.Velocity = actor.body.Velocity.Add(moveDir.Mul(actor.AccelRate * deltaTime))
	// Apply gravity
	actor.body.Velocity = actor.body.Velocity.Add(mgl32.Vec3{0.0, -actor.GravityAccel * deltaTime, 0.0})
	// Limit falling speed
	if actor.onGround {
		actor.body.Velocity = math2.Vec3WithY(actor.body.Velocity, 0.0)
	} else if actor.body.Velocity.Y() < -actor.MaxFallSpeed {
		actor.body.Velocity = math2.Vec3WithY(actor.body.Velocity, -actor.MaxFallSpeed)
	}

	// Apply friction
	if speed := actor.body.Velocity.Len(); speed > mgl32.Epsilon {
		frictionVec := actor.body.Velocity.Mul(-min(speed, actor.Friction*deltaTime) / speed)
		actor.body.Velocity = actor.body.Velocity.Add(frictionVec)
	}

	// Limit moving speed
	if speed := actor.body.Velocity.Len(); speed > actor.MaxSpeed && actor.MaxSpeed > mgl32.Epsilon {
		actor.body.Velocity = actor.body.Velocity.Mul(actor.MaxSpeed / speed)
	}

	// Apply knockback
	if !actor.knockbackForce.ApproxEqual(mgl32.Vec3{}) {
		// Apply friction to knockback
		if knockbackSpeed := actor.knockbackForce.Len(); knockbackSpeed > mgl32.Epsilon {
			frictionVec := actor.knockbackForce.Mul(-min(knockbackSpeed, KNOCKBACK_FRICTION*deltaTime) / knockbackSpeed)
			actor.knockbackForce = actor.knockbackForce.Add(frictionVec)
		} else {
			actor.knockbackForce = mgl32.Vec3{}
		}

		actor.body.Velocity = actor.body.Velocity.Add(actor.knockbackForce)
	}

	// Do collisions
	movement := actor.body.Velocity.Mul(deltaTime)
	movement = movement.Add(gWorld.bspTree.ResolveCollisions(&actor.body, movement, !doGravity, actor.collisionFilter))
	movement = movement.Add(gWorld.ResolveMapCollisions(&actor.body, movement, !doGravity, actor.collisionFilter))
	actor.body.TranslateV(movement)
}

func (actor *Actor) SetYaw(newYaw float32) {
	actor.YawAngle = newYaw
}

func (actor *Actor) FacingVec() mgl32.Vec3 {
	return mgl32.Vec3{
		-math2.Sin(actor.YawAngle),
		0.0,
		-math2.Cos(actor.YawAngle),
	}
}

func (actor *Actor) Body() *comps.Body {
	return &actor.body
}

func (actor *Actor) Position() mgl32.Vec3 {
	return actor.body.Position
}

func (actor *Actor) ApplyKnockback(force mgl32.Vec3) {
	actor.knockbackForce = force
}
