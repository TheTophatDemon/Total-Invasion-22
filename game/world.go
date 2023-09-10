package game

import (
	"fmt"

	"github.com/go-gl/mathgl/mgl32"
	"tophatdemon.com/total-invasion-ii/engine"
	"tophatdemon.com/total-invasion-ii/engine/assets/te3"
	"tophatdemon.com/total-invasion-ii/engine/math2"
	"tophatdemon.com/total-invasion-ii/engine/render"
	"tophatdemon.com/total-invasion-ii/engine/world"
	"tophatdemon.com/total-invasion-ii/engine/world/comps/ui"
	"tophatdemon.com/total-invasion-ii/game/ents"
	"tophatdemon.com/total-invasion-ii/game/settings"
)

type World struct {
	UI            *ui.Scene
	Players       *world.Storage[ents.Player]
	GameMap       *world.Map
	CurrentPlayer world.Id[ents.Player]
	FPSCounter    world.Id[ui.Text]
}

func NewWorld(mapPath string) (*World, error) {
	UI := ui.NewUIScene(256, 64)
	Players := world.NewStorage[ents.Player](8)

	te3File, err := te3.LoadTE3File(mapPath)
	if err != nil {
		return nil, err
	}

	GameMap, err := world.NewMap(te3File)
	if err != nil {
		return nil, err
	}

	// Spawn player
	playerSpawn, _ := te3File.FindEntWithProperty("type", "player")
	CurrentPlayer, _, _ := Players.New(ents.NewPlayer(playerSpawn.Position, playerSpawn.Angles))

	// UI
	fpsText, _ := ui.NewText("assets/textures/atlases/font.fnt", "FPS: 0")
	fpsText.SetDest(math2.Rect{X: 4.0, Y: 20.0, Width: 160.0, Height: 32.0})
	FPSCounter, _, _ := UI.Texts.New(fpsText)

	return &World{
		UI,
		Players,
		GameMap,
		CurrentPlayer,
		FPSCounter,
	}, nil
}

func (w *World) Update(deltaTime float32) {
	// Update entities
	w.GameMap.Update(deltaTime)
	w.Players.Update((*ents.Player).Update, deltaTime)
	w.UI.Update(deltaTime)

	// Resolve collisions
	w.Players.ForEach(func(p *ents.Player) {
		_ = w.GameMap.ResolveCollision(&p.Body)
	})

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

	// Setup 2D render context
	renderContext = render.Context{
		View:       mgl32.Ident4(),
		Projection: mgl32.Ortho(0.0, settings.WINDOW_WIDTH, settings.WINDOW_HEIGHT, 0.0, -1.0, 10.0),
	}

	// Render 2D game elements
	w.UI.Render(&renderContext)
}
