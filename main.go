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

	ACTION_FORWARD    = "MoveForward"
	ACTION_BACK       = "MoveBack"
	ACTION_LEFT       = "StrafeLeft"
	ACTION_RIGHT      = "StrafeRight"
	ACTION_LOOK_HORZ  = "LookHorz"
	ACTION_LOOK_VERT  = "LookVert"
	ACTION_TRAP_MOUSE = "TrapMouse"

	MOUSE_SENSITIVITY = 0.005
)

var Transforms *scene.ComponentStorage[comps.Transform]
var Movements *scene.ComponentStorage[comps.Movement]
var Cameras *scene.ComponentStorage[comps.Camera]
var FirstPersonControllers *scene.ComponentStorage[comps.FirstPersonController]
var MeshRenders *scene.ComponentStorage[comps.MeshRender]
var AnimationPlayers *scene.ComponentStorage[comps.AnimationPlayer]

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
	defer assets.FreeTextures()
	defer assets.FreeMeshes()
	defer assets.FreeBuiltInAssets()

	input.BindActionKey(ACTION_FORWARD, glfw.KeyW)
	input.BindActionKey(ACTION_BACK, glfw.KeyS)
	input.BindActionKey(ACTION_LEFT, glfw.KeyA)
	input.BindActionKey(ACTION_RIGHT, glfw.KeyD)
	input.BindActionKey(ACTION_TRAP_MOUSE, glfw.KeyEscape)
	input.BindActionMouseMove(ACTION_LOOK_HORZ, input.MOUSE_AXIS_X, MOUSE_SENSITIVITY)
	input.BindActionMouseMove(ACTION_LOOK_VERT, input.MOUSE_AXIS_Y, MOUSE_SENSITIVITY)

	engine.CheckOpenGLError()

	//Create scene
	sc := scene.NewScene(1024)
	Transforms = scene.RegisterComponent[comps.Transform](sc)
	Movements = scene.RegisterComponent[comps.Movement](sc)
	Cameras = scene.RegisterComponent[comps.Camera](sc)
	FirstPersonControllers = scene.RegisterComponent[comps.FirstPersonController](sc)
	MeshRenders = scene.RegisterComponent[comps.MeshRender](sc)
	AnimationPlayers = scene.RegisterComponent[comps.AnimationPlayer](sc)

	//Load map
	gameMap, err := engine.LoadGameMap("assets/maps/ti2-malicious-intents.te3")
	if err != nil {
		log.Println("Map loading error: ", err)
	}

	// Find player spawn
	playerSpawn, _ := gameMap.FindEntWithProperty("type", "player")

	// Spawn sprites
	for _, mapEnt := range gameMap.FindEntsWithProperty("type", "enemy") {
		enemyEnt, _ := sc.AddEntity()
		Transforms.Assign(enemyEnt, comps.TransformFromTranslation(mgl32.Vec3{mapEnt.Position[0], mapEnt.Position[1], mapEnt.Position[2]}))
		MeshRenders.Assign(enemyEnt, comps.MeshRender{
			Mesh:   assets.SpriteMesh,
			Shader: assets.SpriteShader,
		})
	}

	camEnt, _ := sc.AddEntity()
	Cameras.Assign(camEnt, comps.NewCamera(70.0, WINDOW_ASPECT_RATIO, 0.1, 1000.0))
	Movements.Assign(camEnt, comps.Movement{
		MaxSpeed:   12.0,
		YawAngle:   mgl32.DegToRad(playerSpawn.Angles[1]),
		PitchAngle: 0.0,
	})
	FirstPersonControllers.Assign(camEnt, comps.FirstPersonController{
		ForwardAction:     ACTION_FORWARD,
		BackAction:        ACTION_BACK,
		StrafeLeftAction:  ACTION_LEFT,
		StrafeRightAction: ACTION_RIGHT,
		LookHorzAction:    ACTION_LOOK_HORZ,
		LookVertAction:    ACTION_LOOK_VERT,
	})
	Transforms.Assign(camEnt, comps.TransformFromTranslation(playerSpawn.Position))

	input.TrapMouse()

	// Configure global settings
	gl.Enable(gl.DEPTH_TEST)
	gl.Enable(gl.CULL_FACE)
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
		deltaTime := float32(time - previousTime)
		previousTime = time

		//Calc FPS
		fpsTimer += deltaTime
		if fpsTimer > 1.0 {
			fpsTimer = 0.0
			fps = fpsTicks
			fpsTicks = 0
			fmt.Printf("FPS: %v\n", fps)
		} else {
			fpsTicks += 1
		}

		//Free mouse
		if input.IsActionJustPressed(ACTION_TRAP_MOUSE) {
			if input.IsMouseTrapped() {
				input.UntrapMouse()
			} else {
				input.TrapMouse()
			}
		}

		//Update scene
		for iter := sc.EntsIter(); iter.Valid(); iter = iter.Next() {
			ent := iter.Entity()

			controller, ok := FirstPersonControllers.Get(ent)
			if ok {
				controller.Update(Movements, ent, deltaTime)
			}

			movement, ok := Movements.Get(ent)
			if ok {
				movement.Update(Transforms, ent, deltaTime)
			}

			animPlayer, ok := AnimationPlayers.Get(ent)
			if ok {
				animPlayer.Update(deltaTime)
			}
		}

		gameMap.Update(deltaTime)

		//Render setup
		cameraTransform, _ := Transforms.Get(camEnt)
		camera, _ := Cameras.Get(camEnt)
		viewMat := cameraTransform.GetMatrix().Inv()
		projMat := camera.GetProjectionMatrix()
		renderContext := scene.RenderContext{
			View:           viewMat,
			Projection:     projMat,
			FogStart:       1.0,
			FogLength:      50.0,
			LightDirection: mgl32.Vec3{1.0, 0.0, 1.0}.Normalize(),
			AmbientColor:   mgl32.Vec3{0.4, 0.4, 0.4},
		}

		//Draw the scene
		for iter := sc.EntsIter(); iter.Valid(); iter = iter.Next() {
			ent := iter.Entity()

			meshRender, ok := MeshRenders.Get(ent)
			if ok {
				meshRender.Render(Transforms, ent, &renderContext)
			}
		}

		//Draw the map
		gameMap.Render(&renderContext)

		engine.CheckOpenGLError()

		input.Update()
		window.SwapBuffers()
		glfw.PollEvents()
	}
}
