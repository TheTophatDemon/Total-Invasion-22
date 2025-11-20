package input

import (
	"log"
	"math"

	"github.com/go-gl/glfw/v3.3/glfw"
	"github.com/go-gl/mathgl/mgl32"
	"tophatdemon.com/total-invasion-ii/engine/failure"
)

type Action string
type MouseAxis uint8

const (
	MouseAxisX    MouseAxis = 0
	MouseAxisY    MouseAxis = 1
	MouseDeadZone           = 0.05
)

// Maximum number of bindings allowed per action.
const MaxBindCount = 2

const (
	txtNoAction string = "WARNING: Action %v not bound.\n"
)

var bindingMap map[Action][MaxBindCount]Binding
var bindingsWerePressed map[Action]bool

var mousePrevX, mousePrevY float64
var mouseDeltaX, mouseDeltaY float64

func init() {
	bindingMap = make(map[Action][MaxBindCount]Binding)
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

	for action, bindings := range bindingMap {
		anyPressed := false
		for _, binding := range bindings {
			if binding != nil && binding.IsPressed() {
				anyPressed = true
				break
			}
		}
		bindingsWerePressed[action] = anyPressed
	}
}

func TrapMouse() {
	glfw.GetCurrentContext().SetInputMode(glfw.CursorMode, glfw.CursorDisabled)
	mousePrevX, mousePrevY = glfw.GetCurrentContext().GetCursorPos()
}

func UntrapMouse() {
	glfw.GetCurrentContext().SetInputMode(glfw.CursorMode, glfw.CursorNormal)
}

func IsMouseTrapped() bool {
	return glfw.GetCurrentContext().GetInputMode(glfw.CursorMode) == glfw.CursorDisabled
}

func MousePosition() mgl32.Vec2 {
	x, y := glfw.GetCurrentContext().GetCursorPos()
	return mgl32.Vec2{float32(x), float32(y)}
}

func MouseDelta() mgl32.Vec2 {
	return mgl32.Vec2{float32(mouseDeltaX), float32(mouseDeltaY)}
}

func SetMousePosition(x, y float32) {
	mousePrevX = float64(x)
	mousePrevY = float64(y)
	mouseDeltaX, mouseDeltaY = 0, 0
	glfw.GetCurrentContext().SetCursorPos(mousePrevX, mousePrevY)
}

func BindActionKey(action Action, key glfw.Key) {
	appendBinding(action, &KeyBinding{key})
	bindingsWerePressed[action] = false
}

func BindActionMouseButton(action Action, button glfw.MouseButton) {
	appendBinding(action, &MouseButtonBinding{button})
	bindingsWerePressed[action] = false
}

func BindActionMouseMove(action Action, axis MouseAxis, sensitivity float32) {
	appendBinding(action, &MouseMovementBinding{axis, sensitivity})
	bindingsWerePressed[action] = false
}

func BindActionCharSequence(action Action, sequence []glfw.Key) {
	appendBinding(action, &CharSequenceBinding{sequence: sequence, progress: 0})
	bindingsWerePressed[action] = false
}

// Returns booleans indicating if the action was just pressed, just released, or is otherwise being held down.
func ActionPressStates(action Action) (pressed, justPressed, justReleased bool) {
	wasPressed, ok := bindingsWerePressed[action]
	if !ok {
		log.Printf(txtNoAction, action)
		return
	}
	pressed = IsActionPressed(action)
	justPressed = pressed && !wasPressed
	justReleased = !pressed && wasPressed
	return
}

func IsActionPressed(action Action) bool {
	bindings, ok := bindingMap[action]
	if !ok {
		failure.LogErrWithLocation(txtNoAction, action)
		return false
	}
	for _, bind := range bindings {
		if bind != nil && bind.IsPressed() {
			return true
		}
	}
	return false
}

func IsActionJustPressed(action Action) bool {
	bindings, ok := bindingMap[action]
	wasPressed, ok2 := bindingsWerePressed[action]
	if !ok || !ok2 {
		failure.LogErrWithLocation(txtNoAction, action)
		return false
	}
	anyPressed := false
	for _, bind := range bindings {
		if bind != nil && bind.IsPressed() {
			anyPressed = true
			break
		}
	}
	return anyPressed && !wasPressed
}

func IsActionJustReleased(action Action) bool {
	bindings, ok := bindingMap[action]
	wasPressed, ok2 := bindingsWerePressed[action]
	if !ok || !ok2 {
		failure.LogErrWithLocation(txtNoAction, action)
		return false
	}
	anyPressed := false
	for _, bind := range bindings {
		if bind != nil && bind.IsPressed() {
			anyPressed = true
			break
		}
	}
	return !anyPressed && wasPressed
}

func ActionAxis(action Action) float32 {
	bindings, ok := bindingMap[action]
	if !ok {
		failure.LogErrWithLocation(txtNoAction, action)
		return 0.0
	}
	for _, bind := range bindings {
		if bind == nil {
			continue
		}
		if axis := bind.Axis(); axis != 0.0 {
			return axis
		}
	}
	return 0.0
}

func ActionBindings(action Action) ([MaxBindCount]Binding, bool) {
	binds, ok := bindingMap[action]
	return binds, ok
}

func IsMouseButtonDown(button glfw.MouseButton) bool {
	return glfw.GetCurrentContext().GetMouseButton(button) == glfw.Press
}

func keyCallback(w *glfw.Window, key glfw.Key, scancode int, action glfw.Action, mods glfw.ModifierKey) {
	if action == glfw.Press {
		for _, bindings := range bindingMap {
			for _, binding := range bindings {
				csb, isCSB := binding.(*CharSequenceBinding)
				if isCSB {
					csb.OnKeyPress(key)
				}
			}
		}
	}
}

func appendBinding(action Action, newBinding Binding) {
	bindings := bindingMap[action]
	for b, binding := range bindings {
		if binding == nil || b == len(bindings)-1 {
			bindings[b] = newBinding
			break
		}
	}
	bindingMap[action] = bindings
}
