package input

import (
	"log"
	"math"

	"github.com/go-gl/glfw/v3.3/glfw"
)

type Action string

type Binding interface {
	IsPressed() bool
	Axis()  float32
}

type MouseAxis uint8

const (
	MOUSE_AXIS_X MouseAxis = 0
	MOUSE_AXIS_Y MouseAxis = 1
	MOUSE_DEADZONE = 0.05
)

const (
	ERRT_NO_ACTION string = "WARNING: Action %v not bound.\n"
)

var bindings map[Action]Binding

var mousePrevX, mousePrevY float64
var mouseDeltaX, mouseDeltaY float64

func init() {
	bindings = make(map[Action]Binding)
	mousePrevX, mousePrevY = math.NaN(), math.NaN()
}

func Update() {
	mousePosX, mousePosY := glfw.GetCurrentContext().GetCursorPos()
	if !math.IsNaN(mousePrevX) && !math.IsNaN(mousePrevY) {
		mouseDeltaX = mousePosX - mousePrevX
		mouseDeltaY = mousePosY - mousePrevY
	}
	mousePrevX, mousePrevY = mousePosX, mousePosY
}

func TrapMouse() {
	glfw.GetCurrentContext().SetInputMode(glfw.CursorMode, glfw.CursorDisabled)
}

func UntrapMouse() {
	glfw.GetCurrentContext().SetInputMode(glfw.CursorMode, glfw.CursorNormal)
}

func BindActionKey(action Action, key glfw.Key) {
	bindings[action] = &KeyBinding{ key }
}

func BindActionMouseMove(action Action, axis MouseAxis, sensitivity float32) {
	bindings[action] = &MouseMovementBinding{ axis, sensitivity }
}

func IsActionPressed(action Action) bool {
	bind, ok := bindings[action]
	if !ok {
		log.Printf(ERRT_NO_ACTION, action)
		return false
	}
	return bind.IsPressed()
}

func ActionAxis(action Action) float32 {
	bind, ok := bindings[action]
	if !ok {
		log.Printf(ERRT_NO_ACTION, action)
		return 0.0
	}
	return bind.Axis()
}