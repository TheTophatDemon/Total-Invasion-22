package input

import (
	"fmt"

	"tophatdemon.com/total-invasion-ii/engine/glfw"
)

const (
	MouseAxisNone MouseAxis = iota
	MouseAxisPosX
	MouseAxisNegX
	MouseAxisPosY
	MouseAxisNegY
	MouseDeadZone = 0.05
)

type (
	MouseAxis uint8
)

func (mbb *MouseButtonBinding) LocalizationKey() string {
	switch mbb.MouseButton {
	case glfw.MouseButtonLeft:
		return "mouseButtonLeft"
	case glfw.MouseButtonRight:
		return "mouseButtonRight"
	case glfw.MouseButtonMiddle:
		return "mouseButtonMiddle"
	case glfw.MouseButton4:
		return "mouseButton4"
	case glfw.MouseButton5:
		return "mouseButton5"
	case glfw.MouseButton6:
		return "mouseButton6"
	case glfw.MouseButton7:
		return "mouseButton7"
	case glfw.MouseButton8:
		return "mouseButton8"
	}
	return "???"
}

func (mbb *MouseButtonBinding) MarshalTOML() ([]byte, error) {
	return fmt.Appendf(nil, "{ MouseButton = %v }", mbb.MouseButton), nil
}

func (mmb *MouseMovementBinding) LocalizationKey() string {
	switch mmb.MouseAxis {
	case MouseAxisPosX:
		return "mousePosX"
	case MouseAxisNegX:
		return "mouseNegX"
	case MouseAxisPosY:
		return "mousePosY"
	case MouseAxisNegY:
		return "mouseNegY"
	}
	return "???"
}

func (mmb *MouseMovementBinding) MarshalTOML() ([]byte, error) {
	return fmt.Appendf(nil, "{ MouseAxis = %v }", mmb.MouseAxis), nil
}
