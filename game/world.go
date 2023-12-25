package game

import (
	"errors"
	"fmt"
	"log"
	"math"

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
	UI                        *ui.Scene
	Players                   *world.Storage[ents.Player]
	Enemies                   *world.Storage[ents.Enemy]
	Walls                     *world.Storage[ents.Wall]
	GameMap                   *world.Map
	CurrentPlayer             world.Id[ents.Player]
	FPSCounter, SpriteCounter world.Id[ui.Text]
	Message                   world.Id[ui.Text]
	messageTimer              float32
	messagePriority           int
}

func NewWorld(mapPath string) (*World, error) {
	w := &World{
		messageTimer:    2.0,
		messagePriority: 0,
	}

	w.UI = ui.NewUIScene(256, 64)
	w.Players = world.NewStorage[ents.Player](8)
	w.Enemies = world.NewStorage[ents.Enemy](256)
	w.Walls = world.NewStorage[ents.Wall](256)

	te3File, err := te3.LoadTE3File(mapPath)
	if err != nil {
		return nil, err
	}

	w.GameMap, err = world.NewMap(te3File)
	if err != nil {
		return nil, err
	}

	// Set panel collision shapes
	panelShapeX := collision.NewBox(math2.BoxFromExtents(1.0, 1.0, 0.5))
	panelShapeZ := collision.NewBox(math2.BoxFromExtents(0.5, 1.0, 1.0))
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
		err = w.GameMap.SetTileCollisionShapes(shapeName, collision.NewBox(math2.BoxFromRadius(1.0)))
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

	// Spawn dynamic tiles
	for _, spawn := range te3File.FindEntsWithProperty("type", "door") {
		if wall, err := ents.NewWallFromTE3(spawn, w); err == nil {
			w.Walls.New(wall)
		} else {
			log.Printf("entity at %v caused an error: %v\n", spawn.Position, err)
		}
	}

	// UI
	fpsText, _ := ui.NewText("assets/textures/atlases/font.fnt", "FPS: 0")
	fpsText.SetDest(math2.Rect{X: 4.0, Y: 20.0, Width: 160.0, Height: 32.0})
	w.FPSCounter, _, _ = w.UI.Texts.New(fpsText)

	var spriteCounter *ui.Text
	w.SpriteCounter, spriteCounter, _ = w.UI.Texts.New(ui.Text{})
	spriteCounter.
		SetText("Sprites drawn: 0").
		SetDest(math2.Rect{X: 4.0, Y: 56.0, Width: 320.0, Height: 32.0}).
		SetScale(1.0).
		SetColor(color.Blue).
		SetFont("assets/textures/atlases/font.fnt")

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
	w.Walls.Update((*ents.Wall).Update, deltaTime)
	w.UI.Update(deltaTime)

	// Update bodies and resolve collisions
	bodiesIter := w.BodyIter()
	for bodyEnt := bodiesIter(); bodyEnt != nil; bodyEnt = bodiesIter() {
		body := bodyEnt.Body()
		before := body.Transform.Position()
		body.Update(deltaTime)

		if before.Sub(body.Transform.Position()).LenSqr() != 0.0 {
			innerBodiesIter := w.BodyIter()
			for innerBodyEnt := innerBodiesIter(); innerBodyEnt != nil; innerBodyEnt = innerBodiesIter() {
				if innerBodyEnt != body {
					body.ResolveCollision(innerBodyEnt.Body())
				}
			}
		}
		w.GameMap.ResolveCollision(body)

		// Restrict movement to the XZ plane
		after := body.Transform.Position()
		body.Transform.SetPosition(mgl32.Vec3{after.X(), before.Y(), after.Z()})
	}

	// Update message text
	if message, ok := w.Message.Get(); ok {
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
	cameraTransform := player.Body().Transform.Matrix()
	camera := player.Camera

	// Setup 3D game render context
	viewMat := cameraTransform.Inv()
	projMat := camera.ProjectionMatrix()
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
	w.Walls.Render((*ents.Wall).Render, &renderContext)

	if sprCountTxt, ok := w.SpriteCounter.Get(); ok {
		sprCountTxt.SetText(fmt.Sprintf("Sprites drawn: %v", renderContext.DrawnSpriteCount))
	}

	// Setup 2D render context
	renderContext = render.Context{
		View:       mgl32.Ident4(),
		Projection: mgl32.Ortho(0.0, settings.WINDOW_WIDTH, settings.WINDOW_HEIGHT, 0.0, -1.0, 10.0),
	}

	// Render 2D game elements
	w.UI.Render(&renderContext)
}

func (w *World) BodyIter() func() comps.HasBody {
	playerIter := w.Players.Iter()
	enemiesIter := w.Enemies.Iter()
	wallsIter := w.Walls.Iter()
	return func() comps.HasBody {
		if player := playerIter(); player != nil {
			return player
		}
		if enemy := enemiesIter(); enemy != nil {
			return enemy
		}
		if wall := wallsIter(); wall != nil {
			return wall
		}
		return nil
	}
}

func (w *World) ShowMessage(text string, duration float32, priority int, colr color.Color) {
	if priority >= w.messagePriority {
		w.messageTimer = duration
		w.messagePriority = priority
		if message, ok := w.Message.Get(); ok {
			message.SetText(text).SetColor(colr)
		}
	}
}

func (w *World) Raycast(rayOrigin, rayDir mgl32.Vec3, includeBodies bool, maxDist float32, excludeBody comps.HasBody) (collision.RaycastResult, comps.HasBody) {
	rayBB := math2.BoxFromPoints(rayOrigin, rayOrigin.Add(rayDir.Mul(maxDist)))
	mapHit := w.GameMap.CastRay(rayOrigin, rayDir)
	var closestEnt comps.HasBody
	var closestBodyHit collision.RaycastResult
	closestBodyHit.Distance = math.MaxFloat32
	nextBody := w.BodyIter()
	for bodyEnt := nextBody(); bodyEnt != nil; bodyEnt = nextBody() {
		body := bodyEnt.Body()
		if bodyEnt == excludeBody || !body.Shape.Extents().Translate(body.Transform.Position()).Intersects(rayBB) {
			continue
		}
		bodyHit := body.Shape.Raycast(rayOrigin, rayDir, body.Transform.Position())
		if mapHit.Hit && bodyHit.Distance > mapHit.Distance {
			continue
		}
		if bodyHit.Hit && bodyHit.Distance < closestBodyHit.Distance {
			closestBodyHit = bodyHit
			closestEnt = bodyEnt
		}
	}
	if closestEnt != nil {
		return closestBodyHit, closestEnt
	}
	return mapHit, nil
}
