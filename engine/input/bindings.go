package input

import (
	"math"

	"github.com/go-gl/glfw/v3.3/glfw"
)

type Binding interface {
	IsPressed() bool
	Axis() float32
}

type KeyBinding struct {
	key glfw.Key
}

func (kb *KeyBinding) IsPressed() bool {
	return (glfw.GetCurrentContext().GetKey(kb.key) == glfw.Press)
}

func (kb *KeyBinding) Axis() float32 {
	if kb.IsPressed() {
		return 1.0
	} else {
		return 0.0
	}
}

type MouseButtonBinding struct {
	button glfw.MouseButton
}

func (mbb *MouseButtonBinding) IsPressed() bool {
	return (glfw.GetCurrentContext().GetMouseButton(mbb.button) == glfw.Press)
}

func (mbb *MouseButtonBinding) Axis() float32 {
	if mbb.IsPressed() {
		return 1.0
	} else {
		return 0.0
	}
}

type MouseMovementBinding struct {
	axis        MouseAxis
	sensitivity float32
}

func (mmb *MouseMovementBinding) IsPressed() bool {
	return (math.Abs(float64(mmb.Axis())) > MOUSE_DEADZONE)
}

func (mmb *MouseMovementBinding) Axis() float32 {
	switch mmb.axis {
	case MOUSE_AXIS_X:
		// return float32(math.Min(math.Max(mouseDeltaX, -1.0), 1.0))
		return float32(mouseDeltaX) * mmb.sensitivity
	case MOUSE_AXIS_Y:
		// return float32(math.Min(math.Max(mouseDeltaY, -1.0), 1.0))
		return float32(mouseDeltaY) * mmb.sensitivity
	}
	return 0.0
}

type CharSequenceBinding struct {
	sequence []glfw.Key
	progress int
}

func (csb *CharSequenceBinding) IsPressed() bool {
	return csb.progress == len(csb.sequence)
}

func (csb *CharSequenceBinding) Axis() float32 {
	return float32(csb.progress) / float32(len(csb.sequence))
}

func (csb *CharSequenceBinding) OnKeyPress(key glfw.Key) {
	if csb.progress < len(csb.sequence) && key == csb.sequence[csb.progress] {
		csb.progress += 1
	} else {
		csb.progress = 0
	}
}
