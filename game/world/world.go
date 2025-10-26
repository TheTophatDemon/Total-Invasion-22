package world

import (
	"errors"
	"log"
	"math"
	"path"
	"strings"

	"github.com/go-gl/mathgl/mgl32"
	"tophatdemon.com/total-invasion-ii/engine"
	"tophatdemon.com/total-invasion-ii/engine/assets/cache"
	"tophatdemon.com/total-invasion-ii/engine/assets/shaders"
	"tophatdemon.com/total-invasion-ii/engine/assets/te3"
	"tophatdemon.com/total-invasion-ii/engine/failure"
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
	ColLayerNone collision.Mask = 0
	ColLayerMap  collision.Mask = 1 << (iota - 1)
	ColLayerActors
	ColLayerProjectiles
	ColLayerInvisible // Includes invisible walls around holes and lava
	ColLayerPlayers
	ColLayerNPCs     // Includes enemies, chickens, and Geoffrey
	ColLayerKillzone // Kills any actor that touches
	ColLayerUsable   // Marks entities that can have the use key applied to interact with them
)

//go:generate go run ../../cmd/world_gen_iters/world_gen_iters.go
type World struct {
	Hud            hud.Hud
	Players        scene.Storage[Player]
	Enemies        scene.Storage[Enemy]
	Chickens       scene.Storage[Chicken]
	Walls          scene.Storage[Wall]
	Triggers       scene.Storage[Trigger]
	Projectiles    scene.Storage[Projectile]
	Effects        scene.Storage[Effect]
	Items          scene.Storage[Item]
	DebugShapes    scene.Storage[DebugShape]
	Cameras        scene.Storage[Camera]
	MapLayers      scene.Storage[comps.MapLayer]
	Props          scene.Storage[Prop]
	GameMap        *comps.MapLayer // An easy access pointer to the main map layer
	InvisibleLayer *comps.MapLayer
	CurrentPlayer  scene.Id[*Player]
	CurrentCamera  scene.Id[*Camera]
	removalQueue   []scene.Handle  // Holds entities to be removed at the end of the frame.
	app            engine.Observer // Communicates with the main application
	impendingLevel string          // Path to the next level. Set once the player reaches an exit.
	bspTree        tree.BspTree    // The BSP tree built in the previous frame.
	skyRender      comps.SkyRender
}

var gWorld *World

func NewWorld(app engine.Observer, mapPath string, changeInfo game.MapChangeSignal) (*World, error) {
	gWorld = &World{
		removalQueue: make([]scene.Handle, 0, 8),
		app:          app,
	}

	gWorld.Hud.Init()

	gWorld.Players = scene.NewStorageWithFuncs(8, (*Player).Update, (*Player).Render)
	gWorld.Enemies = scene.NewStorageWithFuncs(256, (*Enemy).Update, (*Enemy).Render)
	gWorld.Chickens = scene.NewStorageWithFuncs(64, (*Chicken).Update, (*Chicken).Render)
	gWorld.Walls = scene.NewStorageWithFuncs(256, (*Wall).Update, (*Wall).Render)
	gWorld.Props = scene.NewStorageWithFuncs(256, (*Prop).Update, (*Prop).Render)
	gWorld.Triggers = scene.NewStorageWithFuncs(512, (*Trigger).Update, (*Trigger).Render)
	gWorld.Projectiles = scene.NewStorageWithFuncs(256, (*Projectile).Update, (*Projectile).Render)
	gWorld.Effects = scene.NewStorageWithFuncs(256, (*Effect).Update, (*Effect).Render)
	gWorld.Items = scene.NewStorageWithFuncs(256, (*Item).Update, (*Item).Render)
	gWorld.DebugShapes = scene.NewStorageWithFuncs(128, (*DebugShape).Update, (*DebugShape).Render)
	gWorld.Cameras = scene.NewStorageWithFuncs(64, (*Camera).Update, nil)
	gWorld.MapLayers = scene.NewStorageWithFuncs(3, gWorld.UpdateMapLayer, (*comps.MapLayer).Render)

	te3File, err := te3.LoadTE3File(mapPath)
	if err != nil {
		return nil, err
	}

	var err2 error
	_, gWorld.GameMap, err = gWorld.MapLayers.New()
	_, gWorld.InvisibleLayer, err2 = gWorld.MapLayers.New()
	_, killLayer, err3 := gWorld.MapLayers.New()
	if err = errors.Join(err, err2, err3); err != nil {
		return nil, err
	}

	const texFlagInvisible = "invisible"
	*gWorld.GameMap, err = comps.NewMainMapLayer(te3File, ColLayerMap, []string{texFlagInvisible})
	if err != nil {
		return nil, err
	}

	*gWorld.InvisibleLayer = comps.NewExtraMapLayer(te3File, ColLayerInvisible)
	*killLayer = comps.NewExtraMapLayer(te3File, ColLayerKillzone)

	// Process tiles after mesh is generated.
	for id, tile := range te3File.Tiles.Data {
		if tile.ShapeID < 0 {
			continue
		}

		primaryTexture := cache.GetTexture(te3File.Tiles.Textures[tile.TextureIDs[0]])

		mapLayer := gWorld.GameMap
		if primaryTexture.HasFlag(texFlagInvisible) {
			mapLayer = gWorld.InvisibleLayer
		} else if primaryTexture.HasFlag("killzone") {
			mapLayer = killLayer
		} else if primaryTexture.HasFlag("liquid") {
			// Remove collision from liquid tiles.
			continue
		}

		// Set collision shapes
		shape, err := cache.GetCollisionShape(te3File.Tiles.Shapes[tile.ShapeID], tile.GetRotationMatrix())
		if err != nil {
			continue
		}

		mapLayer.GridShape.SetShapeAtFlatIndex(id, shape)
	}

	if levelFileName := path.Base(mapPath); len(levelFileName) >= 4 &&
		levelFileName[0] == 'e' &&
		levelFileName[1] >= '0' && levelFileName[1] <= '9' &&
		levelFileName[2] == 'm' &&
		levelFileName[3] >= '0' && levelFileName[3] <= '9' {

		// Level files starting with e#m# activate the level intro.
		gWorld.Hud.Intro.Init(settings.Localize(levelFileName[0:4]+"Title"), strings.ToUpper(levelFileName[0:4]))
	} else {
		gWorld.Hud.Intro.Init("", "")
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
				tdaudio.QueueSong("assets/music/"+songPath+".ogg", true, 0)
			}

			if skyPath, hasSky := ent.Properties["sky"]; hasSky {
				// Create sky model
				skyMesh, meshErr := cache.GetMesh("assets/models/sky.obj")
				skyTex := cache.GetTexture("assets/textures/skies/" + skyPath + ".png")
				if meshErr != nil {
					failure.LogErrWithLocation("Error loading sky: %v\n", meshErr)
				} else {
					gWorld.skyRender = comps.NewSkyRender(skyMesh, shaders.SkyShader, skyTex)
				}
			}

			continue
		}

		entType := ent.Properties["type"]
		var err error
		switch entType {
		case "enemy":
			_, _, err = SpawnEnemyFromTE3(ent)
		case WallTypeDoor, WallTypePushWall, WallTypeSwitch:
			_, _, err = SpawnWallFromTE3(ent)
		case "prop":
			_, _, err = SpawnPropFromTE3(ent)
		case "trigger":
			_, _, err = SpawnTriggerFromTE3(ent)
		case "item":
			_, _, err = SpawnItemFromTE3(ent)
		case "camera":
			_, _, err = SpawnCameraFromTE3(ent)
		case "player":
			gWorld.CurrentCamera, _, err = SpawnCameraFromTE3(ent)
			if err != nil {
				log.Printf("error spawning player camera: %v\n", err)
			}
			gWorld.CurrentPlayer, _, err = SpawnPlayer(ent.Position, ent.Angles, gWorld.CurrentCamera, changeInfo)
		}
		if err != nil {
			log.Printf("%v entity at %v caused an error: %v\n", entType, ent.GridPosition(), err)
		}
	}

	return gWorld, nil
}

func (world *World) Update(deltaTime float32) {
	world.removalQueue = world.removalQueue[0:0]

	// Free mouse
	if input.IsActionJustPressed(settings.ActionTrapMouse) {
		if input.IsMouseTrapped() {
			input.UntrapMouse()
		} else {
			input.TrapMouse()
		}
	}

	// Create BSP tree
	{
		it := world.IterBodies()
		world.bspTree = tree.BuildBspTree(scene.CollectSet(&it))
	}

	// Update entities
	world.UpdateStores(deltaTime)
	world.Hud.Update(deltaTime)

	// Set audio listener position
	if player, ok := world.CurrentPlayer.Get(); ok {
		pos := player.actor.Position()
		dir := player.actor.FacingVec()
		tdaudio.SetListenerOrientation(pos[0], pos[1], pos[2], dir[0], dir[1], dir[2])
	}

	// Remove deleted entities
	for _, handle := range world.removalQueue {
		handle.Remove()
	}
}

func (world *World) UpdateMapLayer(layer *comps.MapLayer, deltaTime float32) {
	layer.Update(deltaTime)

	// Damage any actors touching the killzone layer
	if layer.Layer == ColLayerKillzone {
		actorsIter := world.IterActors()
		for ent, _ := actorsIter.Next(); ent != nil; ent, _ = actorsIter.Next() {
			actor := ent.Actor()
			if layer.GridShape.OtherBodyTouches(mgl32.Vec3{}, actor.body.Position, actor.body.Shape) {
				ent.OnDamage(layer, math2.Inf32())
			}
		}
	}
}

func (world *World) ResolveMapCollisions(
	body *comps.Body,
	movement mgl32.Vec3,
	lockY bool,
	filter collision.Mask,
) mgl32.Vec3 {
	if body == nil {
		return mgl32.Vec3{}
	}

	layerIt := world.MapLayers.Iter()
	push := mgl32.Vec3{}
	for layer, _ := layerIt.Next(); layer != nil; layer, _ = layerIt.Next() {
		nextPos := body.Position.Add(movement).Add(push)
		if (layer.Layer & filter) != 0 {
			pushVec := layer.GridShape.PushOut(mgl32.Vec3{}, nextPos, body.Shape)
			push = push.Add(pushVec)
		}
	}
	if lockY {
		push[1] = 0.0
	}
	return push
}

func (world *World) Render() {
	// Find camera
	camera, cameraExists := world.CurrentCamera.Get()
	if !cameraExists {
		failure.LogErrWithLocation("missing camera during rendering")
		return
	}

	// Setup 3D game render context
	viewMat := camera.Transform.Matrix().Inv()
	projMat := camera.ProjectionMatrix()
	renderContext := render.Context{
		View:           viewMat,
		ViewInverse:    viewMat.Inv(),
		Projection:     projMat,
		AspectRatio:    settings.Current.WindowAspectRatio(),
		FogStart:       1.0,
		FogLength:      50.0,
		LightDirection: mgl32.Vec3{1.0, 0.0, 1.0}.Normalize(),
		AmbientColor:   mgl32.Vec3{0.5, 0.5, 0.5},
	}

	if world.Hud.Intro.TimeLeft() < 2.0 {
		// Render sky
		world.skyRender.Render(&renderContext)

		// Render 3D game elements
		world.RenderStores(&renderContext)
		renderContext.RenderTranslucentObjects()
	}

	world.Hud.Debug.UpdateCounters(&renderContext)
	if player, playerExists := world.CurrentPlayer.Get(); playerExists && (world.CurrentCamera.Equals(player.Camera.Handle) || world.InWinState()) {
		world.Hud.Render()
	}
}

func (world *World) TearDown() {
	world.TearDownStores()
}

func (world *World) QueueRemoval(entHandle scene.Handle) {
	world.removalQueue = append(world.removalQueue, entHandle)
}

func (world *World) InWinState() bool {
	return len(world.impendingLevel) != 0
}

func (world *World) EnterWinState(nextLevel string, winCamera scene.Handle) {
	world.impendingLevel = nextLevel
	world.CurrentCamera = scene.Id[*Camera]{Handle: winCamera}
	camera, _ := scene.Get[*Camera](world.CurrentCamera.Handle)
	camera.waitTime = 0.0
	tdaudio.QueueSong("assets/music/viktor_the_victor.ogg", false, 0.0)
	world.Hud.VictoryScreen.EndLevel()
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

func (world *World) Raycast(
	rayOrigin, rayDir mgl32.Vec3, filter collision.Mask,
	maxDist float32, excludeBody *comps.Body,
) (collision.Result, scene.Handle) {
	var rayBB math2.Box = math2.BoxFromPoints(rayOrigin, rayOrigin.Add(rayDir.Mul(maxDist)))
	var closestEnt scene.Handle
	var closestHit collision.Result
	closestHit.Distance = math.MaxFloat32

	// Check bodies
	//TODO: Could optimize this using bsp tree
	iter := world.IterBodies()
	for bodyEnt, bodyId := iter.Next(); bodyEnt != nil; bodyEnt, bodyId = iter.Next() {
		body := bodyEnt.Body()
		if body == nil || body == excludeBody ||
			!body.OnLayer(filter) ||
			!body.Shape.Extents().Translate(body.Position).Intersects(rayBB) {
			continue
		}
		bodyHit := body.Shape.Raycast(body.Position, rayOrigin, rayDir, maxDist)
		if bodyHit.Hit && bodyHit.Distance < closestHit.Distance {
			closestHit = bodyHit
			closestEnt = bodyId
		}
	}

	// Check map layers
	layerIt := world.MapLayers.Iter()
	for layer, layerId := layerIt.Next(); layer != nil; layer, _ = layerIt.Next() {
		if !layer.Layer.On(filter) {
			continue
		}
		mapHit := layer.GridShape.Raycast(rayOrigin, rayDir, maxDist)
		if mapHit.Hit && mapHit.Distance < closestHit.Distance {
			closestHit = mapHit
			closestEnt = layerId
		}
	}

	if !closestEnt.IsNil() {
		return closestHit, closestEnt
	}
	return collision.Result{}, scene.Handle{}
}

// Returns an iterator over all linkables with the given non-zero link number.
func (world *World) NextLinkableWithNumber(iter *LinkablesIter, linkNumber int) (Linkable, scene.Handle) {
	if linkNumber == 0 {
		return nil, scene.Handle{}
	}

	for ent, id := iter.Next(); ent != nil; ent, id = iter.Next() {
		if ent.LinkNumber() == linkNumber {
			return ent, id
		}
	}
	return nil, scene.Handle{}
}

func (world *World) ActivateLinks(source Linkable) {
	iter := world.IterLinkables()
	for {
		ent, handle := world.NextLinkableWithNumber(&iter, source.LinkNumber())
		if ent == nil {
			break
		}
		if !handle.Equals(source.Handle()) {
			ent.OnLinkActivate(source)
		}
	}
}

func (world *World) DeactivateLinks(source Linkable) {
	iter := world.IterLinkables()
	for {
		ent, handle := world.NextLinkableWithNumber(&iter, source.LinkNumber())
		if ent == nil {
			break
		}
		if !handle.Equals(source.Handle()) {
			ent.OnLinkDeactivate(source)
		}
	}
}
