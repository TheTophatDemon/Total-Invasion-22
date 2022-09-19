package systems

import (
	"github.com/go-gl/mathgl/mgl32"
	"tophatdemon.com/total-invasion-ii/engine/comps"
	"tophatdemon.com/total-invasion-ii/engine/ecs"
	// "tophatdemon.com/total-invasion-ii/engine/math2"
)



func UpdateMovement(
	deltaTime float32,
	world *ecs.World,
	movements *ecs.ComponentStorage[comps.Movement],
	transforms *ecs.ComponentStorage[comps.Transform]) {
	
	movers := world.Query(func(id ecs.EntID)bool {
		return movements.Has(id) && transforms.Has(id)
	})

	for _, ent := range movers {
		movement, _ := movements.Get(ent)
		transform, _ := transforms.Get(ent)
		
		strafe := movement.InputStrafe * movement.MaxSpeed * deltaTime
		forward := movement.InputForward * movement.MaxSpeed * deltaTime
		
		globalMove := mgl32.TransformCoordinate(mgl32.Vec3{strafe, 0.0, -forward}, transform.GetMatrix().Mat3().Mat4())
		transform.Translate(globalMove[0], globalMove[1], globalMove[2])
		transform.SetRotation(movement.PitchAngle, movement.YawAngle, 0.0)
	}
}