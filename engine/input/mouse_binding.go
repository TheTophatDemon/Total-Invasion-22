package input

import (
	"math"

	"github.com/go-gl/glfw/v3.3/glfw"
)

const (
	MouseAxisPosX MouseAxis = iota
	MouseAxisNegX
	MouseAxisPosY
	MouseAxisNegY
	MouseDeadZone = 0.05
)

type (
	MouseButtonBinding struct {
		bindingBase
		Button glfw.MouseButton
	}
	MouseMovementBinding struct {
		bindingBase
		WhichAxis   MouseAxis
		Sensitivity float32
	}
	MouseAxis uint8
)

// MOUSE BUTTON BINDING

func NewMouseButtonBinding(butt glfw.MouseButton) *MouseButtonBinding {
	return new(MouseButtonBinding{Button: butt})
}

func (mbb *MouseButtonBinding) Pressed() bool {
	return mbb.updatePressStates(glfw.GetCurrentContext().GetMouseButton(mbb.Button) == glfw.Press)
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

func (mbb *MouseButtonBinding) String() string {
	switch mbb.Button {
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

// MOUSE MOVEMENT BINDING

func NewMouseMovementBinding(axis MouseAxis, sensitivity float32) *MouseMovementBinding {
	return new(MouseMovementBinding{WhichAxis: axis, Sensitivity: sensitivity})
}

func (mmb *MouseMovementBinding) Pressed() bool {
	return mmb.updatePressStates(math.Abs(float64(mmb.Axis())) > MouseDeadZone)
}

func (mmb *MouseMovementBinding) Axis() float32 {
	switch mmb.WhichAxis {
	case MouseAxisPosX:
		return max(0.0, float32(mouseDeltaX)*mmb.Sensitivity)
	case MouseAxisNegX:
		return -min(0.0, float32(mouseDeltaX)*mmb.Sensitivity)
	case MouseAxisPosY:
		return max(0.0, float32(mouseDeltaY)*mmb.Sensitivity)
	case MouseAxisNegY:
		return -min(0.0, float32(mouseDeltaY)*mmb.Sensitivity)
	}
	return 0.0
}

func (mmb *MouseMovementBinding) String() string {
	switch mmb.WhichAxis {
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
