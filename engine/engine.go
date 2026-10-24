package engine

import (
	"runtime"

	"github.com/go-gl/gl/v3.3-core/gl"

	"tophatdemon.com/total-invasion-ii/engine/assets/cache"
	"tophatdemon.com/total-invasion-ii/engine/clock"
	"tophatdemon.com/total-invasion-ii/engine/failure"
	"tophatdemon.com/total-invasion-ii/engine/input"
	"tophatdemon.com/total-invasion-ii/engine/rgfw"
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
var window *rgfw.Window
var glContext *rgfw.OpenGLContext

// We're tracking this ourselves so that switching to fullscreen doesn't mess up the UI
var gScreenWidth, gScreenHeight int = 1280, 720
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
	return window.IsFullscreen()
}

func IsVsync() bool {
	return gVsync
}

func SetVsync(isVsync bool) {
	gVsync = isVsync
	if isVsync {
		window.SetSwapInterval(1)
	} else {
		window.SetSwapInterval(0)
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
		window.SetFullscreen(rgfw.FULLSCREEN_BORDERLESS)
	} else if !isFullscreen && IsFullscreen() {
		window.SetFullscreen(rgfw.FULLSCREEN_NONE)
		window.Center()
	}
}

func Shutdown() {
	window.SetShouldClose(true)
}

func CenterWindow() {
	window.Center()
}

func Init(screenWidth, screenHeight int, windowTitle string, enableDebug, isFullscreen, isVsync bool) {
	err := rgfw.Init("Total Invasion 22", rgfw.INIT_OPENGL)
	if err != nil {
		panic(err)
	}

	debugMode = enableDebug

	gScreenWidth = screenWidth
	gScreenHeight = screenHeight

	var flags rgfw.WindowFlags = rgfw.WINDOW_CENTER |
		rgfw.WINDOW_OPEN_GL |
		rgfw.WINDOW_NO_RESIZE |
		rgfw.WINDOW_FOCUS_ON_SHOW |
		rgfw.WINDOW_HIDE // Have to hide the window until the OpenGL context is set up.

	// This flag doesn't work because of a RGFW bug.
	// if isFullscreen {
	// 	flags |= rgfw.WINDOW_FULLSCREEN_BORDERLESS
	// }

	window = rgfw.CreateWindow(windowTitle, 0, 0, screenWidth, screenHeight, flags)
	glContext = rgfw.CreateOpenGLContext(rgfw.GLHints{
		MajorVersion: 3, MinorVersion: 3,
		Profile:      rgfw.GL_CORE,
		Debug:        debugMode,
		DoubleBuffer: 1,
	})
	window.SetContext(glContext)
	window.Show()
	window.MakeContextCurrent()
	SetVsync(isVsync)

	if isFullscreen {
		window.SetFullscreen(rgfw.FULLSCREEN_BORDERLESS)
	}

	if err := gl.Init(); err != nil {
		panic(err)
	}

	input.Init(window)

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
			gl.ClearColor(0.0, 0.0, 0.2, 1.0)
			gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)
			app.Render()

			failure.CheckOpenGLError()

			window.SwapBuffers()
		}
		rgfw.PollEvents()
	}
}

func DeInit() {
	cache.FreeAll()
	rgfw.Deinit()
	tdaudio.Teardown()
}
