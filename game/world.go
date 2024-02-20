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
	"tophatdemon.com/total-invasion-ii/engine/scene"
	"tophatdemon.com/total-invasion-ii/engine/scene/comps"
	"tophatdemon.com/total-invasion-ii/engine/scene/comps/ui"
	"tophatdemon.com/total-invasion-ii/game/ents"
	"tophatdemon.com/total-invasion-ii/game/settings"
)

const (
	MESSAGE_FADE_SPEED = 2.0
	DEFAULT_FONT_PATH  = "assets/textures/ui/font.fnt"
)

type World struct {
	UI                        *ui.Scene
	Players                   *scene.Storage[ents.Player]
	Enemies                   *scene.Storage[ents.Enemy]
	Walls                     *scene.Storage[ents.Wall]
	Props                     *scene.Storage[ents.Prop]
	Triggers                  *scene.Storage[ents.Trigger]
	GameMap                   *scene.Map
	CurrentPlayer             scene.Id[ents.Player]
	FPSCounter, SpriteCounter scene.Id[ui.Text]
	messageText               scene.Id[ui.Text]
	messageTimer              float32
	messagePriority           int
	flashRect                 scene.Id[ui.Box]
	flashSpeed                float32
}

var _ ents.WorldOps = (*World)(nil)

func NewWorld(mapPath string) (*World, error) {
	w := &World{
		messageTimer:    2.0,
		messagePriority: 0,
	}

	w.UI = ui.NewUIScene(256, 64)
	w.Players = scene.NewStorage[ents.Player](8)
	w.Enemies = scene.NewStorage[ents.Enemy](256)
	w.Walls = scene.NewStorage[ents.Wall](256)
	w.Props = scene.NewStorage[ents.Prop](256)
	w.Triggers = scene.NewStorage[ents.Trigger](64)

	te3File, err := te3.LoadTE3File(mapPath)
	if err != nil {
		return nil, err
	}

	w.GameMap, err = scene.NewMap(te3File)
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
	w.CurrentPlayer, _, _ = ents.SpawnPlayer(w.Players, w, playerSpawn.Position, playerSpawn.Angles)

	// Spawn enemies
	for _, spawn := range te3File.FindEntsWithProperty("type", "enemy") {
		ents.SpawnEnemy(w.Enemies, spawn.Position, spawn.Angles)
	}

	// Spawn dynamic tiles
	for _, spawn := range te3File.FindEntsWithProperty("type", "door") {
		if _, _, err := ents.SpawnWallFromTE3(w.Walls, w, spawn); err != nil {
			log.Printf("entity at %v caused an error: %v\n", spawn.Position, err)
		}
	}

	// Spawn props
	for _, spawn := range te3File.FindEntsWithProperty("type", "prop") {
		if _, _, err := ents.SpawnPropFromTE3(w.Props, w, spawn); err != nil {
			log.Printf("entity at %v caused an error: %v\n", spawn.Position, err)
		}
	}

	// Spawn triggers
	for _, spawn := range te3File.FindEntsWithProperty("type", "trigger") {
		if _, _, err := ents.SpawnTriggerFromTE3(w.Triggers, w, spawn); err != nil {
			log.Printf("entity at %v caused an error: %v\n", spawn.Position, err)
		}
	}

	// UI
	var fpsText *ui.Text
	w.FPSCounter, fpsText, _ = w.UI.Texts.New()
	fpsText.
		SetFont(DEFAULT_FONT_PATH).
		SetText("FPS: 0").
		SetDest(math2.Rect{X: 4.0, Y: 20.0, Width: 160.0, Height: 32.0}).
		SetScale(1.0).
		SetColor(color.White)

	var spriteCounter *ui.Text
	w.SpriteCounter, spriteCounter, _ = w.UI.Texts.New()
	spriteCounter.
		SetFont(DEFAULT_FONT_PATH).
		SetText("Sprites drawn: 0\nWalls drawn: 0").
		SetDest(math2.Rect{X: 4.0, Y: 56.0, Width: 320.0, Height: 64.0}).
		SetScale(1.0).
		SetColor(color.Blue)

	var message *ui.Text
	w.messageText, message, _ = w.UI.Texts.New()
	message.
		SetFont(DEFAULT_FONT_PATH).
		SetText("This is a test message! Это подопытное сообщение!").
		SetDest(math2.Rect{
			X:      float32(settings.UI_WIDTH) / 3.0,
			Y:      float32(settings.UI_HEIGHT) / 2.0,
			Width:  float32(settings.UI_WIDTH / 3.0),
			Height: float32(settings.UI_HEIGHT) / 2.0,
		}).
		SetAlignment(ui.TEXT_ALIGN_CENTER).
		SetColor(color.Red).
		SetScale(1.0)

	var flashBox *ui.Box
	w.flashRect, flashBox, _ = w.UI.Boxes.New()
	flashBox.
		SetDest(math2.Rect{
			X:      -settings.WINDOW_WIDTH,
			Y:      -settings.WINDOW_HEIGHT,
			Width:  settings.WINDOW_WIDTH * 2,
			Height: settings.WINDOW_HEIGHT * 2,
		}).
		SetColor(color.Blue.WithAlpha(0.5))
	w.flashSpeed = 0.5

	return w, nil
}

func (w *World) Update(deltaTime float32) {
	// Update entities
	w.GameMap.Update(deltaTime)
	w.Players.Update((*ents.Player).Update, deltaTime)
	w.Enemies.Update((*ents.Enemy).Update, deltaTime)
	w.Walls.Update((*ents.Wall).Update, deltaTime)
	w.Props.Update((*ents.Prop).Update, deltaTime)
	w.Triggers.Update((*ents.Trigger).Update, deltaTime)
	w.UI.Update(deltaTime)

	// Update bodies and resolve collisions
	bodiesIter := w.BodiesIter()
	for bodyEnt, _ := bodiesIter(); bodyEnt != nil; bodyEnt, _ = bodiesIter() {
		body := bodyEnt.Body()
		before := body.Transform.Position()
		body.Update(deltaTime)

		if before.Sub(body.Transform.Position()).LenSqr() != 0.0 {
			innerBodiesIter := w.BodiesIter()
			for innerBodyEnt, _ := innerBodiesIter(); innerBodyEnt != nil; innerBodyEnt, _ = innerBodiesIter() {
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
	if message, ok := w.messageText.Get(); ok {
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

	// Update screen flash
	if flash, ok := w.flashRect.Get(); ok {
		flash.Color = flash.Color.Fade(w.flashSpeed * deltaTime)
	}

	// Update FPS counter
	if fpsText, ok := w.FPSCounter.Get(); ok {
		fpsText.SetText(fmt.Sprintf("FPS: %v", engine.FPS()))
	}
}

func (w *World) Render() {
	// Find camera
	player, ok := w.CurrentPlayer.Get()
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
		AspectRatio:    settings.WINDOW_ASPECT_RATIO,
		FogStart:       1.0,
		FogLength:      50.0,
		LightDirection: mgl32.Vec3{1.0, 0.0, 1.0}.Normalize(),
		AmbientColor:   mgl32.Vec3{0.5, 0.5, 0.5},
	}

	// Render 3D game elements
	w.GameMap.Render(&renderContext)
	w.Enemies.Render((*ents.Enemy).Render, &renderContext)
	w.Walls.Render((*ents.Wall).Render, &renderContext)
	w.Props.Render((*ents.Prop).Render, &renderContext)

	if sprCountTxt, ok := w.SpriteCounter.Get(); ok {
		sprCountTxt.SetText(fmt.Sprintf("Sprites drawn: %v\nWalls drawn: %v", renderContext.DrawnSpriteCount, renderContext.DrawnWallCount))
	}

	uiMargin := ((float32(settings.WINDOW_WIDTH) * (float32(settings.UI_HEIGHT) / float32(settings.WINDOW_HEIGHT))) - float32(settings.UI_WIDTH)) / 2.0

	// Setup 2D render context
	renderContext = render.Context{
		View:       mgl32.Ident4(),
		Projection: mgl32.Ortho(-uiMargin, settings.UI_WIDTH+uiMargin, settings.UI_HEIGHT, 0.0, -1.0, 10.0),
	}

	// Render 2D game elements
	w.UI.Render(&renderContext)
}

func (w *World) ShowMessage(text string, duration float32, priority int, colr color.Color) {
	if priority >= w.messagePriority {
		w.messageTimer = duration
		w.messagePriority = priority
		if message, ok := w.messageText.Get(); ok {
			message.SetText(text).SetColor(colr)
		}
	}
}

func (w *World) FlashScreen(color color.Color, fadeSpeed float32) {
	if flash, ok := w.flashRect.Get(); ok {
		flash.Color = color
		w.flashSpeed = fadeSpeed
	}
}

func (w *World) Raycast(rayOrigin, rayDir mgl32.Vec3, includeBodies bool, maxDist float32, excludeBody comps.HasBody) (collision.RaycastResult, scene.Handle) {
	rayBB := math2.BoxFromPoints(rayOrigin, rayOrigin.Add(rayDir.Mul(maxDist)))
	mapHit := w.GameMap.CastRay(rayOrigin, rayDir, maxDist)
	if includeBodies {
		var closestEnt scene.Handle
		var closestBodyHit collision.RaycastResult
		closestBodyHit.Distance = math.MaxFloat32
		nextBody := w.BodiesIter()
		for bodyEnt, bodyId := nextBody(); bodyEnt != nil; bodyEnt, bodyId = nextBody() {
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
				closestEnt = bodyId
			}
		}
		if !closestEnt.IsNil() {
			return closestBodyHit, closestEnt
		}
	}
	return mapHit, scene.Handle{}
}

func (w *World) ActorsInSphere(spherePos mgl32.Vec3, sphereRadius float32, exception ents.HasActor) []scene.Handle {
	radiusSq := sphereRadius * sphereRadius
	nextActor := w.ActorsIter()
	result := make([]scene.Handle, 0)
	for actorEnt, actorId := nextActor(); actorEnt != nil; actorEnt, actorId = nextActor() {
		if actorEnt == exception {
			continue
		}
		body := actorEnt.Body()
		if body.Transform.Position().Sub(spherePos).LenSqr() < radiusSq {
			result = append(result, actorId)
		}
	}
	return result
}

func (w *World) BodiesInSphere(spherePos mgl32.Vec3, sphereRadius float32, exception comps.HasBody) []scene.Handle {
	nextBody := w.BodiesIter()
	result := make([]scene.Handle, 0)
	for bodyEnt, bodyId := nextBody(); bodyEnt != nil; bodyEnt, bodyId = nextBody() {
		if bodyEnt == exception {
			continue
		}
		body := bodyEnt.Body()

		var hit bool
		switch shape := body.Shape.(type) {
		case collision.Sphere:
			hit = collision.SphereTouchesSphere(spherePos, sphereRadius, body.Transform.Position(), shape.Radius())
		case collision.Box:
			hit = collision.SphereTouchesBox(spherePos, sphereRadius, shape.Extents().Translate(body.Transform.Position()))
		case collision.Mesh:
			for _, tri := range shape.Mesh().Triangles() {
				if h, _ := collision.SphereTriangleCollision(spherePos, sphereRadius, tri, body.Transform.Position()); h != collision.TRIHIT_NONE {
					hit = true
					break
				}
			}
		}
		if hit {
			result = append(result, bodyId)
		}
	}
	return result
}

func (w *World) AddUiBox(box ui.Box) (scene.Id[ui.Box], error) {
	id, b, err := w.UI.Boxes.New()
	if err != nil {
		return scene.Id[ui.Box]{}, err
	}
	*b = box
	return id, nil
}
