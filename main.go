package main

import (
	"fmt"
	"log"
	"math"
	"runtime"

	"github.com/go-gl/gl/v3.3-core/gl"
	"github.com/go-gl/glfw/v3.3/glfw"
	"github.com/go-gl/mathgl/mgl32"

	"tophatdemon.com/total-invasion-ii/engine"
	"tophatdemon.com/total-invasion-ii/engine/assets"
	"tophatdemon.com/total-invasion-ii/engine/comps"
	"tophatdemon.com/total-invasion-ii/engine/ecs"
	"tophatdemon.com/total-invasion-ii/engine/input"
	"tophatdemon.com/total-invasion-ii/engine/systems"
	"tophatdemon.com/total-invasion-ii/engine/math2"
)

const (
	WINDOW_WIDTH        = 1280
	WINDOW_HEIGHT       = 720
	WINDOW_ASPECT_RATIO = float32(WINDOW_WIDTH) / WINDOW_HEIGHT

	ECS_RESERVE_COUNT = 1024

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

	assets.InitShaders()

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

	// Load the texture
	texture := assets.GetTexture("assets/textures/tiles/psa2.png")
	texture2 := assets.GetTexture("assets/textures/tiles/pillar_head.png")

	//Generate cube
	cube := assets.CreateMesh(cubeVertices, cubeIndices)
	cube.SetGroup("front", assets.Group{ Offset: 12, Length: 6 })

	cylinder, err := assets.GetMesh("assets/models/shapes/wedge.obj")
	if err != nil {
		log.Fatalln(err)
	}

	//Create ECS world
	world := ecs.CreateWorld(ECS_RESERVE_COUNT)
	cameras       := ecs.CreateStorage[comps.Camera]               (ECS_RESERVE_COUNT)
	transforms    := ecs.CreateStorage[comps.Transform]            (ECS_RESERVE_COUNT)
	movers        := ecs.CreateStorage[comps.Movement]             (ECS_RESERVE_COUNT)
	fpControllers := ecs.CreateStorage[comps.FirstPersonController](ECS_RESERVE_COUNT)
	
	//Load map
	te3, err := assets.LoadTE3File("assets/maps/E3M1.te3")
	if err != nil {
		log.Println("Map loading error: ", err)
	}
	te3Mesh, err := te3.BuildMesh()
	if err != nil {
		log.Println("Map generating error: ", err)
	}

	//Camera controller
	cameraEnt := world.NewEnt()
	cameras.Assign(cameraEnt, comps.NewCamera(45.0, WINDOW_ASPECT_RATIO, 0.1, 1000.0))

	//Place camera at player's position
	camTrans := comps.Transform{}
	playerSpawn, _ := te3.FindEntWithProperty("type", "player spawn")
	camTrans.SetPosition(playerSpawn.Position)
	transforms.Assign(cameraEnt, camTrans)
	
	movers.Assign(cameraEnt, comps.Movement{MaxSpeed: 12.0, YawAngle: mgl32.DegToRad(playerSpawn.Angles[1]), PitchAngle: 0.0})
	fpControllers.Assign(cameraEnt, comps.FirstPersonController{ 
		ForwardAction: ACTION_FORWARD, 
		BackAction: ACTION_BACK, 
		StrafeLeftAction: ACTION_LEFT, 
		StrafeRightAction: ACTION_RIGHT,
		LookHorzAction: ACTION_LOOK_HORZ, 
		LookVertAction: ACTION_LOOK_VERT,
	})

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

	var angle float64

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

		systems.UpdateFirstPersonControllers(elapsed, world, movers, fpControllers)
		systems.UpdateMovement(elapsed, world, movers, transforms)

		tf, _ := transforms.Get(cameraEnt)
		viewMat := tf.GetMatrix().Inv()
		mvp := projMat.Mul4(viewMat)
		// tf, _ := transforms.Get(cameraEnt)
		// mvp = mvp.Mul4(tf.GetMatrix())

		// Render
		assets.MapShader.Use()
		cube.Bind()
		
		gl.UniformMatrix4fv(assets.MapShader.GetUniformLoc("uMVP"), 1, false, &mvp[0])
		gl.Uniform1i(assets.MapShader.GetUniformLoc("uTex"), 0)

		gl.Uniform1f(assets.MapShader.GetUniformLoc("uFogStart"), 1.0)
		gl.Uniform1f(assets.MapShader.GetUniformLoc("uFogLength"), 50.0)
		
		gl.ActiveTexture(gl.TEXTURE0)
		gl.BindTexture(gl.TEXTURE_2D, texture.GetID())
		
		// cube.DrawAll()
		
		//cube.DrawGroup("front")
		cylinder.Bind()
		cylinder.DrawAll()
		// cylinder.DrawGroup("culld")

		//Draw the map
		te3Mesh.Bind()
		model := mgl32.Ident4()
		gl.UniformMatrix4fv(assets.MapShader.GetUniformLoc("uModelTransform"), 1, false, &model[0])
		for _, group := range te3Mesh.GetGroupNames() {
			gl.BindTexture(gl.TEXTURE_2D, assets.GetTexture(group).GetID())
			te3Mesh.DrawGroup(group)
		}
		// te3Mesh.DrawAll()

		//Draw second one pointing at the thing
		cylinder.Bind()
		angle += float64(elapsed)
		eye := mgl32.Vec3{3 * float32(math.Cos(angle)), 3, 3 * float32(math.Sin(angle))}
		lookMtx := math2.LookAtV(eye, mgl32.Vec3{0, 0, 0}, mgl32.Vec3{0, 1, 0})
		look := comps.TransformFromMatrix(lookMtx)
		// look.SetPosition(eye)
		mvp = mvp.Mul4(look.GetMatrix())
		// mvp = mvp.Mul4(lookMtx)
		gl.UniformMatrix4fv(assets.MapShader.GetUniformLoc("uMVP"), 1, false, &mvp[0])
		gl.BindTexture(gl.TEXTURE_2D, texture2.GetID())
		cylinder.DrawAll()

		engine.CheckOpenGLError()

		input.Update()
		window.SwapBuffers()
		glfw.PollEvents()
	}

	cylinder.Free()
	cube.Free()
	assets.FreeShaders()
}

var cubeIndices = []uint32{
	0, 1, 2, 3, 4, 5,
	6, 7, 8, 9, 10, 11,
	12, 13, 14, 15, 16, 17,
	18, 19, 20, 21, 22, 23,
	24, 25, 26, 27, 28, 29,
	30, 31, 32, 33, 34, 35,
}

var cubeVertices = assets.Vertices{
	Pos: []mgl32.Vec3{
		//Bottom
		{-1.0, -1.0, -1.0},
		{ 1.0, -1.0, -1.0},
		{-1.0, -1.0,  1.0},
		{ 1.0, -1.0, -1.0},
		{ 1.0, -1.0,  1.0},
		{-1.0, -1.0,  1.0},
		//Top
		{-1.0,  1.0, -1.0},
		{-1.0,  1.0,  1.0},
		{ 1.0,  1.0, -1.0},
		{ 1.0,  1.0, -1.0},
		{-1.0,  1.0,  1.0},
		{ 1.0,  1.0,  1.0},
		//Front
		{-1.0, -1.0,  1.0},
		{ 1.0, -1.0,  1.0},
		{-1.0,  1.0,  1.0},
		{ 1.0, -1.0,  1.0},
		{ 1.0,  1.0,  1.0},
		{-1.0,  1.0,  1.0},
		//Back
		{-1.0, -1.0, -1.0},
		{-1.0,  1.0, -1.0},
		{ 1.0, -1.0, -1.0},
		{ 1.0, -1.0, -1.0},
		{-1.0,  1.0, -1.0},
		{ 1.0,  1.0, -1.0},
		//Left
		{-1.0, -1.0,  1.0},
		{-1.0,  1.0, -1.0},
		{-1.0, -1.0, -1.0},
		{-1.0, -1.0,  1.0},
		{-1.0,  1.0,  1.0}, 
		{-1.0,  1.0, -1.0},
		//Right
		{ 1.0, -1.0,  1.0}, 
		{ 1.0, -1.0, -1.0},
		{ 1.0,  1.0, -1.0}, 
		{ 1.0, -1.0,  1.0}, 
		{ 1.0,  1.0, -1.0}, 
		{ 1.0,  1.0,  1.0},  
	},
	TexCoord: []mgl32.Vec2{
		// Bottom
		{0.0, 0.0},
		{1.0, 0.0},
		{0.0, 1.0},
		{1.0, 0.0},
		{1.0, 1.0},
		{0.0, 1.0},
		//Top
		{0.0, 0.0},
		{0.0, 1.0},
		{1.0, 0.0},
		{1.0, 0.0},
		{0.0, 1.0},
		{1.0, 1.0},
		//Front
		{1.0, 0.0},
		{0.0, 0.0},
		{1.0, 1.0},
		{0.0, 0.0},
		{0.0, 1.0},
		{1.0, 1.0},
		//Back
		{0.0, 0.0},
		{0.0, 1.0},
		{1.0, 0.0},
		{1.0, 0.0},
		{0.0, 1.0},
		{1.0, 1.0},
		//Left
		{0.0, 1.0},
		{1.0, 0.0},
		{0.0, 0.0},
		{0.0, 1.0},
		{1.0, 1.0},
		{1.0, 0.0},
		//Right
		{1.0, 1.0},
		{1.0, 0.0},
		{0.0, 0.0},
		{1.0, 1.0},
		{0.0, 0.0},
		{0.0, 1.0},
	},
	Normal:   nil,
	Color:    nil,
}