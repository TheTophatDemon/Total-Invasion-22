package engine

import (
	"runtime"
	"time"

	"github.com/go-gl/gl/v3.3-core/gl"
	"github.com/go-gl/glfw/v3.3/glfw"

	"tophatdemon.com/total-invasion-ii/engine/assets/audio"
	"tophatdemon.com/total-invasion-ii/engine/assets/cache"
	"tophatdemon.com/total-invasion-ii/engine/input"
)

func init() {
	runtime.LockOSThread()
}

type App interface {
	Update(deltaTime float32)
	Render()
}

var fps int
var updateRate float32 = 1.0 / 60.0
var window *glfw.Window

func FPS() int {
	return fps
}

func SetUpdateRate(fps int) {
	updateRate = 1.0 / float32(fps)
}

func Init(screenWidth, screenHeight int, windowTitle string) error {
	err := glfw.Init()
	if err != nil {
		return err
	}

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

	if err := audio.Init(); err != nil {
		return err
	}

	if err := gl.Init(); err != nil {
		return err
	}

	cache.InitBuiltInAssets()

	return nil
}

func Run(app App) {
	previousTime := glfw.GetTime()

	// FPS counters
	var fpsTimer float32
	var fpsTicks int
	for !window.ShouldClose() {
		// Update
		now := glfw.GetTime()
		deltaTime := float32(now - previousTime)
		previousTime = now

		//Calc FPS
		fpsTimer += deltaTime
		if fpsTimer > 1.0 {
			fpsTimer = 0.0
			fps = fpsTicks
			fpsTicks = 0
		} else {
			fpsTicks += 1
		}

		app.Update(deltaTime)

		// OpenGL settings
		gl.Enable(gl.DEPTH_TEST)
		gl.Enable(gl.CULL_FACE)
		gl.Enable(gl.BLEND)
		gl.DepthFunc(gl.LESS)
		gl.BlendFuncSeparate(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA, gl.ONE_MINUS_DST_ALPHA, gl.ONE)
		gl.CullFace(gl.BACK)
		gl.ClearColor(0.0, 0.0, 0.2, 1.0)
		gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)

		app.Render()

		CheckOpenGLError()

		input.Update()
		window.SwapBuffers()
		glfw.PollEvents()

		// Throttle the update rate if the game is running faster than max FPS
		now = glfw.GetTime()
		if frameTime := float32(now - previousTime); frameTime < updateRate {
			waitTime := time.Duration((updateRate - frameTime) * 1000.0)
			time.Sleep(waitTime * time.Millisecond)
		}
	}
}

func DeInit() {
	cache.FreeAll()
	glfw.Terminate()
}
