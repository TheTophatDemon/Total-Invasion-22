package game

import (
	"errors"
	"fmt"

	"github.com/go-gl/mathgl/mgl32"
	"tophatdemon.com/total-invasion-ii/engine"
	"tophatdemon.com/total-invasion-ii/engine/assets/te3"
	"tophatdemon.com/total-invasion-ii/engine/color"
	"tophatdemon.com/total-invasion-ii/engine/math2"
	"tophatdemon.com/total-invasion-ii/engine/math2/collision"
	"tophatdemon.com/total-invasion-ii/engine/render"
	"tophatdemon.com/total-invasion-ii/engine/world"
	"tophatdemon.com/total-invasion-ii/engine/world/comps"
	"tophatdemon.com/total-invasion-ii/engine/world/comps/ui"
	"tophatdemon.com/total-invasion-ii/game/ents"
	"tophatdemon.com/total-invasion-ii/game/settings"
)

const (
	MESSAGE_FADE_SPEED = 2.0
)

type World struct {
	UI              *ui.Scene
	Players         *world.Storage[ents.Player]
	Enemies         *world.Storage[ents.Enemy]
	GameMap         *world.Map
	CurrentPlayer   world.Id[ents.Player]
	FPSCounter      world.Id[ui.Text]
	Message         world.Id[ui.Text]
	messageTimer    float32
	messagePriority int
}

func NewWorld(mapPath string) (*World, error) {
	w := &World{
		messageTimer:    2.0,
		messagePriority: 0,
	}

	w.UI = ui.NewUIScene(256, 64)
	w.Players = world.NewStorage[ents.Player](8)
	w.Enemies = world.NewStorage[ents.Enemy](256)

	te3File, err := te3.LoadTE3File(mapPath)
	if err != nil {
		return nil, err
	}

	w.GameMap, err = world.NewMap(te3File)
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
			w.GameMap.SetTileCollisionShapesForAngles(shapeName, 0, 45, 0, 360, panelShapeX),
			w.GameMap.SetTileCollisionShapesForAngles(shapeName, 45, 135, 0, 360, panelShapeZ),
			w.GameMap.SetTileCollisionShapesForAngles(shapeName, 135, 225, 0, 360, panelShapeX),
			w.GameMap.SetTileCollisionShapesForAngles(shapeName, 225, 315, 0, 360, panelShapeZ),
			w.GameMap.SetTileCollisionShapesForAngles(shapeName, 315, 360, 0, 360, panelShapeX),
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
		err = w.GameMap.SetTileCollisionShapes(shapeName, collision.ShapeBox(math2.BoxFromRadius(1.0)))
		if err != nil {
			return nil, err
		}
	}

	// Spawn player
	playerSpawn, _ := te3File.FindEntWithProperty("type", "player")
	w.CurrentPlayer, _, _ = w.Players.New(ents.NewPlayer(playerSpawn.Position, playerSpawn.Angles, w))

	// Spawn enemies
	for _, spawn := range te3File.FindEntsWithProperty("type", "enemy") {
		w.Enemies.New(ents.NewEnemy(spawn.Position, spawn.Angles))
	}

	// UI
	fpsText, _ := ui.NewText("assets/textures/atlases/font.fnt", "FPS: 0")
	fpsText.SetDest(math2.Rect{X: 4.0, Y: 20.0, Width: 160.0, Height: 32.0})
	w.FPSCounter, _, _ = w.UI.Texts.New(fpsText)

	message, _ := ui.NewText("assets/textures/atlases/font.fnt", "This is a test message!")
	message.SetDest(math2.Rect{
		X:      float32(settings.WINDOW_WIDTH) / 3.0,
		Y:      float32(settings.WINDOW_HEIGHT) / 2.0,
		Width:  float32(settings.WINDOW_WIDTH / 3.0),
		Height: float32(settings.WINDOW_HEIGHT) / 2.0,
	}).SetAlignment(ui.TEXT_ALIGN_CENTER).SetColor(color.Red)
	w.Message, _, _ = w.UI.Texts.New(message)

	return w, nil
}

func (w *World) Update(deltaTime float32) {
	// Update entities
	w.GameMap.Update(deltaTime)
	w.Players.Update((*ents.Player).Update, deltaTime)
	w.Enemies.Update((*ents.Enemy).Update, deltaTime)
	w.UI.Update(deltaTime)

	// Update bodies and resolve collisions
	bodiesIter := w.BodyIter()
	for body := bodiesIter(); body != nil; body = bodiesIter() {
		before := body.Transform.Position()
		body.Update(deltaTime)

		if before.Sub(body.Transform.Position()).LenSqr() != 0.0 {
			innerBodiesIter := w.BodyIter()
			for innerBody := innerBodiesIter(); innerBody != nil; innerBody = innerBodiesIter() {
				if innerBody != body {
					body.ResolveCollision(innerBody)
				}
			}
		}
		w.GameMap.ResolveCollision(body)

		// Restrict movement to the XZ plane
		after := body.Transform.Position()
		body.Transform.SetPosition(mgl32.Vec3{after.X(), before.Y(), after.Z()})
	}

	// Update message text
	if message, ok := w.Message.Get().(*ui.Text); ok {
		if w.messageTimer > 0.0 {
			w.messageTimer -= deltaTime
		} else {
			message.SetColor(message.Color().Fade(deltaTime * MESSAGE_FADE_SPEED))
			if message.Color().A <= 0.0 {
				message.SetColor(color.Transparent)
				w.messageTimer = 0.0
				w.messagePriority = 0
				message.SetText("")
			}
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

func (w *World) ShowMessage(text string, duration float32, priority int, colr color.Color) {
	if priority >= w.messagePriority {
		w.messageTimer = duration
		w.messagePriority = priority
		if message, ok := w.Message.Get().(*ui.Text); ok {
			message.SetText(text).SetColor(colr)
		}
	}
}
