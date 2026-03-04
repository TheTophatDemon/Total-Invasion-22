package input

import (
	"math"

	"github.com/go-gl/glfw/v3.3/glfw"
	"github.com/go-gl/mathgl/mgl32"
)

type (
	InputCaptureCallback func(newBinding Binding)
)

var anythingPressed, anyMouseButtonPressed, anyMouseButtonWasPressed bool
var captureCallback InputCaptureCallback = nil
var lastKeyPressed glfw.Key = glfw.KeyUnknown
var inputFrameNumber uint64
var mousePrevX, mousePrevY float64
var mouseDeltaX, mouseDeltaY float64
var mouseScrollY, mousePrevScrollY float32

func init() {
	mousePrevX, mousePrevY = math.NaN(), math.NaN()
}

func Init() {
	glfw.GetCurrentContext().SetKeyCallback(keyCallback)
	glfw.GetCurrentContext().SetMouseButtonCallback(mouseCallback)
	glfw.GetCurrentContext().SetScrollCallback(scrollCallback)
}

func PostUpdate() {
	inputFrameNumber += 1
	anythingPressed = false
	anyMouseButtonWasPressed = anyMouseButtonPressed
	anyMouseButtonPressed = false
	lastKeyPressed = glfw.KeyUnknown
	mousePosX, mousePosY := glfw.GetCurrentContext().GetCursorPos()
	if !math.IsNaN(mousePrevX) && !math.IsNaN(mousePrevY) {
		mouseDeltaX = mousePosX - mousePrevX
		mouseDeltaY = mousePosY - mousePrevY
	}
	mousePrevX, mousePrevY = mousePosX, mousePosY
	mousePrevScrollY = mouseScrollY
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

func CaptureInput(callback InputCaptureCallback) {
	captureCallback = callback
}

func IsCapturingInput() bool {
	return captureCallback != nil
}

func endCaptureInput(newBinding Binding) {
	if captureCallback != nil {
		captureCallback(newBinding)
		captureCallback = nil
	}
}

func MousePosition() mgl32.Vec2 {
	x, y := glfw.GetCurrentContext().GetCursorPos()
	return mgl32.Vec2{float32(x), float32(y)}
}

func MouseDelta() mgl32.Vec2 {
	return mgl32.Vec2{float32(mouseDeltaX), float32(mouseDeltaY)}
}

func MouseScroll() float32 {
	return mouseScrollY
}

func MouseScrollDelta() float32 {
	return mouseScrollY - mousePrevScrollY
}

func SetMousePosition(x, y float32) {
	mousePrevX = float64(x)
	mousePrevY = float64(y)
	mouseDeltaX, mouseDeltaY = 0, 0
	glfw.GetCurrentContext().SetCursorPos(mousePrevX, mousePrevY)
}

func IsMouseButtonDown(button glfw.MouseButton) bool {
	return glfw.GetCurrentContext().GetMouseButton(button) == glfw.Press
}

func AnyMouseButtonPressed() bool {
	return anyMouseButtonPressed
}

func AnyMouseButtonJustPressed() bool {
	return anyMouseButtonPressed && !anyMouseButtonWasPressed
}

func IsAnythingPressed() bool {
	return anythingPressed
}

func keyCallback(w *glfw.Window, key glfw.Key, scancode int, action glfw.Action, mods glfw.ModifierKey) {
	if action == glfw.Press {
		anythingPressed = true
		lastKeyPressed = key
		if captureCallback != nil {
			endCaptureInput(NewKeyBinding(key))
		}
	}
}

func mouseCallback(w *glfw.Window, button glfw.MouseButton, action glfw.Action, mods glfw.ModifierKey) {
	if action == glfw.Press {
		anythingPressed = true
		anyMouseButtonPressed = true
	}
}

func scrollCallback(w *glfw.Window, xoff float64, yoff float64) {
	mouseScrollY += float32(yoff)
}
