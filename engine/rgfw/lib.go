package rgfw

import (
	"fmt"
	"structs"
	"unsafe"

	"github.com/ebitengine/purego"
	"tophatdemon.com/total-invasion-ii/engine/math2"
)

type InitFlags uint8

const (
	INIT_OPENGL InitFlags = 1 << iota
	INIT_EGL
	INIT_VULKAN
	INIT_X11
)

type ErrorCode uint8

const (
	NO_ERROR ErrorCode = iota
	ERROR_OUT_OF_MEMORY
	ERROR_OPEN_GL_CONTEXT
	ERROR_EGL_CONTEXT
	ERROR_WAYLAND
	ERROR_X11
	ERROR_DIRECT_X_CONTEXT
	ERROR_IO_KIT
	ERROR_CLIPBOARD
	ERROR_FAILED_FUNC_LOAD
	ERROR_BUFFER
	ERROR_METAL
	ERROR_PLATFORM
	ERROR_EVENT_QUEUE
	ERROR_NO_INIT
	INFO_WINDOW
	INFO_BUFFER
	INFO_GLOBAL
	INFO_OPEN_GL
	WARNING_WAYLAND
	WARNING_OPEN_GL
)

func Error(result int32) error {
	switch ErrorCode(math2.Abs(result)) {
	case ERROR_OUT_OF_MEMORY:
		return fmt.Errorf("ERROR_OUT_OF_MEMORY")
	case ERROR_OPEN_GL_CONTEXT:
		return fmt.Errorf("ERROR_OPEN_GL_CONTEXT")
	case ERROR_EGL_CONTEXT:
		return fmt.Errorf("ERROR_EGL_CONTEXT")
	case ERROR_WAYLAND:
		return fmt.Errorf("ERROR_WAYLAND")
	case ERROR_X11:
		return fmt.Errorf("ERROR_X11")
	case ERROR_DIRECT_X_CONTEXT:
		return fmt.Errorf("ERROR_DIRECT_X_CONTEXT")
	case ERROR_IO_KIT:
		return fmt.Errorf("ERROR_IO_KIT")
	case ERROR_CLIPBOARD:
		return fmt.Errorf("ERROR_CLIPBOARD")
	case ERROR_FAILED_FUNC_LOAD:
		return fmt.Errorf("ERROR_FAILED_FUNC_LOAD")
	case ERROR_BUFFER:
		return fmt.Errorf("ERROR_BUFFER")
	case ERROR_METAL:
		return fmt.Errorf("ERROR_METAL")
	case ERROR_PLATFORM:
		return fmt.Errorf("ERROR_PLATFORM")
	case ERROR_EVENT_QUEUE:
		return fmt.Errorf("ERROR_EVENT_QUEUE")
	case ERROR_NO_INIT:
		return fmt.Errorf("ERROR_NO_INIT")
	case INFO_WINDOW:
		return fmt.Errorf("INFO_WINDOW")
	case INFO_BUFFER:
		return fmt.Errorf("INFO_BUFFER")
	case INFO_GLOBAL:
		return fmt.Errorf("INFO_GLOBAL")
	case INFO_OPEN_GL:
		return fmt.Errorf("INFO_OPEN_GL")
	case WARNING_WAYLAND:
		return fmt.Errorf("WARNING_WAYLAND")
	case WARNING_OPEN_GL:
		return fmt.Errorf("WARNING_OPEN_GL")
	default:
		return nil
	}
}

type WindowFlags uint32

const (
	WINDOW_NO_BORDER WindowFlags = 1 << iota
	WINDOW_NO_RESIZE
	WINDOW_ALLOW_DND
	WINDOW_HIDE_MOUSE
	WINDOW_FULLSCREEN_EXCLUSIVE
	WINDOW_TRANSLUCENT
	WINDOW_CENTER
	WINDOW_RAW_MOUSE
	WINDOW_SCALE_TO_MONITOR
	WINDOW_HIDE
	WINDOW_MAXIMIZE
	WINDOW_CENTER_CURSOR
	WINDOW_FLOATING
	WINDOW_FOCUS_ON_SHOW
	WINDOW_MINIMIZE
	WINDOW_FOCUS
	WINDOW_CAPTURE_MOUSE
	WINDOW_OPEN_GL
	WINDOW_EGL
	WINDOW_FULLSCREEN_BORDERLESS
	WINDOW_CAPTURE_RAW_MOUSE = WINDOW_CAPTURE_MOUSE | WINDOW_RAW_MOUSE
)

type Window struct{}

type OpenGLContext struct{}

type GLReleaseBehavior int32

const (
	GL_RELEASE_FLUSH GLReleaseBehavior = iota
	GL_RELEASE_NONE
)

type GLProfile int32

const (
	GL_CORE GLProfile = iota
	GL_FORWARD_COMPATIBILITY
	GL_COMPATIBILITY
	GL_ES
	GL_WEB
)

type GLRenderer int32

const (
	GL_ACCELERATED GLRenderer = iota
	GL_SOFTWARE
)

type GLHints struct {
	_                                                                                  structs.HostLayout
	Stencil, Samples, Stereo, AuxBuffers, DoubleBuffer, Red, Green, Blue, Alpha, Depth int32
	AccumRed, AccumGreen, AccumBlue, AccumAlpha                                        int32
	SRGB, Robustness, Debug, NoError                                                   bool
	ReleaseBehavior                                                                    GLReleaseBehavior
	Profile                                                                            GLProfile
	MajorVersion, MinorVersion                                                         int32
	Share                                                                              *OpenGLContext
	shareEgl                                                                           uintptr
	Renderer                                                                           GLRenderer
}

type MouseButton uint8

const (
	MOUSE_BUTTON_LEFT MouseButton = iota
	MOUSE_BUTTON_MIDDLE
	MOUSE_BUTTON_RIGHT
	MOUSE_BUTTON_MISC_1
	MOUSE_BUTTON_MISC_2
	MOUSE_BUTTON_MISC_3
	MOUSE_BUTTON_MISC_4
	MOUSE_BUTTON_MISC_5
	MOUSE_BUTTON_FINAL
)

type EventType uint8

const (
	EVENT_NONE EventType = iota
	EVENT_KEY_PRESSED
	EVENT_KEY_RELEASED
	EVENT_KEY_CHAR
	EVENT_MOUSE_BUTTON_PRESSED
	EVENT_MOUSE_BUTTON_RELEASED
	EVENT_MOUSE_SCROLL
	EVENT_MOUSE_MOTION
	EVENT_MOUSE_RAW_MOTION
	EVENT_MOUSE_ENTER
	EVENT_MOUSE_LEAVE
	EVENT_WINDOW_MOVED
	EVENT_WINDOW_RESIZED
	EVENT_WINDOW_FOCUS_IN
	EVENT_WINDOW_FOCUS_OUT
	EVENT_WINDOW_REFRESH
	EVENT_WINDOW_CLOSE
	EVENT_WINDOW_MAXIMIZED
	EVENT_WINDOW_MINIMIZED
	EVENT_WINDOW_RESTORED
	EVENT_DATA_DROP
	EVENT_DATA_DRAG
	EVENT_SCALE_UPDATED
	EVENT_MONITOR_CONNECTED
	EVENT_MONITOR_DISCONNECTED
	EVENT_COUNT
	EVENT_MOUSE_POS_CHANGED = EVENT_MOUSE_MOTION
)

type Event struct {
	_       structs.HostLayout
	Type    EventType
	Window  *Window
	padding [23]byte // Union is 32 bytes large
}

type Key uint8

const (
	KEY_NULL          Key = 0
	KEY_ESCAPE            = 27
	KEY_BACKTICK          = Key('`')
	KEY0                  = Key('0')
	KEY1                  = Key('1')
	KEY2                  = Key('2')
	KEY3                  = Key('3')
	KEY4                  = Key('4')
	KEY5                  = Key('5')
	KEY6                  = Key('6')
	KEY7                  = Key('7')
	KEY8                  = Key('8')
	KEY9                  = Key('9')
	KEY_MINUS             = Key('-')
	KEY_EQUAL             = Key('=')
	KEY_EQUALS            = KEY_EQUAL
	KEY_BACK_SPACE        = Key('\b')
	KEY_TAB               = Key('\t')
	KEY_SPACE             = Key(' ')
	KEY_A                 = Key('a')
	KEY_B                 = Key('b')
	KEY_C                 = Key('c')
	KEY_D                 = Key('d')
	KEY_E                 = Key('e')
	KEY_F                 = Key('f')
	KEY_G                 = Key('g')
	KEY_H                 = Key('h')
	KEY_I                 = Key('i')
	KEY_J                 = Key('j')
	KEY_K                 = Key('k')
	KEY_L                 = Key('l')
	KEY_M                 = Key('m')
	KEY_N                 = Key('n')
	KEY_O                 = Key('o')
	KEY_P                 = Key('p')
	KEY_Q                 = Key('q')
	KEY_R                 = Key('r')
	KEY_S                 = Key('s')
	KEY_T                 = Key('t')
	KEY_U                 = Key('u')
	KEY_V                 = Key('v')
	KEY_W                 = Key('w')
	KEY_X                 = Key('x')
	KEY_Y                 = Key('y')
	KEY_Z                 = Key('z')
	KEY_PERIOD            = Key('.')
	KEY_COMMA             = Key(',')
	KEY_SLASH             = Key('/')
	KEY_BRACKET           = Key('[')
	KEY_CLOSE_BRACKET     = Key(']')
	KEY_SEMICOLON         = Key(';')
	KEY_APOSTROPHE        = Key('\'')
	KEY_BACK_SLASH        = Key('\\')
	KEY_RETURN            = Key('\n')
	KEY_ENTER             = KEY_RETURN
	KEY_DELETE            = 127
)

const (
	KEY_F1 Key = 128 + iota
	KEY_F2
	KEY_F3
	KEY_F4
	KEY_F5
	KEY_F6
	KEY_F7
	KEY_F8
	KEY_F9
	KEY_F10
	KEY_F11
	KEY_F12
	KEY_F13
	KEY_F14
	KEY_F15
	KEY_F16
	KEY_F17
	KEY_F18
	KEY_F19
	KEY_F20
	KEY_F21
	KEY_F22
	KEY_F23
	KEY_F24
	KEY_F25
	KEY_CAPS_LOCK
	KEY_SHIFT_L
	KEY_CONTROL_L
	KEY_ALT_L
	KEY_SUPER_L
	KEY_SHIFT_R
	KEY_CONTROL_R
	KEY_ALT_R
	KEY_SUPER_R
	KEY_UP
	KEY_DOWN
	KEY_LEFT
	KEY_RIGHT
	KEY_INSERT
	KEY_MENU
	KEY_END
	KEY_HOME
	KEY_PAGE_UP
	KEY_PAGE_DOWN
	KEY_NUM_LOCK
	KEY_PAD_SLASH
	KEY_PAD_MULTIPLY
	KEY_PAD_PLUS
	KEY_PAD_MINUS
	KEY_PAD_EQUAL
	KEY_PAD_EQUALS = KEY_PAD_EQUAL
)

const (
	KEY_PAD1 = 178 + iota
	KEY_PAD2
	KEY_PAD3
	KEY_PAD4
	KEY_PAD5
	KEY_PAD6
	KEY_PAD7
	KEY_PAD8
	KEY_PAD9
	KEY_PAD0
	KEY_PAD_PERIOD
	KEY_PAD_RETURN
	KEY_SCROLL_LOCK
	KEY_PRINT_SCREEN
	KEY_PAUSE
	KEY_WORLD1
	KEY_WORLD2
	KEY_LAST = 256
)

type KeyMod uint8

const (
	KEY_MOD_CAPS_LOCK KeyMod = iota
	KEY_MOD_NUM_LOCK
	KEY_MOD_CONTROL
	KEY_MOD_ALT
	KEY_MOD_SHIFT
	KEY_MOD_SUPER
	KEY_MOD_SCROLL_LOCK
)

type KeyEvent struct {
	_      structs.HostLayout
	Type   EventType
	Window *Window
	Value  Key
	Repeat bool
	Mod    KeyMod
	State  bool
}

func (event Event) Key() (keyEvent KeyEvent, ok bool) {
	if event.Type != EVENT_KEY_PRESSED && event.Type != EVENT_KEY_RELEASED {
		return
	}
	keyEvent = *(*KeyEvent)(unsafe.Pointer(&event))
	ok = true
	return
}

type KeyCharEvent struct {
	_      structs.HostLayout
	Type   EventType
	Window *Window
	Value  uint32
}

func (event Event) KeyChar() (keyEvent KeyCharEvent, ok bool) {
	if event.Type != EVENT_KEY_CHAR {
		return
	}
	keyEvent = *(*KeyCharEvent)(unsafe.Pointer(&event))
	ok = true
	return
}

type MouseButtonEvent struct {
	_      structs.HostLayout
	Type   EventType
	Window *Window
	Value  MouseButton
	State  bool
}

func (event Event) MouseButton() (mouseEvent MouseButtonEvent, ok bool) {
	if event.Type != EVENT_MOUSE_BUTTON_PRESSED && event.Type != EVENT_MOUSE_BUTTON_RELEASED {
		return
	}
	mouseEvent = *(*MouseButtonEvent)(unsafe.Pointer(&event))
	ok = true
	return
}

type MouseDeltaEvent struct {
	_      structs.HostLayout
	Type   EventType
	Window *Window
	X, Y   float32
}

func (event Event) MouseDelta() (mouseEvent MouseDeltaEvent, ok bool) {
	if event.Type != EVENT_MOUSE_MOTION && event.Type != EVENT_MOUSE_RAW_MOTION && event.Type != EVENT_MOUSE_SCROLL {
		return
	}
	mouseEvent = *(*MouseDeltaEvent)(unsafe.Pointer(&event))
	ok = true
	return
}

type FullscreenMode uint8

const (
	FULLSCREEN_NONE FullscreenMode = iota
	FULLSCREEN_BORDERLESS
	FULLSCREEN_EXCLUSIVE
)

var (
	rgfwInit                           func(className string, flags InitFlags) int32
	rgfwDeinit                         func()
	rgfwPollEvents                     func()
	rgfwGetMouseScroll                 func(x, y *float32)
	rgfwCreateContextOpenGL            func(hints *GLHints) *OpenGLContext
	rgfwCreateWindow                   func(title string, x, y, w, h int32, flags WindowFlags) *Window
	rgfwCloseWindow                    func(window *Window)
	rgfwWindowShow                     func(window *Window)
	rgfwWindowShouldClose              func(window *Window) bool
	rgfwWindowSetShouldClose           func(window *Window, shouldClose bool)
	rgfwWindowIsKeyDown                func(window *Window, key Key) bool
	rgfwWindowGetMouse                 func(window *Window, x, y *int32) bool
	rgfwWindowIsMouseDown              func(window *Window, button MouseButton) bool
	rgfwWindowIsInFocus                func(window *Window) bool
	rgfwWindowCaptureMouse             func(window *Window, capture bool)
	rgfwWindowIsCaptured               func(window *Window) bool
	rgfwWindowSwapBuffersOpenGL        func(window *Window)
	rgfwWindowMoveMouse                func(window *Window, x, y int32)
	rgfwWindowCheckEvent               func(window *Window, event *Event) bool
	rgfwWindowSetFullscreen            func(window *Window, mode FullscreenMode)
	rgfwWindowIsFullscreen             func(window *Window) bool
	rgfwWindowSwapInterval             func(window *Window, interval int32)
	rgfwWindowResize                   func(window *Window, w, h int32)
	rgfwWindowGetSize                  func(window *Window, w, h *int32) bool
	rgfwWindowCenter                   func(window *Window)
	rgfwWindowMakeCurrentContextOpenGL func(window *Window)
	rgfwWindowSetContextOpenGL         func(window *Window, context *OpenGLContext) bool
)

func Init(className string, initFlags InitFlags) error {
	lib, err := openLibrary()
	if err != nil {
		return err
	}
	purego.RegisterLibFunc(&rgfwInit, lib, "RGFW_init")
	purego.RegisterLibFunc(&rgfwDeinit, lib, "RGFW_deinit")
	purego.RegisterLibFunc(&rgfwPollEvents, lib, "RGFW_pollEvents")
	purego.RegisterLibFunc(&rgfwCreateContextOpenGL, lib, "RGFW_createContext_OpenGL")
	purego.RegisterLibFunc(&rgfwCreateWindow, lib, "RGFW_createWindow")
	purego.RegisterLibFunc(&rgfwCloseWindow, lib, "RGFW_window_close")
	purego.RegisterLibFunc(&rgfwWindowShow, lib, "RGFW_window_show")
	purego.RegisterLibFunc(&rgfwWindowShouldClose, lib, "RGFW_window_shouldClose")
	purego.RegisterLibFunc(&rgfwWindowSetShouldClose, lib, "RGFW_window_setShouldClose")
	purego.RegisterLibFunc(&rgfwWindowIsKeyDown, lib, "RGFW_window_isKeyDown")
	purego.RegisterLibFunc(&rgfwWindowGetMouse, lib, "RGFW_window_getMouse")
	purego.RegisterLibFunc(&rgfwGetMouseScroll, lib, "RGFW_getMouseScroll")
	purego.RegisterLibFunc(&rgfwWindowIsMouseDown, lib, "RGFW_window_isMouseDown")
	purego.RegisterLibFunc(&rgfwWindowIsInFocus, lib, "RGFW_window_isInFocus")
	purego.RegisterLibFunc(&rgfwWindowCaptureMouse, lib, "RGFW_window_captureMouse")
	purego.RegisterLibFunc(&rgfwWindowIsCaptured, lib, "RGFW_window_isCaptured")
	purego.RegisterLibFunc(&rgfwWindowSwapBuffersOpenGL, lib, "RGFW_window_swapBuffers_OpenGL")
	purego.RegisterLibFunc(&rgfwWindowMoveMouse, lib, "RGFW_window_moveMouse")
	purego.RegisterLibFunc(&rgfwWindowCheckEvent, lib, "RGFW_window_checkEvent")
	purego.RegisterLibFunc(&rgfwWindowSetFullscreen, lib, "RGFW_window_setFullscreen")
	purego.RegisterLibFunc(&rgfwWindowIsFullscreen, lib, "RGFW_window_isFullscreen")
	purego.RegisterLibFunc(&rgfwWindowSwapInterval, lib, "RGFW_window_swapInterval_OpenGL")
	purego.RegisterLibFunc(&rgfwWindowResize, lib, "RGFW_window_resize")
	purego.RegisterLibFunc(&rgfwWindowGetSize, lib, "RGFW_window_getSize")
	purego.RegisterLibFunc(&rgfwWindowCenter, lib, "RGFW_window_center")
	purego.RegisterLibFunc(&rgfwWindowMakeCurrentContextOpenGL, lib, "RGFW_window_makeCurrentContext_OpenGL")
	purego.RegisterLibFunc(&rgfwWindowSetContextOpenGL, lib, "RGFW_window_setContext_OpenGL")

	res := rgfwInit(className, initFlags)
	if res != 0 {
		return Error(res)
	}

	return nil
}

func Deinit() {
	rgfwDeinit()
}

func PollEvents() {
	rgfwPollEvents()
}

func GetMouseScroll() (x, y float32) {
	rgfwGetMouseScroll(&x, &y)
	return
}

func CreateOpenGLContext(hints GLHints) *OpenGLContext {
	return rgfwCreateContextOpenGL(&hints)
}

func CreateWindow(name string, x, y, width, height int, flags WindowFlags) *Window {
	return rgfwCreateWindow(name, int32(x), int32(y), int32(width), int32(height), flags)
}

func (window *Window) Close() {
	rgfwCloseWindow(window)
}

func (window *Window) Show() {
	rgfwWindowShow(window)
}

func (window *Window) ShouldClose() bool {
	return rgfwWindowShouldClose(window)
}

func (window *Window) SetShouldClose(shouldClose bool) {
	rgfwWindowSetShouldClose(window, shouldClose)
}

func (window *Window) IsKeyDown(key Key) bool {
	return rgfwWindowIsKeyDown(window, key)
}

func (window *Window) GetMousePosition() (x, y int) {
	var x32, y32 int32
	_ = rgfwWindowGetMouse(window, &x32, &y32)
	return int(x32), int(y32)
}

func (window *Window) IsMouseDown(button MouseButton) bool {
	return rgfwWindowIsMouseDown(window, button)
}

func (window *Window) IsInFocus() bool {
	return rgfwWindowIsInFocus(window)
}

func (window *Window) SetCaptureMouse(capture bool) {
	rgfwWindowCaptureMouse(window, capture)
}

func (window *Window) IsMouseCaptured() bool {
	return rgfwWindowIsCaptured(window)
}

func (window *Window) SwapBuffers() {
	rgfwWindowSwapBuffersOpenGL(window)
}

func (window *Window) MoveMouse(x, y int) {
	rgfwWindowMoveMouse(window, int32(x), int32(y))
}

func (window *Window) CheckEvent() (event Event, ok bool) {
	ok = rgfwWindowCheckEvent(window, &event)
	return
}

func (window *Window) SetFullscreen(mode FullscreenMode) {
	rgfwWindowSetFullscreen(window, mode)
}

func (window *Window) IsFullscreen() bool {
	return rgfwWindowIsFullscreen(window)
}

func (window *Window) SetSwapInterval(interval int) {
	rgfwWindowSwapInterval(window, int32(interval))
}

func (window *Window) SetSize(width, height int) {
	rgfwWindowResize(window, int32(width), int32(height))
}

func (window *Window) GetSize() (width, height int) {
	var w32, h32 int32
	rgfwWindowGetSize(window, &w32, &h32)
	return int(w32), int(h32)
}

func (window *Window) Center() {
	rgfwWindowCenter(window)
}

func (window *Window) MakeContextCurrent() {
	rgfwWindowMakeCurrentContextOpenGL(window)
}

func (window *Window) SetContext(context *OpenGLContext) bool {
	return rgfwWindowSetContextOpenGL(window, context)
}
