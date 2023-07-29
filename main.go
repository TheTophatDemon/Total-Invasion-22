package main

import (
	"runtime"

	"github.com/go-gl/glfw/v3.3/glfw"
	"github.com/go-gl/mathgl/mgl32"

	"tophatdemon.com/total-invasion-ii/engine"
	"tophatdemon.com/total-invasion-ii/engine/assets"
	"tophatdemon.com/total-invasion-ii/engine/ecomps"
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

func init() {
	runtime.LockOSThread()
}

type Game struct {
	sc     *scene.Scene
	camEnt scene.Entity
}

func (game *Game) Update(deltaTime float32) {
	//Free mouse
	if input.IsActionJustPressed(ACTION_TRAP_MOUSE) {
		if input.IsMouseTrapped() {
			input.UntrapMouse()
		} else {
			input.TrapMouse()
		}
	}

	//Update scene
	for iter := game.sc.EntsIter(); iter.Valid(); iter = iter.Next() {
		ent := iter.Entity()
		ecomps.UpdateDefaultComps(game.sc, ent, deltaTime)
	}
}

func (game *Game) Render() {
	//Render setup
	cameraTransform, _ := ecomps.TransformComps.Get(game.camEnt)
	camera, _ := ecomps.CameraComps.Get(game.camEnt)
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
	for iter := game.sc.EntsIter(); iter.Valid(); iter = iter.Next() {
		ent := iter.Entity()
		ecomps.RenderDefaultComps(game.sc, ent, &renderContext)
	}
}

func main() {
	err := engine.Init(WINDOW_WIDTH, WINDOW_HEIGHT, "Total Invasion II")
	defer engine.DeInit()
	if err != nil {
		panic(err)
	}

	input.BindActionKey(ACTION_FORWARD, glfw.KeyW)
	input.BindActionKey(ACTION_BACK, glfw.KeyS)
	input.BindActionKey(ACTION_LEFT, glfw.KeyA)
	input.BindActionKey(ACTION_RIGHT, glfw.KeyD)
	input.BindActionKey(ACTION_TRAP_MOUSE, glfw.KeyEscape)
	input.BindActionMouseMove(ACTION_LOOK_HORZ, input.MOUSE_AXIS_X, MOUSE_SENSITIVITY)
	input.BindActionMouseMove(ACTION_LOOK_VERT, input.MOUSE_AXIS_Y, MOUSE_SENSITIVITY)

	//Create scene
	sc := scene.NewScene(2048)
	ecomps.RegisterAll(sc)

	//Load map
	var gameMap *assets.TE3File
	if gameMap, err = assets.LoadTE3File("assets/maps/ti2-malicious-intents.te3"); err != nil {
		panic(err)
	}

	if _, err := engine.SpawnGameMap(sc, gameMap); err != nil {
		panic(err)
	}

	// Find player spawn
	playerSpawn, _ := gameMap.FindEntWithProperty("type", "player")

	// Spawn sprites
	for _, mapEnt := range gameMap.FindEntsWithProperty("type", "enemy") {
		enemyEnt, _ := sc.AddEntity()

		ecomps.AddTransform(enemyEnt,
			ecomps.TransformFromTranslationAngles(
				mapEnt.Position, mapEnt.Angles))

		tex := assets.GetTexture("assets/textures/sprites/wraith.png")

		ecomps.AddMeshRender(
			enemyEnt,
			assets.SpriteMesh,
			assets.SpriteShader,
			tex)

		ecomps.AddAnimationPlayer(enemyEnt, tex.GetAnimation(0), true)
	}

	camEnt, _ := sc.AddEntity()
	ecomps.AddCamera(camEnt, 70.0, WINDOW_ASPECT_RATIO, 0.1, 1000.0)
	ecomps.AddMovement(camEnt, ecomps.Movement{
		MaxSpeed:   12.0,
		YawAngle:   mgl32.DegToRad(playerSpawn.Angles[1]),
		PitchAngle: 0.0,
	})
	ecomps.AddFirstPersonController(camEnt,
		ACTION_FORWARD, ACTION_BACK,
		ACTION_LEFT, ACTION_RIGHT,
		ACTION_LOOK_HORZ, ACTION_LOOK_VERT)
	ecomps.AddTransform(camEnt,
		ecomps.TransformFromTranslationAngles(
			playerSpawn.Position, playerSpawn.Angles))

	input.TrapMouse()

	engine.Run(&Game{
		sc, camEnt,
	})
}
