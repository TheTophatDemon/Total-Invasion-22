package main

import (
	"fmt"
	"log"
	"runtime"

	"github.com/go-gl/gl/v3.3-core/gl"
	"github.com/go-gl/glfw/v3.3/glfw"
	"github.com/go-gl/mathgl/mgl32"

	"tophatdemon.com/total-invasion-ii/engine"
	"tophatdemon.com/total-invasion-ii/engine/assets"
	"tophatdemon.com/total-invasion-ii/engine/comps"
	"tophatdemon.com/total-invasion-ii/engine/input"
	"tophatdemon.com/total-invasion-ii/engine/scene"
)

const (
	WINDOW_WIDTH        = 1280
	WINDOW_HEIGHT       = 720
	WINDOW_ASPECT_RATIO = float32(WINDOW_WIDTH) / WINDOW_HEIGHT

	ACTION_FORWARD   = "MoveForward"
	ACTION_BACK      = "MoveBack"
	ACTION_LEFT      = "StrafeLeft"
	ACTION_RIGHT     = "StrafeRight"
	ACTION_LOOK_HORZ = "LookHorz"
	ACTION_LOOK_VERT = "LookVert"

	MOUSE_SENSITIVITY = 0.005
)

func init() {
	runtime.LockOSThread()
}

func main() {
	err := glfw.Init()
	if err != nil {
		panic(err)
	}
	defer glfw.Terminate()

	glfw.WindowHint(glfw.Resizable, glfw.False)
	glfw.WindowHint(glfw.ContextVersionMajor, 3)
	glfw.WindowHint(glfw.ContextVersionMinor, 3)
	glfw.WindowHint(glfw.OpenGLProfile, glfw.OpenGLCoreProfile)
	glfw.WindowHint(glfw.OpenGLForwardCompatible, glfw.True)
	glfw.WindowHint(glfw.OpenGLDebugContext, glfw.True)
	window, err := glfw.CreateWindow(WINDOW_WIDTH, WINDOW_HEIGHT, "Total Invasion II", nil, nil)
	if err != nil {
		panic(err)
	}

	window.MakeContextCurrent()

	if err := gl.Init(); err != nil {
		panic(err)
	}

	assets.InitBuiltInAssets()

	input.BindActionKey(ACTION_FORWARD, glfw.KeyW)
	input.BindActionKey(ACTION_BACK, glfw.KeyS)
	input.BindActionKey(ACTION_LEFT, glfw.KeyA)
	input.BindActionKey(ACTION_RIGHT, glfw.KeyD)
	input.BindActionMouseMove(ACTION_LOOK_HORZ, input.MOUSE_AXIS_X, MOUSE_SENSITIVITY)
	input.BindActionMouseMove(ACTION_LOOK_VERT, input.MOUSE_AXIS_Y, MOUSE_SENSITIVITY)
	// input.BindActionKey(ACTION_LOOK_VERT, glfw.KeyUp)
	// input.BindActionKey(ACTION_LOOK_HORZ, glfw.KeyRight)

	engine.CheckOpenGLError()

	projMat := mgl32.Perspective(mgl32.DegToRad(45.0), WINDOW_ASPECT_RATIO, 0.1, 100.0)
	// viewMat := mgl32.LookAtV(mgl32.Vec3{3, 3, 3}, mgl32.Vec3{0, 0, 0}, mgl32.Vec3{0, 1, 0})

	//Create scene
	sc := scene.NewScene()

	//Load map
	gameMap, err := engine.LoadGameMap("assets/maps/E3M1.te3")
	if err != nil {
		log.Println("Map loading error: ", err)
	}

	// Find player spawn
	playerSpawn, _ := gameMap.FindEntWithProperty("type", "player spawn")

	// Spawn sprites
	for _, mapEnt := range gameMap.FindEntsWithProperty("type", "enemy") {
		enemyEnt := sc.AddEntity()
		transform := comps.TransformFromTranslation(mgl32.Vec3{mapEnt.Position[0], mapEnt.Position[1], mapEnt.Position[2]})
		sprite := comps.NewSpriteRender(assets.GetTexture("assets/textures/sprites/wraith.png"))
		sc.AddComponents(enemyEnt, transform, sprite)
		sprite.Anim.Play()
	}

	camEnt := sc.AddEntity()
	sc.AddComponents(camEnt,
		comps.NewCamera(70.0, WINDOW_ASPECT_RATIO, 0.1, 1000.0),
		&comps.Transform{},
		&comps.Movement{
			MaxSpeed:   12.0,
			YawAngle:   mgl32.DegToRad(playerSpawn.Angles[1]),
			PitchAngle: 0.0,
		},
		&comps.FirstPersonController{
			ForwardAction:     ACTION_FORWARD,
			BackAction:        ACTION_BACK,
			StrafeLeftAction:  ACTION_LEFT,
			StrafeRightAction: ACTION_RIGHT,
			LookHorzAction:    ACTION_LOOK_HORZ,
			LookVertAction:    ACTION_LOOK_VERT,
		},
	)

	//Place camera at player's position
	var tr *comps.Transform
	tr, err = scene.GetComponent(sc, camEnt, tr)
	if err != nil {
		log.Fatalln(err)
	}
	tr.SetPosition(playerSpawn.Position)

	//input.TrapMouse()

	// Configure global settings
	gl.Enable(gl.DEPTH_TEST)
	gl.Disable(gl.CULL_FACE)
	gl.DepthFunc(gl.LESS)
	gl.ClearColor(0.0, 0.0, 0.2, 1.0)

	previousTime := glfw.GetTime()

	//FPS counters
	var fpsTimer float32
	var fps, fpsTicks int

	for !window.ShouldClose() {
		gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)

		// Update
		time := glfw.GetTime()
		elapsed := float32(time - previousTime)
		previousTime = time

		//Calc FPS
		fpsTimer += elapsed
		if fpsTimer > 1.0 {
			fpsTimer = 0.0
			fps = fpsTicks
			fpsTicks = 0
			fmt.Printf("FPS: %v\n", fps)
		} else {
			fpsTicks += 1
		}

		sc.Update(elapsed)

		gameMap.Update(elapsed)

		// Render
		viewMat := tr.GetMatrix().Inv()
		mvp := projMat.Mul4(viewMat)

		//Draw the map
		gameMap.Render(mvp)
		//And the scene
		sc.Render(mvp, projMat, viewMat)

		engine.CheckOpenGLError()

		input.Update()
		window.SwapBuffers()
		glfw.PollEvents()
	}

	assets.FreeTextures()
	assets.FreeBuiltInAssets()
}
