package input

import (
	"strings"

	"github.com/go-gl/glfw/v3.3/glfw"
)

type (
	KeyBinding struct {
		bindingBase
		Key glfw.Key
	}
	CharSequenceBinding struct {
		bindingBase
		Sequence []glfw.Key
		progress int
	}
)

// KEYBOARD BINDING

func NewKeyBinding(key glfw.Key) *KeyBinding {
	return new(KeyBinding{Key: key})
}

func (kb *KeyBinding) Pressed() bool {
	return kb.updatePressStates(glfw.GetCurrentContext().GetKey(kb.Key) == glfw.Press)
}

func (kb *KeyBinding) JustPressed() bool {
	kb.Pressed()
	return kb.justPressed
}

func (kb *KeyBinding) JustReleased() bool {
	kb.Pressed()
	return kb.justReleased
}

func (kb *KeyBinding) Axis() float32 {
	if kb.Pressed() {
		return 1.0
	} else {
		return 0.0
	}
}

func (kb *KeyBinding) String() string {
	return strings.ToUpper(glfw.GetKeyName(kb.Key, 0))
}

// CHAR SEQUENCE BINDING

func NewCharSequenceBinding(sequence ...glfw.Key) *CharSequenceBinding {
	return new(CharSequenceBinding{Sequence: sequence})
}

func (csb *CharSequenceBinding) Pressed() bool {
	if lastKeyPressed != glfw.KeyUnknown && csb.lastUpdatedFrame != inputFrameNumber {
		if csb.progress == len(csb.Sequence) {
			csb.progress = 0
		}
		if csb.progress < len(csb.Sequence) && lastKeyPressed == csb.Sequence[csb.progress] {
			csb.progress += 1
		} else {
			csb.progress = 0
		}
		csb.lastUpdatedFrame = inputFrameNumber
	}
	return csb.updatePressStates(csb.progress == len(csb.Sequence))
}

func (csb *CharSequenceBinding) JustPressed() bool {
	csb.Pressed()
	return csb.justPressed
}

func (csb *CharSequenceBinding) JustReleased() bool {
	csb.Pressed()
	return csb.justReleased
}

func (csb *CharSequenceBinding) Axis() float32 {
	return float32(csb.progress) / float32(len(csb.Sequence))
}

func (csb *CharSequenceBinding) String() string {
	var sb strings.Builder
	for _, key := range csb.Sequence {
		sb.WriteString(strings.ToUpper(glfw.GetKeyName(key, 0)))
	}
	return sb.String()
}
