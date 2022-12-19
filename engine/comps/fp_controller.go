package comps

import (
	"fmt"

	"tophatdemon.com/total-invasion-ii/engine/input"
	"tophatdemon.com/total-invasion-ii/engine/math2"
	"tophatdemon.com/total-invasion-ii/engine/scene"
)

type FirstPersonController struct {
	ForwardAction, BackAction           input.Action
	StrafeLeftAction, StrafeRightAction input.Action
	LookHorzAction, LookVertAction      input.Action
}

func (controller *FirstPersonController) UpdateComponent(sc *scene.Scene, entity scene.Entity, deltaTime float32) {
	var movement *Movement
	movement, err := scene.ExtractComponent(sc, entity, movement)
	if err != nil {
		fmt.Println(err)
		return
	}

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
