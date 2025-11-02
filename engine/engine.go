package engine

import (
	"runtime"

	"github.com/go-gl/gl/v3.3-core/gl"
	"github.com/go-gl/glfw/v3.3/glfw"

	"tophatdemon.com/total-invasion-ii/engine/assets/cache"
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

var measuredFps int
var updateRate float64 = 1.0 / 60.0
var debugMode bool
var window *glfw.Window

func FPS() int {
	return measuredFps
}

func SetUpdateRate(fps int) {
	updateRate = 1.0 / float64(fps)
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
	previousTime := glfw.GetTime()
	var accumulator float64

	// FPS counters
	var fpsTimer float64
	var fpsTicks int
	for !window.ShouldClose() {
		// Update
		now := glfw.GetTime()
		deltaTime := now - previousTime
		previousTime = now

		// Calc FPS
		fpsTimer += deltaTime
		fpsTicks++
		if fpsTimer > 1.0 {
			fpsTimer = 0.0
			measuredFps = fpsTicks
			fpsTicks = 0
		}

		// Prevents window moving events from causing the game to skip large portions of time.
		if deltaTime > updateRate {
			deltaTime = updateRate
		}

		// Run updates by splitting the time since the last frame into fixed time steps.
		// There is an upper limit to the number of timesteps ran per frame to prevent lag from spiralling until the game stops completely.
		updateCount := 0
		for accumulator += deltaTime; accumulator >= updateRate && updateCount < 5; accumulator -= updateRate {
			app.Update(float32(updateRate))
			input.Update()
			tdaudio.Update()
			updateCount++
		}

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

		for glfw.GetTime()-previousTime < updateRate {
			// Throttle the update rate if the game is running faster than max FPS
			// Only necessary on Windows for some reason.
		}
	}
}

func DeInit() {
	cache.FreeAll()
	glfw.Terminate()
	tdaudio.Teardown()
}
