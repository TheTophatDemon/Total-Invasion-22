package input

import (
	"fmt"
	"strings"

	"tophatdemon.com/total-invasion-ii/engine/rgfw"
)

type (
	KeyBinding struct {
		bindingBase
		Key rgfw.Key
	}
	CharSequenceBinding struct {
		bindingBase
		Sequence []rgfw.Key
		progress int
	}
)

// KEYBOARD BINDING

func NewKeyBinding(key rgfw.Key) *KeyBinding {
	return new(KeyBinding{Key: key})
}

func (kb *KeyBinding) Pressed() bool {
	return kb.updatePressStates(window.IsKeyDown(kb.Key))
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
	case rgfw.KEY_SPACE:
		return "keySpace"
	case rgfw.KEY_APOSTROPHE:
		return "keyApostrophe"
	case rgfw.KEY_COMMA:
		return "keyComma"
	case rgfw.KEY_MINUS:
		return "keyMinus"
	case rgfw.KEY_PERIOD:
		return "keyPeriod"
	case rgfw.KEY_SLASH:
		return "keySlash"
	case rgfw.KEY_SEMICOLON:
		return "keySemicolon"
	case rgfw.KEY_EQUAL:
		return "keyEqual"
	case rgfw.KEY_BRACKET:
		return "keyLeftBracket"
	case rgfw.KEY_BACK_SLASH:
		return "keyBackslash"
	case rgfw.KEY_CLOSE_BRACKET:
		return "keyRightBracket"
	case rgfw.KEY_BACKTICK:
		return "keyGraveAccent"
	case rgfw.KEY_WORLD1:
		return "keyWorld1"
	case rgfw.KEY_WORLD2:
		return "keyWorld2"
	case rgfw.KEY_ESCAPE:
		return "keyEscape"
	case rgfw.KEY_ENTER:
		return "keyEnter"
	case rgfw.KEY_TAB:
		return "keyTab"
	case rgfw.KEY_BACK_SPACE:
		return "keyBackspace"
	case rgfw.KEY_INSERT:
		return "keyInsert"
	case rgfw.KEY_DELETE:
		return "keyDelete"
	case rgfw.KEY_RIGHT:
		return "keyRight"
	case rgfw.KEY_LEFT:
		return "keyLeft"
	case rgfw.KEY_DOWN:
		return "keyDown"
	case rgfw.KEY_UP:
		return "keyUp"
	case rgfw.KEY_PAGE_UP:
		return "keyPageUp"
	case rgfw.KEY_PAGE_DOWN:
		return "keyPageDown"
	case rgfw.KEY_HOME:
		return "keyHome"
	case rgfw.KEY_END:
		return "keyEnd"
	case rgfw.KEY_CAPS_LOCK:
		return "keyCapsLock"
	case rgfw.KEY_SCROLL_LOCK:
		return "keyScrollLock"
	case rgfw.KEY_NUM_LOCK:
		return "keyNumLock"
	case rgfw.KEY_PRINT_SCREEN:
		return "keyPrintScreen"
	case rgfw.KEY_PAUSE:
		return "keyPause"
	case rgfw.KEY_F1:
		return "keyF1"
	case rgfw.KEY_F2:
		return "keyF2"
	case rgfw.KEY_F3:
		return "keyF3"
	case rgfw.KEY_F4:
		return "keyF4"
	case rgfw.KEY_F5:
		return "keyF5"
	case rgfw.KEY_F6:
		return "keyF6"
	case rgfw.KEY_F7:
		return "keyF7"
	case rgfw.KEY_F8:
		return "keyF8"
	case rgfw.KEY_F9:
		return "keyF9"
	case rgfw.KEY_F10:
		return "keyF10"
	case rgfw.KEY_F11:
		return "keyF11"
	case rgfw.KEY_F12:
		return "keyF12"
	case rgfw.KEY_F13:
		return "keyF13"
	case rgfw.KEY_F14:
		return "keyF14"
	case rgfw.KEY_F15:
		return "keyF15"
	case rgfw.KEY_F16:
		return "keyF16"
	case rgfw.KEY_F17:
		return "keyF17"
	case rgfw.KEY_F18:
		return "keyF18"
	case rgfw.KEY_F19:
		return "keyF19"
	case rgfw.KEY_F20:
		return "keyF20"
	case rgfw.KEY_F21:
		return "keyF21"
	case rgfw.KEY_F22:
		return "keyF22"
	case rgfw.KEY_F23:
		return "keyF23"
	case rgfw.KEY_F24:
		return "keyF24"
	case rgfw.KEY_F25:
		return "keyF25"
	case rgfw.KEY_PAD0:
		return "keyKP0"
	case rgfw.KEY_PAD1:
		return "keyKP1"
	case rgfw.KEY_PAD2:
		return "keyKP2"
	case rgfw.KEY_PAD3:
		return "keyKP3"
	case rgfw.KEY_PAD4:
		return "keyKP4"
	case rgfw.KEY_PAD5:
		return "keyKP5"
	case rgfw.KEY_PAD6:
		return "keyKP6"
	case rgfw.KEY_PAD7:
		return "keyKP7"
	case rgfw.KEY_PAD8:
		return "keyKP8"
	case rgfw.KEY_PAD9:
		return "keyKP9"
	case rgfw.KEY_PAD_PERIOD:
		return "keyKPDecimal"
	case rgfw.KEY_PAD_SLASH:
		return "keyKPDivide"
	case rgfw.KEY_PAD_MULTIPLY:
		return "keyKPMultiply"
	case rgfw.KEY_PAD_MINUS:
		return "keyKPSubtract"
	case rgfw.KEY_PAD_PLUS:
		return "keyKPAdd"
	case rgfw.KEY_PAD_RETURN:
		return "keyKPEnter"
	case rgfw.KEY_PAD_EQUAL:
		return "keyKPEqual"
	case rgfw.KEY_SHIFT_L:
		return "keyLeftShift"
	case rgfw.KEY_CONTROL_L:
		return "keyLeftControl"
	case rgfw.KEY_ALT_L:
		return "keyLeftAlt"
	case rgfw.KEY_SUPER_L:
		return "keyLeftSuper"
	case rgfw.KEY_SHIFT_R:
		return "keyRightShift"
	case rgfw.KEY_CONTROL_R:
		return "keyRightControl"
	case rgfw.KEY_ALT_R:
		return "keyRightAlt"
	case rgfw.KEY_SUPER_R:
		return "keyRightSuper"
	case rgfw.KEY_MENU:
		return "keyMenu"
	default:
		return strings.ToUpper(string(kb.Key))
	}
}

func (kb *KeyBinding) MarshalTOML() ([]byte, error) {
	return fmt.Appendf(nil, "{ Key = %v }", kb.Key), nil
}

// CHAR SEQUENCE BINDING

func NewCharSequenceBinding(sequence ...rgfw.Key) *CharSequenceBinding {
	return new(CharSequenceBinding{Sequence: sequence})
}

func (csb *CharSequenceBinding) Pressed() bool {
	if lastKeyPressed != rgfw.KEY_NULL && csb.lastUpdatedFrame != inputFrameNumber {
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
		sb.WriteString(strings.ToUpper(string(key)))
	}
	return sb.String()
}
