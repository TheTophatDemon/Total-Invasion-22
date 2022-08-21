package systems

import (
	"tophatdemon.com/total-invasion-ii/engine/comps"
	"tophatdemon.com/total-invasion-ii/engine/ecs"
	"tophatdemon.com/total-invasion-ii/engine/input"
	"tophatdemon.com/total-invasion-ii/engine/math2"
)

func UpdateFirstPersonControllers(
	deltaTime float32,
	world *ecs.World,
	movements   *ecs.ComponentStorage[comps.Movement],
	controllers *ecs.ComponentStorage[comps.FirstPersonController]) {

	ents := world.Query(func(id ecs.EntID)bool{
		return movements.Has(id) && controllers.Has(id)
	})

	for _, ent := range ents {
		movement, _ := movements.Get(ent)
		controller, _ := controllers.Get(ent)
		
		if input.IsActionPressed(controller.ForwardAction) {
			movement.InputForward = 1.0
		} else if input.IsActionPressed(controller.BackAction) {
			movement.InputForward = -1.0
		} else {
			movement.InputForward = 0.0
		}

		if input.IsActionPressed(controller.StrafeRightAction) {
			movement.InputStrafe = 1.0
		} else if input.IsActionPressed(controller.StrafeLeftAction) {
			movement.InputStrafe = -1.0
		} else {
			movement.InputStrafe = 0.0
		}

		movement.YawAngle -= input.ActionAxis(controller.LookHorzAction)
		movement.PitchAngle -= input.ActionAxis(controller.LookVertAction)
		movement.PitchAngle = math2.Clamp(movement.PitchAngle, -math2.HALF_PI, math2.HALF_PI)
	}
}