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

func FPS() int {
	return int(clock.ActualTPS())
}

func SetUpdateRate(fps int) {
	clock.SetTPS(fps)
}

func InDebugMode() bool {
	return debugMode
}

func Init(screenWidth, screenHeight int, windowTitle string, enableDebug bool) error {
	err := glfw.Init()
	if err != nil {
		return err
	}

	debugMode = enableDebug

	glfw.WindowHint(glfw.Resizable, glfw.False)
	glfw.WindowHint(glfw.ContextVersionMajor, 3)
	glfw.WindowHint(glfw.ContextVersionMinor, 3)
	glfw.WindowHint(glfw.OpenGLProfile, glfw.OpenGLCoreProfile)
	glfw.WindowHint(glfw.OpenGLForwardCompatible, glfw.True)
	glfw.WindowHint(glfw.OpenGLDebugContext, glfw.True)
	window, err = glfw.CreateWindow(screenWidth, screenHeight, windowTitle, nil, nil)
	if err != nil {
		return err
	}

	window.MakeContextCurrent()
	input.Init()

	tdaudio.Init()

	if err := gl.Init(); err != nil {
		return err
	}

	cache.InitBuiltInAssets()

	return nil
}

func Run(app App) {

	for !window.ShouldClose() {
		numUpdates := clock.UpdateFrame()
		for range numUpdates {
			app.Update(1.0 / float32(clock.TPS()))
			input.Update()
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
			glfw.PollEvents()
		}
	}
}

func DeInit() {
	cache.FreeAll()
	glfw.Terminate()
	tdaudio.Teardown()
}
