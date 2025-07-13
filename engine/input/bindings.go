package input

import (
	"fmt"
	"math"
	"strings"

	"github.com/go-gl/glfw/v3.3/glfw"
)

type Binding interface {
	IsPressed() bool
	Axis() float32
	Name() string
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

func (kb *KeyBinding) Name() string {
	return glfw.GetKeyName(kb.key, 0)
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

func (mbb *MouseButtonBinding) Name() string {
	switch mbb.button {
	case glfw.MouseButtonLeft:
		return "LMB"
	case glfw.MouseButtonRight:
		return "RMB"
	case glfw.MouseButtonMiddle:
		return "MMB"
	case glfw.MouseButton1 - glfw.MouseButton8:
		return fmt.Sprintf("MB%d", mbb.button)
	}
	return "UNKNOWN MOUSE BUTTON"
}

type MouseMovementBinding struct {
	axis        MouseAxis
	sensitivity float32
}

func (mmb *MouseMovementBinding) IsPressed() bool {
	return (math.Abs(float64(mmb.Axis())) > MouseDeadZone)
}

func (mmb *MouseMovementBinding) Axis() float32 {
	switch mmb.axis {
	case MouseAxisX:
		return float32(mouseDeltaX) * mmb.sensitivity
	case MouseAxisY:
		return float32(mouseDeltaY) * mmb.sensitivity
	}
	return 0.0
}

func (mmb *MouseMovementBinding) Name() string {
	switch mmb.axis {
	case MouseAxisX:
		return "Mouse Horizontal"
	case MouseAxisY:
		return "Mouse Vertical"
	}
	return "UNKOWN MOUSE MOVEMENT"
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
	if csb.progress == len(csb.sequence) {
		csb.progress = 0
	}
	if csb.progress < len(csb.sequence) && key == csb.sequence[csb.progress] {
		csb.progress += 1
	} else {
		csb.progress = 0
	}
}

func (csb *CharSequenceBinding) Name() string {
	var sb strings.Builder
	for k, key := range csb.sequence {
		sb.WriteString(glfw.GetKeyName(key, 0))
		if k < len(csb.sequence)-1 {
			sb.WriteString(", ")
		}
	}
	return sb.String()
}
