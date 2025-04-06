package world

import (
	"log"
	"time"

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
	"tophatdemon.com/total-invasion-ii/engine/scene/tree"
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
	TEX_FLAG_KILLZONE         = "killzone"
	TEX_FLAG_LIQUID           = "liquid"
)

//go:generate go run ../../cmd/world_gen_iters/world_gen_iters.go
type World struct {
	Hud              hud.Hud
	Players          scene.Storage[Player]
	Enemies          scene.Storage[Enemy]
	Chickens         scene.Storage[Chicken]
	Walls            scene.Storage[Wall]
	Props            scene.Storage[Prop]
	Triggers         scene.Storage[Trigger]
	Projectiles      scene.Storage[Projectile]
	Effects          scene.Storage[Effect]
	Items            scene.Storage[Item]
	DebugShapes      scene.Storage[DebugShape]
	Cameras          scene.Storage[Camera]
	GameMaps         scene.Storage[comps.Map]
	GameMap          *comps.Map
	CurrentPlayer    scene.Id[*Player]
	CurrentCamera    scene.Id[*Camera]
	removalQueue     []scene.Handle  // Holds entities to be removed at the end of the frame.
	app              engine.Observer // Communicates with the main application
	nextLevel        string          // Path to the next level. Set once the player reaches an exit.
	bspTree          tree.BspTree    // The BSP tree built in the previous frame.
	avgCollisionTime int64
}

func NewWorld(app engine.Observer, mapPath string) (*World, error) {
	world := &World{
		removalQueue: make([]scene.Handle, 0, 8),
		app:          app,
	}

	world.Hud.Init()

	world.Players = scene.NewStorageWithFuncs(8, (*Player).Update, (*Player).Render)
	world.Enemies = scene.NewStorageWithFuncs(256, (*Enemy).Update, (*Enemy).Render)
	world.Chickens = scene.NewStorageWithFuncs(64, (*Chicken).Update, (*Chicken).Render)
	world.Walls = scene.NewStorageWithFuncs(256, (*Wall).Update, (*Wall).Render)
	world.Props = scene.NewStorageWithFuncs(256, (*Prop).Update, (*Prop).Render)
	world.Triggers = scene.NewStorageWithFuncs(256, (*Trigger).Update, (*Trigger).Render)
	world.Projectiles = scene.NewStorageWithFuncs(256, (*Projectile).Update, (*Projectile).Render)
	world.Effects = scene.NewStorageWithFuncs(256, (*Effect).Update, (*Effect).Render)
	world.Items = scene.NewStorageWithFuncs(256, (*Item).Update, (*Item).Render)
	world.DebugShapes = scene.NewStorageWithFuncs(128, (*DebugShape).Update, (*DebugShape).Render)
	world.Cameras = scene.NewStorageWithFuncs(64, (*Camera).Update, nil)
	world.GameMaps = scene.NewStorageWithFuncs(1, (*comps.Map).Update, (*comps.Map).Render)

	te3File, err := te3.LoadTE3File(mapPath)
	if err != nil {
		return nil, err
	}

	// Pre process tiles before mesh is generated.
	for texID, texPath := range te3File.Tiles.Textures {
		tex := cache.GetTexture(texPath)
		for id, tile := range te3File.Tiles.Data {
			if tile.TextureIDs[0] != te3.TextureID(texID) {
				continue
			}
			box := te3File.Tiles.BBoxOfTile(te3File.Tiles.UnflattenGridPos(id))
			pos := box.Center()
			// Remove invisible tiles and spawn entities in their place
			if tex.HasFlag(TEX_FLAG_INVISIBLE) {
				te3File.Tiles.EraseTile(id)
				SpawnInvisibleWall(world, pos, collision.NewBox(box.Translate(pos.Mul(-1.0))))
			}
			if tex.HasFlag(TEX_FLAG_KILLZONE) {
				SpawnKillzone(world, pos, box.Size().Len()/2.0, 9999.0)
			}
		}
	}

	_, world.GameMap, err = world.GameMaps.New()
	if err != nil {
		return nil, err
	}
	*world.GameMap, err = comps.NewMap(te3File, COL_LAYER_MAP)
	if err != nil {
		return nil, err
	}

	type transformedShape struct {
		shapeName  string
		yaw, pitch uint8
	}
	// Contains the collision meshes of shapes at various rotations for reuse.
	transformedShapesCache := make(map[transformedShape]collision.Mesh)

	// Process tiles after mesh is generated.
	for id, tile := range te3File.Tiles.Data {
		if tile.ShapeID < 0 {
			continue
		}

		if cache.GetTexture(te3File.Tiles.Textures[tile.TextureIDs[0]]).HasFlag(TEX_FLAG_LIQUID) {
			// Remove collision from liquid tiles.
			world.GameMap.GridShape.SetShapeAtFlatIndex(id, nil)
			continue
		}

		// Set collision shapes
		switch shapeName := te3File.Tiles.Shapes[tile.ShapeID]; shapeName {
		case "assets/models/shapes/corner.obj",
			"assets/models/shapes/cylinder.obj",
			"assets/models/shapes/right_tetrahedron.obj",
			"assets/models/shapes/tetrahedron_transition.obj",
			"assets/models/shapes/wedge_corner_inner.obj",
			"assets/models/shapes/wedge_corner_outer.obj",
			"assets/models/shapes/wedge.obj":

			// Triangles
			cacheKey := transformedShape{shapeName, tile.Yaw, tile.Pitch}
			trianglesShape, ok := transformedShapesCache[cacheKey]
			if !ok {
				shapeMesh, err := cache.GetMesh(shapeName)
				if err != nil {
					log.Printf("error loading mesh for collisions shape of %v: %v\n", shapeName, err)
					continue
				}
				transform := tile.GetRotationMatrix()
				rawTrianglesIter := shapeMesh.IterTriangles()
				transformedTriangles := make([]math2.Triangle, rawTrianglesIter.Count())
				// Transform the vertices of the triangle according to the tile's orientation.
				for i := 0; rawTrianglesIter.HasNext(); i++ {
					rawTriangle := rawTrianglesIter.Next()
					for p := range rawTriangle {
						transformedTriangles[i][p] = mgl32.TransformNormal(rawTriangle[p], transform)
					}
				}
				trianglesShape = collision.NewMeshFromTriangles(transformedTriangles)
				transformedShapesCache[cacheKey] = trianglesShape
			}

			world.GameMap.GridShape.SetShapeAtFlatIndex(id, trianglesShape)
		case "assets/models/shapes/bars.obj",
			"assets/models/shapes/panel.obj":

			// Panel
			var panelShape collision.Shape
			switch tile.Yaw {
			case 0, 2:
				panelShape = collision.NewBox(math2.BoxFromExtents(1.0, 1.0, 0.5))
			case 1, 3:
				panelShape = collision.NewBox(math2.BoxFromExtents(0.5, 1.0, 1.0))
			}
			world.GameMap.GridShape.SetShapeAtFlatIndex(id, panelShape)
		default:
			// Box
			world.GameMap.GridShape.SetShapeAtFlatIndex(id, collision.NewBox(math2.BoxFromRadius(1.0)))
		}
	}

	// Spawn entities
	for _, ent := range te3File.Ents {
		if ent.Properties == nil {
			continue
		}

		// Read level properties
		if ent.Properties["name"] == "level properties" {
			if songPath, hasSong := ent.Properties["song"]; hasSong {
				// Play the song
				tdaudio.QueueSong(songPath, true, 0)
			}
			continue
		}

		entType := ent.Properties["type"]
		var err error
		switch entType {
		case "enemy":
			_, _, err = SpawnEnemyFromTE3(world, ent)
		case "door", "switch":
			_, _, err = SpawnWallFromTE3(world, ent)
		case "prop":
			_, _, err = SpawnPropFromTE3(world, ent)
		case "trigger":
			_, _, err = SpawnTriggerFromTE3(world, ent)
		case "item":
			_, _, err = SpawnItemFromTE3(world, ent)
		case "camera":
			_, _, err = SpawnCameraFromTE3(world, ent)
		case "player":
			world.CurrentCamera, _, err = SpawnCameraFromTE3(world, ent)
			if err != nil {
				log.Printf("error spawning player camera: %v\n", err)
			}
			world.CurrentPlayer, _, err = SpawnPlayer(world, ent.Position, ent.Angles, world.CurrentCamera)
		}
		if err != nil {
			log.Printf("%v entity at %v caused an error: %v\n", entType, ent.GridPosition(), err)
		}
	}

	world.Hud.LevelStartTime = time.Now()
	return world, nil
}

func (world *World) ChangeMap(mapPath string) {
	world.app.ProcessSignal(game.MapChangeSignal{
		NextMapPath: mapPath,
	})
}

func (world *World) Update(deltaTime float32) {
	world.removalQueue = world.removalQueue[0:0]

	if input.IsActionJustPressed(settings.ACTION_KILL_ENEMIES) {
		iter := world.IterActors()
		for actor, handle := iter.Next(); actor != nil; actor, handle = iter.Next() {
			if !handle.Equals(world.CurrentPlayer.Handle) {
				actor.Actor().Health = 0
			}
		}
	}

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

	startTime := time.Now()
	// Update bodies and resolve collisions
	it := world.IterBodies()
	world.bspTree = tree.BuildBspTree(&it, world.GameMap)
	it = world.IterBodies()
	for {
		bodyEnt, _ := it.Next()
		if bodyEnt == nil {
			break
		}

		innerIter := comps.BodySliceIter{
			Slice: world.bspTree.PotentiallyTouchingEnts(bodyEnt.Body().Transform.Position(), bodyEnt.Body().Shape),
		}

		// The game map should be excluded from the bvh tree due to its large size.
		mapIter := world.GameMaps.Iter()
		_, mapHandle := mapIter.Next()
		innerIter.Slice = append(innerIter.Slice, mapHandle)

		bodyEnt.Body().MoveAndCollide(deltaTime, &innerIter)
	}

	duration := time.Now().Sub(startTime).Milliseconds()
	if world.avgCollisionTime != 0 {
		world.avgCollisionTime = (world.avgCollisionTime + duration) / 2
	} else {
		world.avgCollisionTime = duration
	}

	// Remove deleted entities
	for _, handle := range world.removalQueue {
		handle.Remove()
	}
}

func (world *World) Render() {
	// Find camera
	camera, cameraExists := world.CurrentCamera.Get()
	if !cameraExists {
		log.Println("Error: missing camera during rendering")
		return
	}

	// Setup 3D game render context
	viewMat := camera.Transform.Matrix().Inv()
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

	world.Hud.UpdateDebugCounters(&renderContext, world.avgCollisionTime)
	if player, playerExists := world.CurrentPlayer.Get(); playerExists && (world.CurrentCamera.Equals(player.Camera.Handle) || world.InWinState()) {
		world.Hud.Render()
	}
}

func (world *World) TearDown() {
	scene.TearDownStores(world)
}

func (world *World) QueueRemoval(entHandle scene.Handle) {
	world.removalQueue = append(world.removalQueue, entHandle)
}

func (world *World) InWinState() bool {
	return len(world.nextLevel) != 0
}

func (world *World) EnterWinState(nextLevel string, winCamera scene.Handle) {
	world.nextLevel = nextLevel
	world.CurrentCamera = scene.Id[*Camera]{Handle: winCamera}
	tdaudio.QueueSong("assets/music/viktor_the_victor.ogg", false, 0.0)
	world.Hud.LevelEndTime = time.Now()
	world.Hud.InitVictory()
}

func (world *World) ResetToPlayerCamera() {
	if player, isPlayer := world.CurrentPlayer.Get(); isPlayer {
		world.CurrentCamera = player.Camera
	}
}

func (world *World) IsOnPlayerCamera() bool {
	if player, isPlayer := world.CurrentPlayer.Get(); isPlayer {
		return world.CurrentCamera.Handle.Equals(player.Camera.Handle)
	}
	return false
}
