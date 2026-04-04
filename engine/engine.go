package engine

import (
	"runtime"

	"github.com/go-gl/gl/v3.3-core/gl"
	"github.com/go-gl/glfw/v3.3/glfw"

	"tophatdemon.com/total-invasion-ii/engine/assets/cache"
	"tophatdemon.com/total-invasion-ii/engine/clock"
	"tophatdemon.com/total-invasion-ii/engine/failure"
	"tophatdemon.com/total-invasion-ii/engine/input"
	"tophatdemon.com/total-invasion-ii/engine/tdaudio"
)

func init() {
	runtime.LockOSThread()
}

type App interface {
	Update(deltaTime float32)
	Render()
}

var debugMode bool
var window *glfw.Window

// We're tracking this ourselves so that switching to fullscreen doesn't mess up the UI
var gScreenWidth, gScreenHeight int
var gVsync bool

func FPS() int {
	return int(clock.ActualTPS())
}

func SetUpdateRate(fps int) {
	clock.SetTPS(fps)
}

func InDebugMode() bool {
	return debugMode
}

func ScreenSize() (width, height int) {
	return gScreenWidth, gScreenHeight
}

func IsFullscreen() bool {
	return window.GetMonitor() != nil
}

func IsVsync() bool {
	return gVsync
}

func SetVsync(isVsync bool) {
	gVsync = isVsync
	if isVsync {
		glfw.SwapInterval(1)
	} else {
		glfw.SwapInterval(0)
	}
}

func SetVideoMode(isFullscreen bool, width, height int) {
	if gScreenWidth != width || gScreenHeight != height {
		gScreenWidth = width
		gScreenHeight = height
		window.SetSize(width, height)
		gl.Viewport(0, 0, int32(width), int32(height))
	}
	if isFullscreen && !IsFullscreen() {
		window.SetMonitor(glfw.GetPrimaryMonitor(), 0, 0, width, height, 60)
	} else if !isFullscreen && IsFullscreen() {
		x, y := centerWindowPosition()
		window.SetMonitor(nil, x, y, width, height, 60)
	}
}

func Shutdown() {
	window.SetShouldClose(true)
}

func centerWindowPosition() (x, y int) {
	_, _, monitorW, monitorH := glfw.GetPrimaryMonitor().GetWorkarea()
	winWidth, winHeight := window.GetSize()
	// This isn't _exactly_ the center, but who cares?
	x = (monitorW / 2) - (winWidth / 2)
	y = (monitorH / 2) - (winHeight / 2)
	return
}

func CenterWindow() {
	window.SetPos(centerWindowPosition())
}

func Init(screenWidth, screenHeight int, windowTitle string, enableDebug, isFullscreen bool) {
	err := glfw.Init()
	if err != nil {
		panic(err)
	}

	debugMode = enableDebug

	glfw.WindowHint(glfw.Resizable, glfw.False)
	glfw.WindowHint(glfw.ContextVersionMajor, 3)
	glfw.WindowHint(glfw.ContextVersionMinor, 3)
	glfw.WindowHint(glfw.OpenGLProfile, glfw.OpenGLCoreProfile)
	glfw.WindowHint(glfw.OpenGLForwardCompatible, glfw.True)
	gScreenWidth = screenWidth
	gScreenHeight = screenHeight
	if debugMode {
		glfw.WindowHint(glfw.OpenGLDebugContext, glfw.True)
	} else {
		glfw.WindowHint(glfw.OpenGLDebugContext, glfw.False)
	}
	if isFullscreen {
		window, err = glfw.CreateWindow(screenWidth, screenHeight, windowTitle, glfw.GetPrimaryMonitor(), nil)
	} else {
		window, err = glfw.CreateWindow(screenWidth, screenHeight, windowTitle, nil, nil)
	}
	if err != nil {
		panic(err)
	}
	CenterWindow()
	window.MakeContextCurrent()
	glfw.SwapInterval(1)
	if err := gl.Init(); err != nil {
		panic(err)
	}

	input.Init()

	tdaudio.Init()

	cache.InitBuiltInAssets()
}

func Run(app App) {

	for !window.ShouldClose() {
		numUpdates := clock.UpdateFrame()
		for range numUpdates {
			app.Update(1.0 / float32(clock.TPS()))
			input.PostUpdate()
			tdaudio.Update()
		}

		if numUpdates > 0 {
			// OpenGL settings
			gl.Enable(gl.DEPTH_TEST)
			gl.Enable(gl.CULL_FACE)
			gl.Enable(gl.BLEND)
			gl.DepthFunc(gl.LESS)
			gl.BlendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA)

			gl.CullFace(gl.BACK)
			gl.ClearColor(0.0, 0.0, 0.2, 1.0)
			gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)

			app.Render()

			failure.CheckOpenGLError()

			window.SwapBuffers()
		}
		glfw.PollEvents()
	}

	// Reset video mode before leaving
	winX, winY := window.GetPos()
	winW, winH := window.GetSize()
	window.SetMonitor(nil, winX, winY, winW, winH, 60)
}

func DeInit() {
	cache.FreeAll()
	glfw.Terminate()
	tdaudio.Teardown()
}
