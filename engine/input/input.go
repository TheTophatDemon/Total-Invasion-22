package input

import (
	"github.com/go-gl/mathgl/mgl32"
	"tophatdemon.com/total-invasion-ii/engine/rgfw"
)

type (
	InputCaptureCallback func(newBinding Binding, extraData any)
)

var anythingPressed bool
var wasAnyMouseButtonPressed bool

// Holds the callback that assigns the captured input to a binding
var captureCallback InputCaptureCallback = nil

// Extra data for the capture callback
var captureData any = nil

// Holds the binding of the input action that was captured
var capturedBinding Binding = nil

var lastKeyPressed rgfw.Key = rgfw.KEY_NULL
var inputFrameNumber uint64
var mouseDeltaX, mouseDeltaY float32
var mouseScrollY float32

var window *rgfw.Window

func Init(win *rgfw.Window) {
	window = win
}

func PostUpdate() {
	inputFrameNumber += 1
	anythingPressed = false
	wasAnyMouseButtonPressed = AnyMouseButtonPressed()
	lastKeyPressed = rgfw.KEY_NULL
	mouseScrollY = 0.0
	mouseDeltaX, mouseDeltaY = 0, 0
	for {
		event, hasEvent := window.CheckEvent()
		if !hasEvent {
			break
		}
		switch event.Type {
		case rgfw.EVENT_KEY_PRESSED:
			keyEvent, _ := event.Key()
			if IsCapturingInput() {
				capturedBinding = NewKeyBinding(keyEvent.Value)
			} else {
				anythingPressed = true
				lastKeyPressed = keyEvent.Value
			}
		case rgfw.EVENT_MOUSE_BUTTON_PRESSED:
			mouseEvent, _ := event.MouseButton()
			if IsCapturingInput() {
				capturedBinding = NewMouseButtonBinding(mouseEvent.Value)
			} else {
				anythingPressed = true
			}
		case rgfw.EVENT_KEY_RELEASED, rgfw.EVENT_MOUSE_BUTTON_RELEASED:
			if IsCapturingInput() && capturedBinding != nil {
				endCaptureInput()
			}
		case rgfw.EVENT_MOUSE_SCROLL:
			mouseEvent, _ := event.MouseDelta()
			mouseScrollY = mouseEvent.Y
		case rgfw.EVENT_MOUSE_RAW_MOTION:
			mouseEvent, _ := event.MouseDelta()
			mouseDeltaX += mouseEvent.X
			mouseDeltaY += mouseEvent.Y
		}
	}

	// Capture mouse movement bindings
	movedAxis := MouseAxisNone
	const moveThreshold = 5.0
	switch true {
	case mouseDeltaX > moveThreshold:
		movedAxis = MouseAxisPosX
	case mouseDeltaX < -moveThreshold:
		movedAxis = MouseAxisNegX
	case mouseDeltaY > moveThreshold:
		movedAxis = MouseAxisPosY
	case mouseDeltaY < -moveThreshold:
		movedAxis = MouseAxisNegY
	}
	if movedAxis != MouseAxisNone && IsCapturingInput() {
		capturedBinding = NewMouseMovementBinding(movedAxis)
		endCaptureInput()
	}
}

func TrapMouse() {
	window.CaptureRawMouse(true)
	window.SetMouseVisibility(false)
}

func UntrapMouse() {
	window.CaptureRawMouse(false)
	window.SetMouseVisibility(true)
}

func IsMouseTrapped() bool {
	return window.IsMouseCaptured()
}

func CaptureInput(callback InputCaptureCallback, extraData any) {
	captureCallback = callback
	captureData = extraData
}

func IsCapturingInput() bool {
	return captureCallback != nil
}

func CaptureExtraData() any {
	if captureCallback != nil {
		return captureData
	}
	return nil
}

func endCaptureInput() {
	if captureCallback != nil {
		captureCallback(capturedBinding, captureData)
		captureCallback = nil
		captureData = nil
		capturedBinding = nil
	}
}

func MousePosition() mgl32.Vec2 {
	if !window.IsInFocus() {
		return mgl32.Vec2{-10000.0, -10000.0}
	}
	x, y := window.GetMousePosition()
	return mgl32.Vec2{float32(x), float32(y)}
}

func MouseDelta() mgl32.Vec2 {
	return mgl32.Vec2{float32(mouseDeltaX), float32(mouseDeltaY)}
}

func MouseScrollDelta() float32 {
	return mouseScrollY
}

func IsMouseButtonDown(button rgfw.MouseButton) bool {
	return window.IsMouseDown(button) && !IsCapturingInput()
}

func AnyMouseButtonPressed() bool {
	return IsMouseButtonDown(rgfw.MOUSE_BUTTON_LEFT) || IsMouseButtonDown(rgfw.MOUSE_BUTTON_RIGHT) || IsMouseButtonDown(rgfw.MOUSE_BUTTON_MIDDLE)
}

func AnyMouseButtonJustPressed() bool {
	return AnyMouseButtonPressed() && !wasAnyMouseButtonPressed && !IsCapturingInput()
}

func IsAnythingPressed() bool {
	return anythingPressed && !IsCapturingInput()
}
