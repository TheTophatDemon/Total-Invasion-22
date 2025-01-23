package engine

import (
	"runtime"
	"time"

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

var fps int
var updateRate float64 = 1.0 / 60.0
var window *glfw.Window

func FPS() int {
	return fps
}

func SetUpdateRate(fps int) {
	updateRate = 1.0 / float64(fps)
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

	tdaudio.Init()

	if err := gl.Init(); err != nil {
		return err
	}

	cache.InitBuiltInAssets()

	return nil
}

func Run(app App) {
	previousTime := time.Now()
	var accumulator float64

	// FPS counters
	var fpsTimer float64
	var fpsTicks int
	for !window.ShouldClose() {
		// Update
		now := time.Now()
		deltaTime := float64(now.Sub(previousTime).Seconds())

		//Calc FPS
		fpsTimer += deltaTime
		fpsTicks++
		if fpsTimer > 1.0 {
			fpsTimer = 0.0
			fps = fpsTicks
			fpsTicks = 0
			//fmt.Println("FPS:", fps)
		}

		if deltaTime > updateRate {
			deltaTime = updateRate
		}
		previousTime = now

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
		gl.BlendFuncSeparate(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA, gl.ONE_MINUS_DST_ALPHA, gl.ONE)
		gl.CullFace(gl.BACK)
		gl.ClearColor(0.0, 0.0, 0.2, 1.0)
		gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)

		app.Render()

		failure.CheckOpenGLError()

		window.SwapBuffers()

		glfw.PollEvents()

		// Throttle the update rate if the game is running faster than max FPS
		for {
			now = time.Now()
			if frameTime := now.Sub(previousTime).Seconds(); frameTime >= updateRate {
				break
			}
		}
	}
}

func DeInit() {
	cache.FreeAll()
	glfw.Terminate()
	tdaudio.Teardown()
}
