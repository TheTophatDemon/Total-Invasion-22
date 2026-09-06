package world

import (
	"encoding/json"
	"fmt"
	"log"
	"math"
	"path/filepath"
	"strings"
	"time"

	"github.com/go-gl/mathgl/mgl32"
	"tophatdemon.com/total-invasion-ii/engine"
	"tophatdemon.com/total-invasion-ii/engine/assets/cache"
	"tophatdemon.com/total-invasion-ii/engine/assets/shaders"
	"tophatdemon.com/total-invasion-ii/engine/assets/te3"
	"tophatdemon.com/total-invasion-ii/engine/containers"
	"tophatdemon.com/total-invasion-ii/engine/failure"
	"tophatdemon.com/total-invasion-ii/engine/math2"
	"tophatdemon.com/total-invasion-ii/engine/math2/collision"
	"tophatdemon.com/total-invasion-ii/engine/render"
	"tophatdemon.com/total-invasion-ii/engine/scene"
	"tophatdemon.com/total-invasion-ii/engine/scene/comps"
	"tophatdemon.com/total-invasion-ii/engine/scene/tree"
	"tophatdemon.com/total-invasion-ii/engine/tdaudio"
	"tophatdemon.com/total-invasion-ii/game"
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
	Hud            Hud
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
	MapTitleKey    string
	CurrentPlayer  scene.Id[*Player]
	CurrentCamera  scene.Id[*Camera]
	removalQueue   []scene.Handle  // Holds entities to be removed at the end of the frame.
	app            engine.Observer // Communicates with the main application
	impendingLevel string          // Path to the next level. Set once the player reaches an exit.
	bspTree        tree.BspTree    // The BSP tree built in the previous frame.
	skyRender      comps.SkyRender
	frameBuffer    render.Framebuffer // Contains the rendered texture of the game
	hitCheckpoint  bool               // Turns true after the player hits a checkpoint
}

var gWorld *World

func spawnEntBasedOnType(ent te3.Ent, changeInfo game.MapChangeSignal) (entType string) {
	entType = ent.Properties["type"]
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
	case "chicken":
		_, _, err = SpawnChickenFromTE3(ent)
	case "player":
		gWorld.CurrentCamera, _, err = SpawnCameraFromTE3(ent)
		if err != nil {
			log.Printf("error spawning player camera: %v\n", err)
		}
		if changeInfo.PlayerEnt != nil {
			// Transfer properties from the previous level
			ent.Properties = changeInfo.PlayerEnt.Properties
			// But don't carry over keys
			ent.Properties["keys"] = "0"
		}
		gWorld.CurrentPlayer, _, err = SpawnPlayerFromTE3(ent, gWorld.CurrentCamera)
	}
	if err != nil {
		log.Printf("%v entity at %v caused an error: %v\n", entType, ent.GridPosition(), err)
	}
	return
}

func NewWorld(app engine.Observer, changeInfo game.MapChangeSignal) (*World, error) {
	gWorld = &World{
		removalQueue: make([]scene.Handle, 0, 8),
		app:          app,
	}

	gWorld.Hud.Init()
	// Include stats from save file if applicable
	gWorld.Hud.VictoryScreen.EnemiesKilled += changeInfo.KillCount
	gWorld.Hud.VictoryScreen.SecretsFound += changeInfo.SecretCount
	gWorld.Hud.VictoryScreen.levelStartTime = gWorld.Hud.VictoryScreen.levelStartTime.Add(-changeInfo.TimeSoFar)

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
	gWorld.MapLayers = scene.NewStorageWithFuncs(1, gWorld.UpdateMapLayer, (*comps.MapLayer).Render)

	te3File, err := te3.LoadTE3File(changeInfo.MapPath)
	if err != nil {
		return nil, err
	}

	_, gWorld.GameMap, err = gWorld.MapLayers.New()
	if err != nil {
		return nil, err
	}

	const texFlagInvisible = "invisible"
	*gWorld.GameMap, err = comps.NewMapLayer(te3File, ColLayerMap, []string{texFlagInvisible})
	if err != nil {
		return nil, err
	}

	// Process tiles after mesh is generated.
	for id, tile := range te3File.Tiles.Data {
		if tile.ShapeID < 0 {
			continue
		}

		primaryTexture := cache.GetTexture(te3File.Tiles.Textures[tile.TextureIDs[0]])

		layer := ColLayerMap
		if primaryTexture.HasFlag(texFlagInvisible) {
			layer = ColLayerInvisible
		} else if primaryTexture.HasFlag("killzone") {
			layer = ColLayerKillzone
		}

		// Set collision shapes
		shapeName := te3File.Tiles.Shapes[tile.ShapeID]

		shape, err := cache.GetCollisionShape(shapeName, tile.GetRotationMatrix())
		if err != nil {
			continue
		}

		gWorld.GameMap.GridShape.SetShapeAtFlatIndex(id, shape, layer)
	}

	if levelFileName := filepath.Base(changeInfo.MapPath); len(levelFileName) >= 4 &&
		levelFileName[0] == 'e' &&
		levelFileName[1] >= '0' && levelFileName[1] <= '9' &&
		levelFileName[2] == 'm' &&
		levelFileName[3] >= '0' && levelFileName[3] <= '9' {

		gWorld.MapTitleKey = levelFileName[0:4] + "Title"
		// Level files starting with e#m# activate the level intro.
		gWorld.Hud.Intro.Init(settings.Localize(gWorld.MapTitleKey), strings.ToUpper(levelFileName[0:4]))
	} else {
		gWorld.Hud.Intro.Init("", "")
	}

	// Spawn entities
	savedTypes := containers.NewSet[string](16)
	for _, ent := range changeInfo.SavedEnts {
		typ := spawnEntBasedOnType(ent, changeInfo)
		savedTypes.Add(typ)
	}

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

		if _, isSaved := savedTypes[ent.Properties["type"]]; isSaved {
			continue
		}

		spawnEntBasedOnType(ent, changeInfo)
	}

	// Create an autosave if this level is not being loaded from a save file already
	if changeInfo.SaveAfterLoad {
		app.ProcessSignal(game.SaveSignal{
			Number: 0,
		})
	}

	return gWorld, nil
}

func (world *World) Update(deltaTime float32) {
	world.removalQueue = world.removalQueue[0:0]

	// Create BSP tree
	it := world.IterBodies()
	world.bspTree = tree.BuildBspTree(scene.CollectSet(&it))

	// Update entities
	world.UpdateStores(deltaTime)

	// Set audio listener position
	if player, ok := world.CurrentPlayer.Get(); ok {
		world.Hud.Update(deltaTime, player)
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

	var push mgl32.Vec3
	grid := &world.GameMap.GridShape
	shrunkenShape := body.Shape.ShrunkenBy(0.1)
	sweepResult, _ := grid.SweepAgainst(mgl32.Vec3{}, body.Position, movement, shrunkenShape, filter)
	if sweepResult.Hit {
		speed := movement.Len()
		slide := sweepResult.Normal.Mul(-movement.Dot(sweepResult.Normal)).Add(movement)
		if slide != (mgl32.Vec3{}) {
			slide.Mul(speed / slide.Len())
		}
		push = push.Add(slide.Sub(movement))
	}

	pushVec := grid.PushOut(mgl32.Vec3{}, body.Position.Add(movement.Add(push)), body.Shape, filter)
	push = push.Add(pushVec)
	return push
}

func (world *World) Render() {
	effectiveScreenWidth := int(settings.Current.WindowWidth) / max(1, settings.Current.Pixelization)
	effectiveScreenHeight := int(settings.Current.WindowHeight) / max(1, settings.Current.Pixelization)
	shouldRegenFramebuffer := gWorld.frameBuffer.RenderTexture == nil ||
		gWorld.frameBuffer.RenderTexture.Width() != effectiveScreenWidth ||
		gWorld.frameBuffer.RenderTexture.Height() != effectiveScreenHeight
	if shouldRegenFramebuffer {
		if gWorld.frameBuffer.RenderTexture != nil {
			gWorld.frameBuffer.Free()
		}
		gWorld.frameBuffer = render.NewFramebuffer(effectiveScreenWidth, effectiveScreenHeight)
	}

	// Find camera
	camera, cameraExists := world.CurrentCamera.Get()
	if !cameraExists {
		failure.LogErrWithLocation("missing camera during rendering")
		return
	}

	// Setup 3D game render context
	viewMat := camera.Transform.Matrix().Inv()
	projMat := mgl32.Perspective(mgl32.DegToRad(float32(settings.Current.Fov)), settings.Current.WindowAspectRatio(), 0.1, 200.0)
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

	renderContext.Enable3D()
	gWorld.frameBuffer.Bind()

	// Render sky
	world.skyRender.Render(&renderContext)

	// Render 3D game elements
	world.RenderStores(&renderContext)
	renderContext.RenderTranslucentObjects()

	gWorld.frameBuffer.Unbind()

	world.Hud.Debug.UpdateCounters(&renderContext)
	world.Hud.Render(gWorld.frameBuffer.RenderTexture)
}

func (world *World) TearDown() {
	world.TearDownStores()
	world.frameBuffer.Free()
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
	camera, _ := world.CurrentCamera.Get()
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

	crossingBodies := world.bspTree.PotentiallyCrossingEnts(rayOrigin, rayDir, maxDist)
	for bodyId := range crossingBodies {
		bodyHaver, ok := bodyId.Get[comps.HasBody]()
		if !ok {
			continue
		}
		body := bodyHaver.Body()
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
		mapHit := layer.GridShape.Raycast(rayOrigin, rayDir, maxDist, filter)
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

func (world *World) ProcessSignal(signal any) {
	switch sig := signal.(type) {
	case game.ResumeGameSignal:
		playerIter := world.Players.Iter()
		for player, _ := playerIter.Next(); player != nil; player, _ = playerIter.Next() {
			player.ProcessSignal(signal)
		}
		world.Hud.ProcessSignal(signal)
	case game.SaveSignal:
		// Triggered when reaching a checkpoint
		playerIter := world.Players.Iter()
		for player, _ := playerIter.Next(); player != nil; player, _ = playerIter.Next() {
			player.ProcessSignal(signal)
		}
		if sig.AfterCheckpoint {
			world.hitCheckpoint = true
		}
		world.app.ProcessSignal(signal)
	}
}

func (world *World) MarshalJSON() ([]byte, error) {
	if world == nil || world.GameMap == nil {
		return nil, fmt.Errorf("game map is nil")
	}
	savablesIter := world.IterSavables()
	ents := make([]te3.Ent, 0, savablesIter.Capacity())
	for savable, _ := savablesIter.Next(); savable != nil; savable, _ = savablesIter.Next() {
		ents = append(ents, savable.Save())
	}
	return json.Marshal(game.MapChangeSignal{
		MapPath:         world.GameMap.Name,
		MapTitleKey:     world.MapTitleKey,
		SavedEnts:       ents,
		Timestamp:       time.Now(),
		KillCount:       world.Hud.VictoryScreen.EnemiesKilled,
		SecretCount:     world.Hud.VictoryScreen.SecretsFound,
		TimeSoFar:       time.Since(world.Hud.VictoryScreen.levelStartTime),
		AfterCheckpoint: world.hitCheckpoint,
		DifficultyIndex: settings.Current.DifficultyIndex,
	})
}
