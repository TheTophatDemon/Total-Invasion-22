package glfw

import (
	"fmt"
	"reflect"

	"github.com/ebitengine/purego"
)

type (
	Key                 int
	Action              int
	ModifierKey         int
	MouseButton         int
	Hint                int
	InputMode           int
	Monitor             struct{}
	Window              struct{}
	KeyCallback         func(w *Window, key Key, scancode int, action Action, mods ModifierKey)
	MouseButtonCallback func(w *Window, button MouseButton, action Action, mods ModifierKey)
	ScrollCallback      func(w *Window, xoff float64, yoff float64)
	ErrorCallback       func(code ErrorCode, description string)
)

const (
	CursorMode             InputMode = 0x00033001
	StickyKeysMode         InputMode = 0x00033002
	StickyMouseButtonsMode InputMode = 0x00033003
	LockKeyMods            InputMode = 0x00033004
	RawMouseMotion         InputMode = 0x00033005
)

const (
	CursorNormal   int = 0x00034001
	CursorHidden   int = 0x00034002
	CursorDisabled int = 0x00034003
)

const (
	False int = 0
	True  int = 1
)

const (
	Focused                 Hint = 0x00020001 // Specifies whether the window will be given input focus when created. This hint is ignored for full screen and initially hidden windows.
	Iconified               Hint = 0x00020002 // Specifies whether the window will be minimized.
	Maximized               Hint = 0x00020008 // Specifies whether the window is maximized.
	Visible                 Hint = 0x00020004 // Specifies whether the window will be initially visible.
	Hovered                 Hint = 0x0002000B // Specifies whether the cursor is currently directly over the content area of the window, with no other windows between. See Cursor enter/leave events for details.
	Resizable               Hint = 0x00020003 // Specifies whether the window will be resizable by the user.
	Decorated               Hint = 0x00020005 // Specifies whether the window will have window decorations such as a border, a close widget, etc.
	Floating                Hint = 0x00020007 // Specifies whether the window will be always-on-top.
	AutoIconify             Hint = 0x00020006 // Specifies whether fullscreen windows automatically iconify (and restore the previous video mode) on focus loss.
	CenterCursor            Hint = 0x00020009 // Specifies whether the cursor should be centered over newly created full screen windows. This hint is ignored for windowed mode windows.
	TransparentFramebuffer  Hint = 0x0002000A // Specifies whether the framebuffer should be transparent.
	FocusOnShow             Hint = 0x0002000C // Specifies whether the window will be given input focus when glfwShowWindow is called.
	ScaleToMonitor          Hint = 0x0002200C // Specified whether the window content area should be resized based on the monitor content scale of any monitor it is placed on. This includes the initial placement when the window is created.
	ContextVersionMajor     Hint = 0x00022002
	ContextVersionMinor     Hint = 0x00022003
	OpenGLProfile           Hint = 0x00022008
	OpenGLForwardCompatible Hint = 0x00022006
	OpenGLDebugContext      Hint = 0x00022007
)

const (
	OpenGLCoreProfile int = 0x00032001
)

const (
	Release Action = 0 // The key or button was released.
	Press   Action = 1 // The key or button was pressed.
	Repeat  Action = 2 // The key was held down until it repeated.
)

const (
	MouseButton1      MouseButton = 0
	MouseButton2      MouseButton = 1
	MouseButton3      MouseButton = 2
	MouseButton4      MouseButton = 3
	MouseButton5      MouseButton = 4
	MouseButton6      MouseButton = 5
	MouseButton7      MouseButton = 6
	MouseButton8      MouseButton = 7
	MouseButtonLast   MouseButton = MouseButton8
	MouseButtonLeft   MouseButton = MouseButton1
	MouseButtonRight  MouseButton = MouseButton2
	MouseButtonMiddle MouseButton = MouseButton3
)

const (
	KeyUnknown      Key = -1
	KeySpace        Key = 32
	KeyApostrophe   Key = 39
	KeyComma        Key = 44
	KeyMinus        Key = 45
	KeyPeriod       Key = 46
	KeySlash        Key = 47
	Key0            Key = 48
	Key1            Key = 49
	Key2            Key = 50
	Key3            Key = 51
	Key4            Key = 52
	Key5            Key = 53
	Key6            Key = 54
	Key7            Key = 55
	Key8            Key = 56
	Key9            Key = 57
	KeySemicolon    Key = 59
	KeyEqual        Key = 61
	KeyA            Key = 65
	KeyB            Key = 66
	KeyC            Key = 67
	KeyD            Key = 68
	KeyE            Key = 69
	KeyF            Key = 70
	KeyG            Key = 71
	KeyH            Key = 72
	KeyI            Key = 73
	KeyJ            Key = 74
	KeyK            Key = 75
	KeyL            Key = 76
	KeyM            Key = 77
	KeyN            Key = 78
	KeyO            Key = 79
	KeyP            Key = 80
	KeyQ            Key = 81
	KeyR            Key = 82
	KeyS            Key = 83
	KeyT            Key = 84
	KeyU            Key = 85
	KeyV            Key = 86
	KeyW            Key = 87
	KeyX            Key = 88
	KeyY            Key = 89
	KeyZ            Key = 90
	KeyLeftBracket  Key = 91
	KeyBackslash    Key = 92
	KeyRightBracket Key = 93
	KeyGraveAccent  Key = 96
	KeyWorld1       Key = 161
	KeyWorld2       Key = 162
	KeyEscape       Key = 256
	KeyEnter        Key = 257
	KeyTab          Key = 258
	KeyBackspace    Key = 259
	KeyInsert       Key = 260
	KeyDelete       Key = 261
	KeyRight        Key = 262
	KeyLeft         Key = 263
	KeyDown         Key = 264
	KeyUp           Key = 265
	KeyPageUp       Key = 266
	KeyPageDown     Key = 267
	KeyHome         Key = 268
	KeyEnd          Key = 269
	KeyCapsLock     Key = 280
	KeyScrollLock   Key = 281
	KeyNumLock      Key = 282
	KeyPrintScreen  Key = 283
	KeyPause        Key = 284
	KeyF1           Key = 290
	KeyF2           Key = 291
	KeyF3           Key = 292
	KeyF4           Key = 293
	KeyF5           Key = 294
	KeyF6           Key = 295
	KeyF7           Key = 296
	KeyF8           Key = 297
	KeyF9           Key = 298
	KeyF10          Key = 299
	KeyF11          Key = 300
	KeyF12          Key = 301
	KeyF13          Key = 302
	KeyF14          Key = 303
	KeyF15          Key = 304
	KeyF16          Key = 305
	KeyF17          Key = 306
	KeyF18          Key = 307
	KeyF19          Key = 308
	KeyF20          Key = 309
	KeyF21          Key = 310
	KeyF22          Key = 311
	KeyF23          Key = 312
	KeyF24          Key = 313
	KeyF25          Key = 314
	KeyKP0          Key = 320
	KeyKP1          Key = 321
	KeyKP2          Key = 322
	KeyKP3          Key = 323
	KeyKP4          Key = 324
	KeyKP5          Key = 325
	KeyKP6          Key = 326
	KeyKP7          Key = 327
	KeyKP8          Key = 328
	KeyKP9          Key = 329
	KeyKPDecimal    Key = 330
	KeyKPDivide     Key = 331
	KeyKPMultiply   Key = 332
	KeyKPSubtract   Key = 333
	KeyKPAdd        Key = 334
	KeyKPEnter      Key = 335
	KeyKPEqual      Key = 336
	KeyLeftShift    Key = 340
	KeyLeftControl  Key = 341
	KeyLeftAlt      Key = 342
	KeyLeftSuper    Key = 343
	KeyRightShift   Key = 344
	KeyRightControl Key = 345
	KeyRightAlt     Key = 346
	KeyRightSuper   Key = 347
	KeyMenu         Key = 348
	KeyLast         Key = KeyMenu
)

const (
	ModShift    ModifierKey = 0x0001
	ModControl  ModifierKey = 0x0002
	ModAlt      ModifierKey = 0x0004
	ModSuper    ModifierKey = 0x0008
	ModCapsLock ModifierKey = 0x0010
	ModNumLock  ModifierKey = 0x0020
)

var glfwFuncs struct {
	SetErrorCallback       func(callback uintptr)
	PollEvents             func()
	Terminate              func()
	SwapInterval           func(interval int)
	GetKeyName             func(key Key, scancode int) string
	WindowHint             func(target Hint, hint int)
	GetPrimaryMonitor      func() *Monitor
	GetMonitorWorkarea     func(monitor *Monitor, xpos *int, ypos *int, width *int, height *int)
	CreateWindow           func(width, height int, title string, monitor *Monitor, share *Window) *Window
	GetCurrentContext      func() *Window
	SetKeyCallback         func(window *Window, callback uintptr) KeyCallback
	SetMouseButtonCallback func(window *Window, callback uintptr) MouseButtonCallback
	SetScrollCallback      func(window *Window, callback uintptr) ScrollCallback
	GetCursorPos           func(window *Window, xpos *float64, ypos *float64)
	SetCursorPos           func(window *Window, xpos float64, ypos float64)
	GetInputMode           func(window *Window, mode InputMode) int
	SetInputMode           func(window *Window, mode InputMode, value int)
	GetWindowAttrib        func(window *Window, attrib Hint) int
	GetKey                 func(window *Window, key Key) Action
	GetMouseButton         func(window *Window, button MouseButton) Action
	GetWindowMonitor       func(window *Window) *Monitor
	SetWindowMonitor       func(window *Window, monitor *Monitor, xpos, ypos, width, height, refreshRate int)
	GetWindowPos           func(window *Window, xpos, ypos *int)
	SetWindowPos           func(window *Window, xpos, ypos int)
	GetWindowSize          func(window *Window, width, height *int)
	SetWindowSize          func(window *Window, width, height int)
	MakeContextCurrent     func(window *Window)
	SwapBuffers            func(window *Window)
	WindowShouldClose      func(window *Window) bool
	SetWindowShouldClose   func(window *Window, shouldClose bool)
}

func PollEvents() {
	glfwFuncs.PollEvents()
	panicError()
}

func Terminate() {
	glfwFuncs.Terminate()
	panicError()
}

func SwapInterval(interval int) {
	glfwFuncs.SwapInterval(interval)
	panicError()
}

func GetKeyName(key Key, scancode int) string {
	res := glfwFuncs.GetKeyName(key, scancode)
	panicError()
	return res
}

func WindowHint(target Hint, hint int) {
	glfwFuncs.WindowHint(target, hint)
	panicError()
}

func GetPrimaryMonitor() *Monitor {
	res := glfwFuncs.GetPrimaryMonitor()
	panicError()
	return res
}

func (monitor *Monitor) GetWorkarea() (x, y, width, height int) {
	glfwFuncs.GetMonitorWorkarea(monitor, &x, &y, &width, &height)
	panicError()
	return
}

func CreateWindow(width, height int, title string, monitor *Monitor, share *Window) (*Window, error) {
	window := glfwFuncs.CreateWindow(width, height, title, monitor, share)
	if window == nil {
		return nil, acceptError(APIUnavailable, VersionUnavailable)
	}
	return window, nil
}

func GetCurrentContext() *Window {
	res := glfwFuncs.GetCurrentContext()
	panicError()
	return res
}

func (window *Window) SetKeyCallback(callback KeyCallback) KeyCallback {
	cCallback := purego.NewCallback(func(w *Window, key Key, scancode int, action Action, mods ModifierKey) int {
		callback(w, key, scancode, action, mods)
		return 1
	})
	res := glfwFuncs.SetKeyCallback(window, cCallback)
	panicError()
	return res
}

func (window *Window) SetMouseButtonCallback(callback MouseButtonCallback) MouseButtonCallback {
	cCallback := purego.NewCallback(func(w *Window, button MouseButton, action Action, mods ModifierKey) int {
		callback(w, button, action, mods)
		return 1
	})
	res := glfwFuncs.SetMouseButtonCallback(window, cCallback)
	panicError()
	return res
}

func (window *Window) SetScrollCallback(callback ScrollCallback) ScrollCallback {
	cCallback := purego.NewCallback(func(w *Window, xoff int64, yoff int64) int {
		callback(w, float64(xoff), float64(yoff))
		return 1
	})
	res := glfwFuncs.SetScrollCallback(window, cCallback)
	panicError()
	return res
}

func (window *Window) GetCursorPos() (xpos float64, ypos float64) {
	glfwFuncs.GetCursorPos(window, &xpos, &ypos)
	panicError()
	return
}

func (window *Window) SetCursorPos(xpos float64, ypos float64) {
	glfwFuncs.SetCursorPos(window, xpos, ypos)
	panicError()
}

func (window *Window) GetInputMode(mode InputMode) int {
	res := glfwFuncs.GetInputMode(window, mode)
	panicError()
	return res
}

func (window *Window) SetInputMode(mode InputMode, value int) {
	glfwFuncs.SetInputMode(window, mode, value)
	panicError()
}

func (window *Window) GetAttrib(attrib Hint) int {
	res := glfwFuncs.GetWindowAttrib(window, attrib)
	panicError()
	return res
}

func (window *Window) GetKey(key Key) Action {
	res := glfwFuncs.GetKey(window, key)
	panicError()
	return res
}

func (window *Window) GetMouseButton(button MouseButton) Action {
	res := glfwFuncs.GetMouseButton(window, button)
	panicError()
	return res
}

func (window *Window) GetMonitor() *Monitor {
	res := glfwFuncs.GetWindowMonitor(window)
	panicError()
	return res
}

func (window *Window) SetMonitor(monitor *Monitor, xpos, ypos, width, height, refreshRate int) {
	glfwFuncs.SetWindowMonitor(window, monitor, xpos, ypos, width, height, refreshRate)
	panicError()
}

func (window *Window) GetPos() (x, y int) {
	glfwFuncs.GetWindowPos(window, &x, &y)
	panicError()
	return
}

func (window *Window) SetPos(x, y int) {
	glfwFuncs.SetWindowPos(window, x, y)
	panicError()
}

func (window *Window) GetSize() (width, height int) {
	glfwFuncs.GetWindowSize(window, &width, &height)
	panicError()
	return
}

func (window *Window) SetSize(width, height int) {
	glfwFuncs.SetWindowSize(window, width, height)
	panicError()
}

func (window *Window) MakeContextCurrent() {
	glfwFuncs.MakeContextCurrent(window)
	panicError()
}

func (window *Window) SwapBuffers() {
	glfwFuncs.SwapBuffers(window)
	panicError()
}

func (window *Window) ShouldClose() bool {
	res := glfwFuncs.WindowShouldClose(window)
	panicError()
	return res
}

func (window *Window) SetShouldClose(shouldClose bool) {
	glfwFuncs.SetWindowShouldClose(window, shouldClose)
	panicError()
}

func Init() error {
	lib, err := openLibrary()
	if err != nil {
		return fmt.Errorf("error loading glfw dll: %v", err)
	}

	var init func() bool
	purego.RegisterLibFunc(&init, lib, "glfwInit")

	structVal := reflect.ValueOf(&glfwFuncs).Elem()
	for field, fieldVal := range structVal.Fields() {
		purego.RegisterLibFunc(fieldVal.Addr().Interface(), lib, "glfw"+field.Name)
	}

	setErrorCallback()
	init()
	return acceptError(APIUnavailable)
}
