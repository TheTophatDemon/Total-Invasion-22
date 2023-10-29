package game

import (
	"errors"
	"fmt"

	"github.com/go-gl/mathgl/mgl32"
	"tophatdemon.com/total-invasion-ii/engine"
	"tophatdemon.com/total-invasion-ii/engine/assets/te3"
	"tophatdemon.com/total-invasion-ii/engine/math2"
	"tophatdemon.com/total-invasion-ii/engine/math2/collision"
	"tophatdemon.com/total-invasion-ii/engine/render"
	"tophatdemon.com/total-invasion-ii/engine/world"
	"tophatdemon.com/total-invasion-ii/engine/world/comps"
	"tophatdemon.com/total-invasion-ii/engine/world/comps/ui"
	"tophatdemon.com/total-invasion-ii/game/ents"
	"tophatdemon.com/total-invasion-ii/game/settings"
)

type World struct {
	UI            *ui.Scene
	Players       *world.Storage[ents.Player]
	Enemies       *world.Storage[ents.Enemy]
	GameMap       *world.Map
	CurrentPlayer world.Id[ents.Player]
	FPSCounter    world.Id[ui.Text]
}

func NewWorld(mapPath string) (*World, error) {
	UI := ui.NewUIScene(256, 64)
	Players := world.NewStorage[ents.Player](8)
	Enemies := world.NewStorage[ents.Enemy](256)

	te3File, err := te3.LoadTE3File(mapPath)
	if err != nil {
		return nil, err
	}

	GameMap, err := world.NewMap(te3File)
	if err != nil {
		return nil, err
	}

	// Set panel collision shapes
	panelShapeX := collision.ShapeBox(math2.BoxFromExtents(1.0, 1.0, 0.5))
	panelShapeZ := collision.ShapeBox(math2.BoxFromExtents(0.5, 1.0, 1.0))
	for _, shapeName := range [...]string{
		"assets/models/shapes/bars.obj",
		"assets/models/shapes/panel.obj",
	} {
		err = errors.Join(
			GameMap.SetTileCollisionShapesForAngles(shapeName, 0, 45, 0, 360, panelShapeX),
			GameMap.SetTileCollisionShapesForAngles(shapeName, 45, 135, 0, 360, panelShapeZ),
			GameMap.SetTileCollisionShapesForAngles(shapeName, 135, 225, 0, 360, panelShapeX),
			GameMap.SetTileCollisionShapesForAngles(shapeName, 225, 315, 0, 360, panelShapeZ),
			GameMap.SetTileCollisionShapesForAngles(shapeName, 315, 360, 0, 360, panelShapeX),
		)
		if err != nil {
			return nil, err
		}
	}

	// Set cube collision shapes
	for _, shapeName := range [...]string{
		"assets/models/shapes/cube.obj",
		"assets/models/shapes/cube_2tex.obj",
		"assets/models/shapes/edge_panel.obj",
		"assets/models/shapes/cube_marker.obj",
		"assets/models/shapes/bridge.obj",
	} {
		err = GameMap.SetTileCollisionShapes(shapeName, collision.ShapeBox(math2.BoxFromRadius(1.0)))
		if err != nil {
			return nil, err
		}
	}

	// Spawn player
	playerSpawn, _ := te3File.FindEntWithProperty("type", "player")
	CurrentPlayer, _, _ := Players.New(ents.NewPlayer(playerSpawn.Position, playerSpawn.Angles))

	// Spawn enemies
	for _, spawn := range te3File.FindEntsWithProperty("type", "enemy") {
		Enemies.New(ents.NewEnemy(spawn.Position, spawn.Angles))
	}

	// UI
	fpsText, _ := ui.NewText("assets/textures/atlases/font.fnt", "FPS: 0")
	fpsText.SetDest(math2.Rect{X: 4.0, Y: 20.0, Width: 160.0, Height: 32.0})
	FPSCounter, _, _ := UI.Texts.New(fpsText)

	return &World{
		UI,
		Players,
		Enemies,
		GameMap,
		CurrentPlayer,
		FPSCounter,
	}, nil
}

func (w *World) Update(deltaTime float32) {
	// Update entities
	w.GameMap.Update(deltaTime)
	w.Players.Update((*ents.Player).Update, deltaTime)
	w.Enemies.Update((*ents.Enemy).Update, deltaTime)
	w.UI.Update(deltaTime)

	// Resolve collisions
	for i := 0; i < 3; i++ {
		bodiesIter := w.BodyIter()
		for body := bodiesIter(); body != nil; body = bodiesIter() {
			innerBodiesIter := w.BodyIter()
			for innerBody := innerBodiesIter(); innerBody != nil; innerBody = innerBodiesIter() {
				if innerBody != body {
					body.ResolveCollision(innerBody)
				}
			}
			w.GameMap.ResolveCollision(body)
		}
	}

	// Update FPS counter
	if fpsText, ok := w.UI.Texts.Get(w.FPSCounter); ok {
		fpsText.SetText(fmt.Sprintf("FPS: %v", engine.FPS()))
	}
}

func (w *World) Render() {
	// Find camera
	player, ok := w.Players.Get(w.CurrentPlayer)
	if !ok {
		panic("missing player")
	}
	cameraTransform := player.Body.Transform.Matrix()
	camera := player.Camera

	// Setup 3D game render context
	viewMat := cameraTransform.Inv()
	projMat := camera.GetProjectionMatrix()
	renderContext := render.Context{
		View:           viewMat,
		Projection:     projMat,
		FogStart:       1.0,
		FogLength:      50.0,
		LightDirection: mgl32.Vec3{1.0, 0.0, 1.0}.Normalize(),
		AmbientColor:   mgl32.Vec3{0.5, 0.5, 0.5},
	}

	// Render 3D game elements
	w.GameMap.Render(&renderContext)
	w.Enemies.Render((*ents.Enemy).Render, &renderContext)

	// Setup 2D render context
	renderContext = render.Context{
		View:       mgl32.Ident4(),
		Projection: mgl32.Ortho(0.0, settings.WINDOW_WIDTH, settings.WINDOW_HEIGHT, 0.0, -1.0, 10.0),
	}

	// Render 2D game elements
	w.UI.Render(&renderContext)
}

func (w *World) BodyIter() func() *comps.Body {
	playerIter := w.Players.Iter()
	enemiesIter := w.Enemies.Iter()
	return func() *comps.Body {
		if player := playerIter(); player != nil {
			return &player.Body
		}
		if enemy := enemiesIter(); enemy != nil {
			return &enemy.Body
		}
		return nil
	}
}
