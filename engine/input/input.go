package input

import (
	"runtime"
	"unsafe"

	"github.com/ebitengine/purego"
	"github.com/go-gl/mathgl/mgl32"
	"tophatdemon.com/total-invasion-ii/engine/evil"
	"tophatdemon.com/total-invasion-ii/engine/glfw"
)

type (
	InputCaptureCallback func(newBinding *Binding, extraData uintptr)
)

var (
	inputInit                      func()
	inputPostUpdate                func()
	inputTrapMouse                 func()
	inputUntrapMouse               func()
	inputIsMouseTrapped            func() bool
	inputCaptureInput              func(callback uintptr, extraData uintptr)
	inputIsCapturingInput          func() bool
	inputCaptureExtraData          func() uintptr
	inputMousePosition             func() [2]float32
	inputMouseDelta                func() [2]float32
	inputMouseScroll               func() float32
	inputMouseScrollDelta          func() float32
	inputSetMousePosition          func(x, y float32)
	inputIsMouseButtonDown         func(button glfw.MouseButton) bool
	inputAnyMouseButtonPressed     func() bool
	inputAnyMouseButtonJustPressed func() bool
	inputIsAnythingPressed         func() bool
	inputBindingPressed            func(binding *Binding) bool
	inputBindingJustPressed        func(binding *Binding) bool
	inputBindingJustReleased       func(binding *Binding) bool
	inputBindingAxis               func(binding *Binding) float32
	inputNewKeyBinding             func(key glfw.Key) Binding
	inputNewMouseButtonBinding     func(mouseButton glfw.MouseButton) Binding
	inputNewMouseMovementBinding   func(axis MouseAxis) Binding
	inputNewCharSequenceBinding    func(keys *glfw.Key, count int) Binding
)

func Init() {
	lib := evil.Library()
	purego.RegisterLibFunc(&inputInit, lib, "input_init")
	purego.RegisterLibFunc(&inputPostUpdate, lib, "input_post_update")
	purego.RegisterLibFunc(&inputTrapMouse, lib, "input_trap_mouse")
	purego.RegisterLibFunc(&inputUntrapMouse, lib, "input_untrap_mouse")
	purego.RegisterLibFunc(&inputIsMouseTrapped, lib, "input_is_mouse_trapped")
	purego.RegisterLibFunc(&inputCaptureInput, lib, "input_capture_input")
	purego.RegisterLibFunc(&inputIsCapturingInput, lib, "input_is_capturing_input")
	purego.RegisterLibFunc(&inputCaptureExtraData, lib, "input_capture_extra_data")
	purego.RegisterLibFunc(&inputMousePosition, lib, "input_mouse_position")
	purego.RegisterLibFunc(&inputMouseDelta, lib, "input_mouse_delta")
	purego.RegisterLibFunc(&inputMouseScroll, lib, "input_mouse_scroll")
	purego.RegisterLibFunc(&inputMouseScrollDelta, lib, "input_mouse_scroll_delta")
	purego.RegisterLibFunc(&inputSetMousePosition, lib, "input_set_mouse_position")
	purego.RegisterLibFunc(&inputIsMouseButtonDown, lib, "input_is_mouse_button_down")
	purego.RegisterLibFunc(&inputAnyMouseButtonPressed, lib, "input_any_mouse_button_pressed")
	purego.RegisterLibFunc(&inputAnyMouseButtonJustPressed, lib, "input_any_mouse_button_just_pressed")
	purego.RegisterLibFunc(&inputIsAnythingPressed, lib, "input_is_anything_pressed")
	purego.RegisterLibFunc(&inputBindingPressed, lib, "input_binding_pressed")
	purego.RegisterLibFunc(&inputBindingJustPressed, lib, "input_binding_just_pressed")
	purego.RegisterLibFunc(&inputBindingJustReleased, lib, "input_binding_just_released")
	purego.RegisterLibFunc(&inputBindingAxis, lib, "input_binding_axis")
	purego.RegisterLibFunc(&inputNewKeyBinding, lib, "input_new_key_binding")
	purego.RegisterLibFunc(&inputNewMouseButtonBinding, lib, "input_new_mouse_button_binding")
	purego.RegisterLibFunc(&inputNewMouseMovementBinding, lib, "input_new_mouse_movement_binding")
	purego.RegisterLibFunc(&inputNewCharSequenceBinding, lib, "input_new_char_sequence_binding")

	inputInit()
}

func PostUpdate() {
	inputPostUpdate()
}

func TrapMouse() {
	inputTrapMouse()
}

func UntrapMouse() {
	inputUntrapMouse()
}

func IsMouseTrapped() bool {
	return inputIsMouseTrapped()
}

func CaptureInput(callback InputCaptureCallback, extraData unsafe.Pointer) {
	var pin runtime.Pinner
	pin.Pin(extraData)
	pureCallback := purego.NewCallback(func(newBinding *Binding, extraData uintptr) int {
		callback(newBinding, extraData)
		pin.Unpin()
		return 1
	})
	inputCaptureInput(pureCallback, uintptr(extraData))
}

func IsCapturingInput() bool {
	return inputIsCapturingInput()
}

func CaptureExtraData() any {
	return inputCaptureExtraData()
}

func MousePosition() mgl32.Vec2 {
	return mgl32.Vec2(inputMousePosition())
}

func MouseDelta() mgl32.Vec2 {
	return mgl32.Vec2(inputMouseDelta())
}

func MouseScroll() float32 {
	return inputMouseScroll()
}

func MouseScrollDelta() float32 {
	return inputMouseScrollDelta()
}

func SetMousePosition(x, y float32) {
	inputSetMousePosition(x, y)
}

func IsMouseButtonDown(button glfw.MouseButton) bool {
	return inputIsMouseButtonDown(button)
}

func AnyMouseButtonPressed() bool {
	return inputAnyMouseButtonPressed()
}

func AnyMouseButtonJustPressed() bool {
	return inputAnyMouseButtonJustPressed()
}

func IsAnythingPressed() bool {
	return inputIsAnythingPressed()
}
