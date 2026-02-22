package input

import (
	"fmt"
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

func (kb *KeyBinding) LocalizationKey() string {
	switch kb.Key {
	case glfw.KeySpace:
		return "keySpace"
	case glfw.KeyApostrophe:
		return "keyApostrophe"
	case glfw.KeyComma:
		return "keyComma"
	case glfw.KeyMinus:
		return "keyMinus"
	case glfw.KeyPeriod:
		return "keyPeriod"
	case glfw.KeySlash:
		return "keySlash"
	case glfw.KeySemicolon:
		return "keySemicolon"
	case glfw.KeyEqual:
		return "keyEqual"
	case glfw.KeyLeftBracket:
		return "keyLeftBracket"
	case glfw.KeyBackslash:
		return "keyBackslash"
	case glfw.KeyRightBracket:
		return "keyRightBracket"
	case glfw.KeyGraveAccent:
		return "keyGraveAccent"
	case glfw.KeyWorld1:
		return "keyWorld1"
	case glfw.KeyWorld2:
		return "keyWorld2"
	case glfw.KeyEscape:
		return "keyEscape"
	case glfw.KeyEnter:
		return "keyEnter"
	case glfw.KeyTab:
		return "keyTab"
	case glfw.KeyBackspace:
		return "keyBackspace"
	case glfw.KeyInsert:
		return "keyInsert"
	case glfw.KeyDelete:
		return "keyDelete"
	case glfw.KeyRight:
		return "keyRight"
	case glfw.KeyLeft:
		return "keyLeft"
	case glfw.KeyDown:
		return "keyDown"
	case glfw.KeyUp:
		return "keyUp"
	case glfw.KeyPageUp:
		return "keyPageUp"
	case glfw.KeyPageDown:
		return "keyPageDown"
	case glfw.KeyHome:
		return "keyHome"
	case glfw.KeyEnd:
		return "keyEnd"
	case glfw.KeyCapsLock:
		return "keyCapsLock"
	case glfw.KeyScrollLock:
		return "keyScrollLock"
	case glfw.KeyNumLock:
		return "keyNumLock"
	case glfw.KeyPrintScreen:
		return "keyPrintScreen"
	case glfw.KeyPause:
		return "keyPause"
	case glfw.KeyF1:
		return "keyF1"
	case glfw.KeyF2:
		return "keyF2"
	case glfw.KeyF3:
		return "keyF3"
	case glfw.KeyF4:
		return "keyF4"
	case glfw.KeyF5:
		return "keyF5"
	case glfw.KeyF6:
		return "keyF6"
	case glfw.KeyF7:
		return "keyF7"
	case glfw.KeyF8:
		return "keyF8"
	case glfw.KeyF9:
		return "keyF9"
	case glfw.KeyF10:
		return "keyF10"
	case glfw.KeyF11:
		return "keyF11"
	case glfw.KeyF12:
		return "keyF12"
	case glfw.KeyF13:
		return "keyF13"
	case glfw.KeyF14:
		return "keyF14"
	case glfw.KeyF15:
		return "keyF15"
	case glfw.KeyF16:
		return "keyF16"
	case glfw.KeyF17:
		return "keyF17"
	case glfw.KeyF18:
		return "keyF18"
	case glfw.KeyF19:
		return "keyF19"
	case glfw.KeyF20:
		return "keyF20"
	case glfw.KeyF21:
		return "keyF21"
	case glfw.KeyF22:
		return "keyF22"
	case glfw.KeyF23:
		return "keyF23"
	case glfw.KeyF24:
		return "keyF24"
	case glfw.KeyF25:
		return "keyF25"
	case glfw.KeyKP0:
		return "keyKP0"
	case glfw.KeyKP1:
		return "keyKP1"
	case glfw.KeyKP2:
		return "keyKP2"
	case glfw.KeyKP3:
		return "keyKP3"
	case glfw.KeyKP4:
		return "keyKP4"
	case glfw.KeyKP5:
		return "keyKP5"
	case glfw.KeyKP6:
		return "keyKP6"
	case glfw.KeyKP7:
		return "keyKP7"
	case glfw.KeyKP8:
		return "keyKP8"
	case glfw.KeyKP9:
		return "keyKP9"
	case glfw.KeyKPDecimal:
		return "keyKPDecimal"
	case glfw.KeyKPDivide:
		return "keyKPDivide"
	case glfw.KeyKPMultiply:
		return "keyKPMultiply"
	case glfw.KeyKPSubtract:
		return "keyKPSubtract"
	case glfw.KeyKPAdd:
		return "keyKPAdd"
	case glfw.KeyKPEnter:
		return "keyKPEnter"
	case glfw.KeyKPEqual:
		return "keyKPEqual"
	case glfw.KeyLeftShift:
		return "keyLeftShift"
	case glfw.KeyLeftControl:
		return "keyLeftControl"
	case glfw.KeyLeftAlt:
		return "keyLeftAlt"
	case glfw.KeyLeftSuper:
		return "keyLeftSuper"
	case glfw.KeyRightShift:
		return "keyRightShift"
	case glfw.KeyRightControl:
		return "keyRightControl"
	case glfw.KeyRightAlt:
		return "keyRightAlt"
	case glfw.KeyRightSuper:
		return "keyRightSuper"
	case glfw.KeyMenu:
		return "keyMenu"
	default:
		return strings.ToUpper(glfw.GetKeyName(kb.Key, 0))
	}
}

func (kb *KeyBinding) MarshalTOML() ([]byte, error) {
	return fmt.Appendf(nil, "{ Key = %v }", kb.Key), nil
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

func (csb *CharSequenceBinding) LocalizationKey() string {
	var sb strings.Builder
	for _, key := range csb.Sequence {
		sb.WriteString(strings.ToUpper(glfw.GetKeyName(key, 0)))
	}
	return sb.String()
}
