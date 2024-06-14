package world

import (
	"errors"
	"fmt"
	"log"

	"github.com/go-gl/mathgl/mgl32"
	"tophatdemon.com/total-invasion-ii/engine"
	"tophatdemon.com/total-invasion-ii/engine/assets/cache"
	"tophatdemon.com/total-invasion-ii/engine/assets/te3"
	"tophatdemon.com/total-invasion-ii/engine/color"
	"tophatdemon.com/total-invasion-ii/engine/math2"
	"tophatdemon.com/total-invasion-ii/engine/math2/collision"
	"tophatdemon.com/total-invasion-ii/engine/render"
	"tophatdemon.com/total-invasion-ii/engine/scene"
	"tophatdemon.com/total-invasion-ii/engine/scene/comps"
	"tophatdemon.com/total-invasion-ii/engine/scene/comps/ui"
	"tophatdemon.com/total-invasion-ii/game/settings"
)

const (
	MESSAGE_FADE_SPEED = 2.0
	DEFAULT_FONT_PATH  = "assets/textures/ui/font.fnt"
)

const (
	COL_LAYER_NONE collision.Mask = 0
	COL_LAYER_MAP  collision.Mask = 1 << (iota - 1)
	COL_LAYER_ACTORS
	COL_LAYER_PROJECTILES
	COL_LAYER_INVISIBLE
)

const (
	COL_FILTER_FOR_ACTORS collision.Mask = COL_LAYER_MAP | COL_LAYER_ACTORS | COL_LAYER_INVISIBLE
)

const (
	TEX_FLAG_INVISIBLE string = "invisible"
)

//go:generate ../../world_gen_iters.exe engine/scene/comps.HasBody HasActor Linkable
type World struct {
	UI                        *ui.Scene
	Players                   scene.Storage[Player]
	Enemies                   scene.Storage[Enemy]
	Walls                     scene.Storage[Wall]
	Props                     scene.Storage[Prop]
	Triggers                  scene.Storage[Trigger]
	Projectiles               scene.Storage[Projectile]
	DebugShapes               scene.Storage[DebugShape]
	GameMap                   comps.Map
	CurrentPlayer             scene.Id[*Player]
	FPSCounter, SpriteCounter scene.Id[*ui.Text]
	messageText               scene.Id[*ui.Text]
	messageTimer              float32
	messagePriority           int
	flashRect                 scene.Id[*ui.Box]
	flashSpeed                float32
}

func NewWorld(mapPath string) (*World, error) {
	world := &World{
		messageTimer:    2.0,
		messagePriority: 0,
	}

	world.UI = ui.NewUIScene(256, 64)

	world.Players = scene.NewStorageWithFuncs(8, (*Player).Update, nil)
	world.Enemies = scene.NewStorageWithFuncs(256, (*Enemy).Update, (*Enemy).Render)
	world.Walls = scene.NewStorageWithFuncs(256, (*Wall).Update, (*Wall).Render)
	world.Props = scene.NewStorageWithFuncs(256, (*Prop).Update, (*Prop).Render)
	world.Triggers = scene.NewStorageWithFuncs(64, (*Trigger).Update, nil)
	world.Projectiles = scene.NewStorageWithFuncs(256, (*Projectile).Update, (*Projectile).Render)
	world.DebugShapes = scene.NewStorageWithFuncs(128, (*DebugShape).Update, (*DebugShape).Render)

	te3File, err := te3.LoadTE3File(mapPath)
	if err != nil {
		return nil, err
	}

	// Filter out invisible tiles and spawn invisible wall entities instead
	for texID, texPath := range te3File.Tiles.Textures {
		if cache.GetTexture(texPath).HasFlag(TEX_FLAG_INVISIBLE) {
			invisibleTileIDs := te3File.Tiles.WithTextureId(te3.TextureID(texID))
			te3File.Tiles.EraseTiles(invisibleTileIDs...)
			for _, id := range invisibleTileIDs {
				box := te3File.Tiles.BBoxOfTile(te3File.Tiles.UnflattenGridPos(id))
				pos := box.Center()
				SpawnInvisibleWall(&world.Walls, world, pos, collision.NewBox(box.Translate(pos.Mul(-1.0))))
			}
		}
	}

	world.GameMap, err = comps.NewMap(te3File, COL_LAYER_MAP)
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
			world.GameMap.SetTileCollisionShapesForAngles(shapeName, 0, 45, 0, 360, panelShapeX),
			world.GameMap.SetTileCollisionShapesForAngles(shapeName, 45, 135, 0, 360, panelShapeZ),
			world.GameMap.SetTileCollisionShapesForAngles(shapeName, 135, 225, 0, 360, panelShapeX),
			world.GameMap.SetTileCollisionShapesForAngles(shapeName, 225, 315, 0, 360, panelShapeZ),
			world.GameMap.SetTileCollisionShapesForAngles(shapeName, 315, 360, 0, 360, panelShapeX),
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
		err = world.GameMap.SetTileCollisionShapes(shapeName, collision.NewBox(math2.BoxFromRadius(1.0)))
		if err != nil {
			return nil, err
		}
	}

	// Read level properties
	levelProps, _ := te3File.FindEntWithProperty("name", "level properties")
	if songPath, hasSong := levelProps.Properties["song"]; hasSong {
		// Play the song
		song, err := cache.GetSong(songPath)
		if err == nil {
			song.Play()
		}
	}

	// Spawn player
	playerSpawn, _ := te3File.FindEntWithProperty("type", "player")
	world.CurrentPlayer, _, _ = SpawnPlayer(&world.Players, world, playerSpawn.Position, playerSpawn.Angles)

	// Spawn enemies
	for _, spawn := range te3File.FindEntsWithProperty("type", "enemy") {
		SpawnEnemy(&world.Enemies, spawn.Position, spawn.Angles)
	}

	// Spawn dynamic tiles
	for _, spawn := range te3File.FindEntsWithProperty("type", "door") {
		if _, _, err := SpawnWallFromTE3(&world.Walls, world, spawn); err != nil {
			log.Printf("entity at %v caused an error: %v\n", spawn.Position, err)
		}
	}

	// Spawn props
	for _, spawn := range te3File.FindEntsWithProperty("type", "prop") {
		if _, _, err := SpawnPropFromTE3(&world.Props, world, spawn); err != nil {
			log.Printf("entity at %v caused an error: %v\n", spawn.Position, err)
		}
	}

	// Spawn triggers
	for _, spawn := range te3File.FindEntsWithProperty("type", "trigger") {
		if _, _, err := SpawnTriggerFromTE3(&world.Triggers, world, spawn); err != nil {
			log.Printf("entity at %v caused an error: %v\n", spawn.Position, err)
		}
	}

	// UI
	var fpsText *ui.Text
	world.FPSCounter, fpsText, _ = world.UI.Texts.New()
	fpsText.
		SetFont(DEFAULT_FONT_PATH).
		SetText("FPS: 0").
		SetDest(math2.Rect{X: 4.0, Y: 20.0, Width: 160.0, Height: 32.0}).
		SetScale(1.0).
		SetColor(color.White)

	var spriteCounter *ui.Text
	world.SpriteCounter, spriteCounter, _ = world.UI.Texts.New()
	spriteCounter.
		SetFont(DEFAULT_FONT_PATH).
		SetText("Sprites drawn: 0\nWalls drawn: 0\nParticles drawn: 0").
		SetDest(math2.Rect{X: 4.0, Y: 56.0, Width: 320.0, Height: 64.0}).
		SetScale(1.0).
		SetColor(color.Blue)

	var message *ui.Text
	world.messageText, message, _ = world.UI.Texts.New()
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
		SetShadow(settings.Current.TextShadowColor, mgl32.Vec2{2.0, 2.0})

	var flashBox *ui.Box
	world.flashRect, flashBox, _ = world.UI.Boxes.New()
	flashBox.
		SetDest(math2.Rect{
			X:      -float32(settings.Current.WindowWidth),
			Y:      -float32(settings.Current.WindowHeight),
			Width:  float32(settings.Current.WindowWidth) * 2,
			Height: float32(settings.Current.WindowHeight) * 2,
		}).
		SetColor(color.Blue.WithAlpha(0.5))
	world.flashSpeed = 0.5

	return world, nil
}

func (world *World) Update(deltaTime float32) {
	// Update entities
	scene.UpdateStores(world, deltaTime)
	world.UI.Update(deltaTime)

	// Update bodies and resolve collisions
	bodiesIter := world.BodyIter()
	for bodyEnt, _ := bodiesIter(); bodyEnt != nil; bodyEnt, _ = bodiesIter() {
		bodyEnt.Body().MoveAndCollide(deltaTime, world.BodyIter())
	}

	// Update message text
	if message, ok := world.messageText.Get(); ok {
		if world.messageTimer > 0.0 {
			world.messageTimer -= deltaTime
		} else {
			message.SetColor(message.Color().Fade(deltaTime * MESSAGE_FADE_SPEED))
			if message.Color().A <= 0.0 {
				message.SetColor(color.Transparent)
				world.messageTimer = 0.0
				world.messagePriority = 0
				message.SetText("")
			}
		}
	}

	// Update screen flash
	if flash, ok := world.flashRect.Get(); ok {
		flash.Color = flash.Color.Fade(world.flashSpeed * deltaTime)
	}

	// Update FPS counter
	if fpsText, ok := world.FPSCounter.Get(); ok {
		fpsText.SetText(fmt.Sprintf("FPS: %v", engine.FPS()))
	}
}

func (world *World) Render() {
	// Find camera
	player, ok := world.CurrentPlayer.Get()
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
		AspectRatio:    settings.Current.WindowAspectRatio(),
		FogStart:       1.0,
		FogLength:      50.0,
		LightDirection: mgl32.Vec3{1.0, 0.0, 1.0}.Normalize(),
		AmbientColor:   mgl32.Vec3{0.5, 0.5, 0.5},
	}

	// Render 3D game elements
	scene.RenderStores(world, &renderContext)

	if sprCountTxt, ok := world.SpriteCounter.Get(); ok {
		sprCountTxt.SetText(
			fmt.Sprintf("Sprites drawn: %v\nWalls drawn: %v\nParticles drawn: %v",
				renderContext.DrawnSpriteCount,
				renderContext.DrawnWallCount,
				renderContext.DrawnParticlesCount))
	}

	uiMargin := ((float32(settings.Current.WindowWidth) * (float32(settings.UI_HEIGHT) / float32(settings.Current.WindowHeight))) - float32(settings.UI_WIDTH)) / 2.0

	// Setup 2D render context
	renderContext = render.Context{
		View:       mgl32.Ident4(),
		Projection: mgl32.Ortho(-uiMargin, settings.UI_WIDTH+uiMargin, settings.UI_HEIGHT, 0.0, -1.0, 10.0),
	}

	// Render 2D game elements
	world.UI.Render(&renderContext)
}

func (world *World) ShowMessage(text string, duration float32, priority int, colr color.Color) {
	if priority >= world.messagePriority {
		world.messageTimer = duration
		world.messagePriority = priority
		if message, ok := world.messageText.Get(); ok {
			message.SetText(text).SetColor(colr)
		}
	}
}

func (world *World) FlashScreen(color color.Color, fadeSpeed float32) {
	if flash, ok := world.flashRect.Get(); ok {
		flash.Color = color
		world.flashSpeed = fadeSpeed
	}
}

func (world *World) ListenerTransform() mgl32.Mat4 {
	if player, ok := world.CurrentPlayer.Get(); ok {
		return player.Body().Transform.Matrix()
	}
	return mgl32.Ident4()
}
