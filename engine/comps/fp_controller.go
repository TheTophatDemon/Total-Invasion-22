package comps

import "tophatdemon.com/total-invasion-ii/engine/input"

type FirstPersonController struct {
	ForwardAction, BackAction             input.Action
	StrafeLeftAction, StrafeRightAction   input.Action
	LookHorzAction, LookVertAction        input.Action
}