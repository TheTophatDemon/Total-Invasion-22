package input

import (
	"log"
	"math"

	"github.com/go-gl/glfw/v3.3/glfw"
)

type Action uint16
type MouseAxis uint8

const (
	MOUSE_AXIS_X   MouseAxis = 0
	MOUSE_AXIS_Y   MouseAxis = 1
	MOUSE_DEADZONE           = 0.05
)

const (
	ERRT_NO_ACTION string = "WARNING: Action %v not bound.\n"
)

var bindings map[Action]Binding
var bindingsWerePressed map[Action]bool

var mousePrevX, mousePrevY float64
var mouseDeltaX, mouseDeltaY float64

func init() {
	bindings = make(map[Action]Binding)
	bindingsWerePressed = make(map[Action]bool)
	mousePrevX, mousePrevY = math.NaN(), math.NaN()
}

func Init() {
	glfw.GetCurrentContext().SetKeyCallback(keyCallback)
}

func Update() {
	mousePosX, mousePosY := glfw.GetCurrentContext().GetCursorPos()
	if !math.IsNaN(mousePrevX) && !math.IsNaN(mousePrevY) {
		mouseDeltaX = mousePosX - mousePrevX
		mouseDeltaY = mousePosY - mousePrevY
	}
	mousePrevX, mousePrevY = mousePosX, mousePosY

	for action, binding := range bindings {
		bindingsWerePressed[action] = binding.IsPressed()
	}
}

func TrapMouse() {
	glfw.GetCurrentContext().SetInputMode(glfw.CursorMode, glfw.CursorDisabled)
}

func UntrapMouse() {
	glfw.GetCurrentContext().SetInputMode(glfw.CursorMode, glfw.CursorNormal)
}

func IsMouseTrapped() bool {
	return glfw.GetCurrentContext().GetInputMode(glfw.CursorMode) == glfw.CursorDisabled
}

func BindActionKey(action Action, key glfw.Key) {
	bindings[action] = &KeyBinding{key}
	bindingsWerePressed[action] = false
}

func BindActionMouseButton(action Action, button glfw.MouseButton) {
	bindings[action] = &MouseButtonBinding{button}
	bindingsWerePressed[action] = false
}

func BindActionMouseMove(action Action, axis MouseAxis, sensitivity float32) {
	bindings[action] = &MouseMovementBinding{axis, sensitivity}
	bindingsWerePressed[action] = false
}

func BindActionCharSequence(action Action, sequence []glfw.Key) {
	bindings[action] = &CharSequenceBinding{sequence: sequence, progress: 0}
	bindingsWerePressed[action] = false
}

func IsActionPressed(action Action) bool {
	bind, ok := bindings[action]
	if !ok {
		log.Printf(ERRT_NO_ACTION, action)
		return false
	}
	return bind.IsPressed()
}

func IsActionJustPressed(action Action) bool {
	bind, ok := bindings[action]
	wasPressed, ok2 := bindingsWerePressed[action]
	if !ok || !ok2 {
		log.Printf(ERRT_NO_ACTION, action)
		return false
	}
	return bind.IsPressed() && !wasPressed
}

func ActionAxis(action Action) float32 {
	bind, ok := bindings[action]
	if !ok {
		log.Printf(ERRT_NO_ACTION, action)
		return 0.0
	}
	return bind.Axis()
}

func keyCallback(w *glfw.Window, key glfw.Key, scancode int, action glfw.Action, mods glfw.ModifierKey) {
	if action == glfw.Press {
		for _, binding := range bindings {
			csb, isCSB := binding.(*CharSequenceBinding)
			if isCSB {
				csb.OnKeyPress(key)
			}
		}
	}
}
