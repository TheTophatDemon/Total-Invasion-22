package input

import (
	"fmt"
	"math"

	"tophatdemon.com/total-invasion-ii/engine/rgfw"
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
	MouseButtonBinding struct {
		bindingBase
		MouseButton rgfw.MouseButton
	}
	MouseMovementBinding struct {
		bindingBase
		MouseAxis MouseAxis
	}
	MouseAxis uint8
)

// MOUSE BUTTON BINDING

func NewMouseButtonBinding(butt rgfw.MouseButton) *MouseButtonBinding {
	return new(MouseButtonBinding{MouseButton: butt})
}

func (mbb *MouseButtonBinding) Pressed() bool {
	return mbb.updatePressStates(window.IsMouseDown(mbb.MouseButton))
}

func (mbb *MouseButtonBinding) JustPressed() bool {
	mbb.Pressed()
	return mbb.justPressed
}

func (mbb *MouseButtonBinding) JustReleased() bool {
	mbb.Pressed()
	return mbb.justReleased
}

func (mbb *MouseButtonBinding) Axis() float32 {
	if mbb.Pressed() {
		return 1.0
	} else {
		return 0.0
	}
}

func (mbb *MouseButtonBinding) LocalizationKey() string {
	switch mbb.MouseButton {
	case rgfw.MOUSE_BUTTON_LEFT:
		return "mouseButtonLeft"
	case rgfw.MOUSE_BUTTON_RIGHT:
		return "mouseButtonRight"
	case rgfw.MOUSE_BUTTON_MIDDLE:
		return "mouseButtonMiddle"
	case rgfw.MOUSE_BUTTON_MISC_1:
		return "mouseButton4"
	case rgfw.MOUSE_BUTTON_MISC_2:
		return "mouseButton5"
	case rgfw.MOUSE_BUTTON_MISC_3:
		return "mouseButton6"
	case rgfw.MOUSE_BUTTON_MISC_4:
		return "mouseButton7"
	case rgfw.MOUSE_BUTTON_MISC_5:
		return "mouseButton8"
	}
	return "???"
}

func (mbb *MouseButtonBinding) MarshalTOML() ([]byte, error) {
	return fmt.Appendf(nil, "{ MouseButton = %v }", mbb.MouseButton), nil
}

// MOUSE MOVEMENT BINDING

func NewMouseMovementBinding(axis MouseAxis) *MouseMovementBinding {
	return new(MouseMovementBinding{MouseAxis: axis})
}

func (mmb *MouseMovementBinding) Pressed() bool {
	return mmb.updatePressStates(math.Abs(float64(mmb.Axis())) > MouseDeadZone)
}

func (mmb *MouseMovementBinding) Axis() float32 {
	scaledX := mouseDeltaX / 10.0
	scaledY := mouseDeltaY / 10.0
	switch mmb.MouseAxis {
	case MouseAxisPosX:
		return max(0.0, float32(scaledX))
	case MouseAxisNegX:
		return -min(0.0, float32(scaledX))
	case MouseAxisPosY:
		return max(0.0, float32(scaledY))
	case MouseAxisNegY:
		return -min(0.0, float32(scaledY))
	}
	return 0.0
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

func (mmb *MouseMovementBinding) JustPressed() bool {
	mmb.Pressed()
	return mmb.justPressed
}

func (mmb *MouseMovementBinding) JustReleased() bool {
	mmb.Pressed()
	return mmb.justReleased
}

func (mmb *MouseMovementBinding) MarshalTOML() ([]byte, error) {
	return fmt.Appendf(nil, "{ MouseAxis = %v }", mmb.MouseAxis), nil
}
