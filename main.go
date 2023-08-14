package main

import (
	"image/color"
	"log"
	"math/rand"
	"runtime"

	"github.com/go-gl/glfw/v3.3/glfw"
	"github.com/go-gl/mathgl/mgl32"

	"tophatdemon.com/total-invasion-ii/engine"
	"tophatdemon.com/total-invasion-ii/engine/assets"
	"tophatdemon.com/total-invasion-ii/engine/ecomps"
	"tophatdemon.com/total-invasion-ii/engine/ecomps/ui"
	"tophatdemon.com/total-invasion-ii/engine/ecs"
	"tophatdemon.com/total-invasion-ii/engine/input"
	"tophatdemon.com/total-invasion-ii/engine/math2"
	"tophatdemon.com/total-invasion-ii/engine/render"
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

type GameScene struct {
	ecomps.GameScene
	// Add game specific components here
}

type Game struct {
	gameScene *GameScene
	uiScene   *ui.Scene
	camEnt    ecs.Entity
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
	for iter := game.gameScene.EntsIter(); iter.Valid(); iter = iter.Next() {
		ent := iter.Entity()
		game.gameScene.Update(ent, deltaTime)
	}
	game.uiScene.UpdateAll(deltaTime)
}

func (game *Game) Render() {
	// Render setup
	cameraTransform, _ := game.gameScene.Transforms.Get(game.camEnt)
	camera, _ := game.gameScene.Cameras.Get(game.camEnt)
	viewMat := cameraTransform.GetMatrix().Inv()
	projMat := camera.GetProjectionMatrix()
	renderContext := render.Context{
		View:           viewMat,
		Projection:     projMat,
		FogStart:       1.0,
		FogLength:      50.0,
		LightDirection: mgl32.Vec3{1.0, 0.0, 1.0}.Normalize(),
		AmbientColor:   mgl32.Vec3{0.5, 0.5, 0.5},
	}

	// Draw the scene
	for iter := game.gameScene.EntsIter(); iter.Valid(); iter = iter.Next() {
		ent := iter.Entity()
		game.gameScene.Render(ent, &renderContext)
	}

	// UI render setup
	renderContext = render.Context{
		View:       mgl32.Ident4(),
		Projection: mgl32.Ortho(0.0, WINDOW_WIDTH, WINDOW_HEIGHT, 0.0, -1.0, 10.0),
	}

	game.uiScene.RenderAll(&renderContext)
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
	scene := GameScene{
		ecomps.NewGameScene(2048),
	}

	//Load map
	var gameMap *assets.TE3File
	if gameMap, err = assets.LoadTE3File("assets/maps/ti2-malicious-intents.te3"); err != nil {
		panic(err)
	}

	if _, err := engine.SpawnGameMap(&scene.GameScene, gameMap); err != nil {
		panic(err)
	}

	// Find player spawn
	playerSpawn, _ := gameMap.FindEntWithProperty("type", "player")

	// Spawn sprites
	for _, mapEnt := range gameMap.FindEntsWithProperty("type", "enemy") {
		enemyEnt, _ := scene.AddEntity()

		scene.Transforms.Assign(enemyEnt,
			ecomps.TransformFromTranslationAngles(
				mapEnt.Position, mapEnt.Angles))

		tex := assets.GetTexture("assets/textures/sprites/wraith.png")

		scene.MeshRenders.Assign(
			enemyEnt,
			ecomps.NewMeshRender(
				assets.QuadMesh,
				assets.SpriteShader,
				tex,
			),
		)

		scene.AnimationPlayers.Assign(enemyEnt, ecomps.NewAnimationPlayer(tex.GetAnimation(0), true))
	}

	camEnt, _ := scene.AddEntity()
	scene.Cameras.Assign(camEnt, ecomps.NewCamera(70.0, WINDOW_ASPECT_RATIO, 0.1, 1000.0))
	scene.Movements.Assign(camEnt, ecomps.Movement{
		MaxSpeed:   12.0,
		YawAngle:   mgl32.DegToRad(playerSpawn.Angles[1]),
		PitchAngle: 0.0,
	})
	scene.FirstPersonControllers.Assign(camEnt,
		ecomps.NewFirstPersonController(
			ACTION_FORWARD, ACTION_BACK,
			ACTION_LEFT, ACTION_RIGHT,
			ACTION_LOOK_HORZ, ACTION_LOOK_VERT,
		),
	)
	scene.Transforms.Assign(camEnt,
		ecomps.TransformFromTranslationAngles(
			playerSpawn.Position, playerSpawn.Angles,
		),
	)

	uiScene := ui.NewUIScene(1024)
	fontTex := assets.GetTexture("assets/textures/atlases/font.png")
	for i := 1; i < 16; i += 1 {
		letterEnt, _ := uiScene.AddEntity()
		src := math2.Rect{
			X:      float32(i * 16),
			Y:      0.0,
			Width:  16.0,
			Height: 16.0,
		}
		dest := math2.Rect{
			X:      4.0 + float32(i*20),
			Y:      8.0 + rand.Float32()*32.0,
			Width:  8.0 + rand.Float32()*16.0,
			Height: 8.0 + rand.Float32()*16.0,
		}
		col := color.RGBA{
			R: uint8(rand.Intn(256)),
			G: uint8(rand.Intn(256)),
			B: uint8(rand.Intn(256)),
			A: 255,
		}
		uiScene.Boxes.Assign(letterEnt, ui.NewBox(src, dest, fontTex, col))
	}

	{
		tex := assets.GetTexture("assets/textures/sprites/wraith.png")
		wraithEnt, _ := uiScene.AddEntity()
		uiScene.Boxes.Assign(wraithEnt, ui.NewBoxFull(math2.Rect{
			X: 512.0, Y: 32.0, Width: 64.0, Height: 64.0,
		}, tex, color.RGBA{R: 255, G: 255, B: 255, A: 128}))
		a, hasAnim := tex.GetAnimationByName("die")
		if !hasAnim {
			log.Fatalln("Fuck!!!")
		}
		uiScene.AnimationPlayers.Assign(wraithEnt, ecomps.NewAnimationPlayer(a, true))
	}

	// Test text
	{
		textEnt1, _ := uiScene.AddEntity()
		text1, err := ui.NewText("assets/textures/atlases/font.fnt", "\"The Quick Brown Fox \nJumped Over The\n\n Lazy Dog.\"")
		text1.SetScale(2.0)
		if err != nil {
			panic(err)
		}
		text1.SetDest(math2.Rect{X: 400.0, Y: 100.0, Width: 500.0, Height: 400.0})
		uiScene.Texts.Assign(textEnt1, *text1)
	}
	{
		textEnt1, _ := uiScene.AddEntity()
		text1, err := ui.NewText("assets/textures/atlases/font.fnt", "* Съешь [же] ещё этих мягких\nфранцузских булок ДА ВЫПЕЙ ЧАЮ.")
		text1.SetScale(0.75).SetColor(color.RGBA{255, 0, 0, 255})
		if err != nil {
			panic(err)
		}
		text1.SetDest(math2.Rect{X: 400.0, Y: 300.0, Width: 500.0, Height: 400.0})
		uiScene.Texts.Assign(textEnt1, *text1)
	}

	input.TrapMouse()

	engine.Run(&Game{
		&scene, &uiScene, camEnt,
	})
}
