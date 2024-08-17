package world

import (
	"log"

	"github.com/go-gl/mathgl/mgl32"
	"tophatdemon.com/total-invasion-ii/engine"
	"tophatdemon.com/total-invasion-ii/engine/assets/cache"
	"tophatdemon.com/total-invasion-ii/engine/assets/te3"
	"tophatdemon.com/total-invasion-ii/engine/input"
	"tophatdemon.com/total-invasion-ii/engine/math2"
	"tophatdemon.com/total-invasion-ii/engine/math2/collision"
	"tophatdemon.com/total-invasion-ii/engine/render"
	"tophatdemon.com/total-invasion-ii/engine/scene"
	"tophatdemon.com/total-invasion-ii/engine/scene/comps"
	"tophatdemon.com/total-invasion-ii/engine/tdaudio"
	"tophatdemon.com/total-invasion-ii/game"
	"tophatdemon.com/total-invasion-ii/game/hud"
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
	COL_LAYER_INVISIBLE // Includes invisible walls around holes and lava
	COL_LAYER_PLAYERS
	COL_LAYER_NPCS // Includes enemies, chickens, and Geoffrey
)

const (
	COL_FILTER_FOR_ACTORS collision.Mask = COL_LAYER_MAP | COL_LAYER_ACTORS | COL_LAYER_INVISIBLE
)

const (
	TEX_FLAG_INVISIBLE string = "invisible"
)

//go:generate ../../world_gen_iters.exe engine/scene/comps.HasBody HasActor Linkable
type World struct {
	Hud           hud.Hud
	Players       scene.Storage[Player]
	Enemies       scene.Storage[Enemy]
	Walls         scene.Storage[Wall]
	Props         scene.Storage[Prop]
	Triggers      scene.Storage[Trigger]
	Projectiles   scene.Storage[Projectile]
	DebugShapes   scene.Storage[DebugShape]
	GameMap       comps.Map
	CurrentPlayer scene.Id[*Player]
	removalQueue  []scene.Handle  // Holds entities to be removed at the end of the frame.
	app           engine.Observer // Communicates with the main application
}

func NewWorld(app engine.Observer, mapPath string) (*World, error) {
	world := &World{
		removalQueue: make([]scene.Handle, 0, 8),
		app:          app,
	}

	world.Hud.Init()

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
		world.GameMap.SetTileCollisionShapesForAngles(shapeName, 0, 45, 0, 360, panelShapeX)
		world.GameMap.SetTileCollisionShapesForAngles(shapeName, 45, 135, 0, 360, panelShapeZ)
		world.GameMap.SetTileCollisionShapesForAngles(shapeName, 135, 225, 0, 360, panelShapeX)
		world.GameMap.SetTileCollisionShapesForAngles(shapeName, 225, 315, 0, 360, panelShapeZ)
		world.GameMap.SetTileCollisionShapesForAngles(shapeName, 315, 360, 0, 360, panelShapeX)
	}

	// Set cube collision shapes
	for _, shapeName := range [...]string{
		"assets/models/shapes/cube.obj",
		"assets/models/shapes/cube_2tex.obj",
		"assets/models/shapes/edge_panel.obj",
		"assets/models/shapes/cube_marker.obj",
		"assets/models/shapes/bridge.obj",
	} {
		world.GameMap.SetTileCollisionShapes(shapeName, collision.NewBox(math2.BoxFromRadius(1.0)))
	}

	// Read level properties
	levelProps, _ := te3File.FindEntWithProperty("name", "level properties")
	if songPath, hasSong := levelProps.Properties["song"]; hasSong {
		// Play the song
		tdaudio.QueueSong(songPath, true, 0)
	}

	// Spawn player
	playerSpawn, _ := te3File.FindEntWithProperty("type", "player")
	world.CurrentPlayer, _, _ = SpawnPlayer(&world.Players, world, playerSpawn.Position, playerSpawn.Angles)

	// Spawn enemies
	for _, spawn := range te3File.FindEntsWithProperty("type", "enemy") {
		SpawnEnemy(&world.Enemies, spawn.Position, spawn.Angles, world)
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

	return world, nil
}

func (world *World) ChangeMap(mapPath string) {
	world.app.ProcessSignal(game.MapChangeSignal{
		NextMapPath: mapPath,
	})
}

func (world *World) Update(deltaTime float32) {
	world.removalQueue = world.removalQueue[0:0]

	// Free mouse
	if input.IsActionJustPressed(settings.ACTION_TRAP_MOUSE) {
		if input.IsMouseTrapped() {
			input.UntrapMouse()
		} else {
			input.TrapMouse()
		}
	}

	// Update entities
	scene.UpdateStores(world, deltaTime)
	world.Hud.Update(deltaTime)

	// Set audio listener position
	if player, ok := world.CurrentPlayer.Get(); ok {
		pos := player.actor.Position()
		dir := player.actor.FacingVec()
		tdaudio.SetListenerOrientation(pos[0], pos[1], pos[2], dir[0], dir[1], dir[2])
	}

	// Update bodies and resolve collisions
	bodiesIter := world.BodyIter()
	for bodyEnt, _ := bodiesIter(); bodyEnt != nil; bodyEnt, _ = bodiesIter() {
		bodyEnt.Body().MoveAndCollide(deltaTime, world.BodyIter())
	}

	// Remove deleted entities
	for _, handle := range world.removalQueue {
		handle.Remove()
	}
}

func (world *World) Render() {
	// Find camera
	player, ok := world.CurrentPlayer.Get()
	if !ok {
		panic("missing player")
	}
	camera := player.Camera
	cameraTransform := player.Body().Transform.Matrix().Mul4(camera.Transform.Matrix())

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

	world.Hud.UpdateDebugCounters(&renderContext)
	world.Hud.Render()
}

func (world *World) TearDown() {
	scene.TearDownStores(world)
}

func (world *World) QueueRemoval(entHandle scene.Handle) {
	world.removalQueue = append(world.removalQueue, entHandle)
}
