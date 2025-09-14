package world

import (
	"errors"
	"fmt"
	"log"
	"maps"
	"math"
	"path"
	"slices"
	"strings"
	"time"

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
)

const (
	ColFilterForActors collision.Mask = ColLayerMap | ColLayerActors | ColLayerInvisible
)

//go:generate go run ../../cmd/world_gen_iters/world_gen_iters.go
type World struct {
	Hud              hud.Hud
	Players          scene.Storage[Player]
	Enemies          scene.Storage[Enemy]
	Chickens         scene.Storage[Chicken]
	Walls            scene.Storage[Wall]
	Triggers         scene.Storage[Trigger]
	Projectiles      scene.Storage[Projectile]
	Effects          scene.Storage[Effect]
	Items            scene.Storage[Item]
	DebugShapes      scene.Storage[DebugShape]
	Cameras          scene.Storage[Camera]
	MapLayers        scene.Storage[comps.MapLayer]
	Props            scene.Storage[Prop]
	GameMap          *comps.MapLayer
	CurrentPlayer    scene.Id[*Player]
	CurrentCamera    scene.Id[*Camera]
	removalQueue     []scene.Handle  // Holds entities to be removed at the end of the frame.
	app              engine.Observer // Communicates with the main application
	impendingLevel   string          // Path to the next level. Set once the player reaches an exit.
	bspTree          tree.BspTree    // The BSP tree built in the previous frame.
	avgCollisionTime int64           // Average number of milliseconds spent per frame solving collisions.
	tickCount        int64
	skyRender        comps.SkyRender
}

func NewWorld(app engine.Observer, mapPath string, changeInfo game.MapChangeSignal) (*World, error) {
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
	world.Triggers = scene.NewStorageWithFuncs(512, (*Trigger).Update, (*Trigger).Render)
	world.Projectiles = scene.NewStorageWithFuncs(256, (*Projectile).Update, (*Projectile).Render)
	world.Effects = scene.NewStorageWithFuncs(256, (*Effect).Update, (*Effect).Render)
	world.Items = scene.NewStorageWithFuncs(256, (*Item).Update, (*Item).Render)
	world.DebugShapes = scene.NewStorageWithFuncs(128, (*DebugShape).Update, (*DebugShape).Render)
	world.Cameras = scene.NewStorageWithFuncs(64, (*Camera).Update, nil)
	world.MapLayers = scene.NewStorageWithFuncs(3, world.UpdateMapLayer, (*comps.MapLayer).Render)

	te3File, err := te3.LoadTE3File(mapPath)
	if err != nil {
		return nil, err
	}

	_, world.GameMap, err = world.MapLayers.New()
	_, invisLayer, err2 := world.MapLayers.New()
	_, killLayer, err3 := world.MapLayers.New()
	if err = errors.Join(err, err2, err3); err != nil {
		return nil, err
	}

	const texFlagInvisible = "invisible"
	*world.GameMap, err = comps.NewMainMapLayer(te3File, ColLayerMap, []string{texFlagInvisible})
	if err != nil {
		return nil, err
	}

	*invisLayer = comps.NewExtraMapLayer(te3File, ColLayerInvisible)
	*killLayer = comps.NewExtraMapLayer(te3File, ColLayerKillzone)

	// Process tiles after mesh is generated.
	for id, tile := range te3File.Tiles.Data {
		if tile.ShapeID < 0 {
			continue
		}

		primaryTexture := cache.GetTexture(te3File.Tiles.Textures[tile.TextureIDs[0]])

		mapLayer := world.GameMap
		if primaryTexture.HasFlag(texFlagInvisible) {
			mapLayer = invisLayer
		} else if primaryTexture.HasFlag("killzone") {
			mapLayer = killLayer
		} else if primaryTexture.HasFlag("liquid") {
			// Remove collision from liquid tiles.
			continue
		}

		// Set collision shapes
		switch shapeName := te3File.Tiles.Shapes[tile.ShapeID]; shapeName {
		// case "assets/models/shapes/cylinder.obj":
		// Cylinder
		// mapLayer.GridShape.SetShapeAtFlatIndex(id, collision.NewCylinder(1.0, 2.0))
		case "assets/models/shapes/corner.obj",
			"assets/models/shapes/right_tetrahedron.obj",
			"assets/models/shapes/tetrahedron_transition.obj",
			"assets/models/shapes/wedge_corner_inner.obj",
			"assets/models/shapes/wedge_corner_outer.obj",
			"assets/models/shapes/wedge.obj",
			"assets/models/shapes/cylinder.obj":

			// Triangles

			// transform := tile.GetRotationMatrix()
			shapeMesh, err := cache.GetMesh(shapeName)
			//TODO: Index the cash using the transform

			if err != nil {
				log.Printf("error loading mesh for collisions shape of %v: %v\n", shapeName, err)
				continue
			}

			mapLayer.GridShape.SetShapeAtFlatIndex(id, collision.NewConvex(shapeMesh))
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
			mapLayer.GridShape.SetShapeAtFlatIndex(id, panelShape)
		default:
			// Box
			mapLayer.GridShape.SetShapeAtFlatIndex(id, collision.NewBox(math2.BoxFromRadius(1.0)))
		}
	}

	if levelFileName := path.Base(mapPath); len(levelFileName) >= 4 &&
		levelFileName[0] == 'e' &&
		levelFileName[1] >= '0' && levelFileName[1] <= '9' &&
		levelFileName[2] == 'm' &&
		levelFileName[3] >= '0' && levelFileName[3] <= '9' {

		// Level files starting with e#m# activate the level intro.
		world.Hud.Intro.Init(settings.Localize(levelFileName[0:4]+"Title"), strings.ToUpper(levelFileName[0:4]))
	} else {
		world.Hud.Intro.Init("", "")
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
					world.skyRender = comps.NewSkyRender(skyMesh, shaders.SkyShader, skyTex)
				}
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
			world.CurrentPlayer, _, err = SpawnPlayer(world, ent.Position, ent.Angles, world.CurrentCamera, changeInfo)
		}
		if err != nil {
			log.Printf("%v entity at %v caused an error: %v\n", entType, ent.GridPosition(), err)
		}
	}

	return world, nil
}

func (world *World) Update(deltaTime float32) {
	defer func() { world.tickCount++ }()

	world.removalQueue = world.removalQueue[0:0]

	//TODO: Could handle this in Actor.Update instead...
	if input.IsActionJustPressed(settings.ActionKillEnemies) {
		iter := world.IterActors()
		for actor, handle := iter.Next(); actor != nil; actor, handle = iter.Next() {
			if !handle.Equals(world.CurrentPlayer.Handle) {
				actor.Actor().Health = 0
			}
		}
	}

	// Free mouse
	if input.IsActionJustPressed(settings.ActionTrapMouse) {
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
	startTime := time.Now()

	// Create BSP tree
	it := world.IterBodies()
	world.bspTree = tree.BuildBspTree(scene.CollectSet(&it))

	it = world.IterBodies()
	for bodyEnt, _ := it.Next(); bodyEnt != nil; bodyEnt, _ = it.Next() {
		collidableBodies := slices.Collect(maps.Keys(world.bspTree.PotentiallyTouchingEnts(bodyEnt.Body().Transform.Position(), bodyEnt.Body().Shape)))

		movement := bodyEnt.Body().ResolveBodyCollisions(deltaTime, collidableBodies)

		// Sphere cast against the world.

		castShape := bodyEnt.Body().Shape.Inflate(-0.25)
		for range 2 {
			moveLen := movement.Len()
			minResult := collision.Result{Distance: moveLen}
			layerIt := world.MapLayers.Iter()
			for layer, _ := layerIt.Next(); layer != nil; layer, _ = layerIt.Next() {
				if (layer.Layer & bodyEnt.Body().Filter) != 0 {
					res := layer.GridShape.SweepAgainst(mgl32.Vec3{}, bodyEnt.Body().Transform.Position(), movement, castShape)
					if res.Hit && res.Distance < minResult.Distance {
						minResult = res
					}
				}
			}
			if moveLen > 0.0 {
				if minResult.Hit {
					bodyEnt.Body().Transform.TranslateV(movement.Mul((minResult.Distance - 0.25) / moveLen))
					// Slide
					canceledMove := minResult.Normal.Mul(-movement.Dot(minResult.Normal.Mul(1.0)))
					movement = movement.Add(canceledMove)
					fmt.Printf("Cast radisu was %v, minResult.distance: %v, normal: %v, canceledMove: %v\n", castShape.Radius(), minResult.Distance, minResult.Normal, canceledMove)
				} else {
					bodyEnt.Body().Transform.TranslateV(movement)
					break
				}
			}
		}
	}

	duration := time.Since(startTime).Milliseconds()
	if world.avgCollisionTime != 0 {
		world.avgCollisionTime = (world.avgCollisionTime + duration) / world.tickCount
	} else {
		world.avgCollisionTime = duration
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
			if layer.GridShape.OtherBodyTouches(mgl32.Vec3{}, actor.body.Transform.Position(), actor.body.Shape) {
				ent.OnDamage(layer, math2.Inf32())
			}
		}
	}
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
		scene.RenderStores(world, &renderContext)
		renderContext.RenderTranslucentObjects()
	}

	world.Hud.Debug.UpdateCounters(&renderContext, world.avgCollisionTime)
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

func (world *World) Raycast(rayOrigin, rayDir mgl32.Vec3, filter collision.Mask, maxDist float32, excludeBody comps.HasBody) (collision.Result, scene.Handle) {
	var rayBB math2.Box = math2.BoxFromPoints(rayOrigin, rayOrigin.Add(rayDir.Mul(maxDist)))
	var closestEnt scene.Handle
	var closestBodyHit collision.Result
	closestBodyHit.Distance = math.MaxFloat32
	iter := world.IterBodies()
	for bodyEnt, bodyId := iter.Next(); bodyEnt != nil; bodyEnt, bodyId = iter.Next() {
		body := bodyEnt.Body()
		if bodyEnt == excludeBody ||
			!bodyEnt.Body().OnLayer(filter) ||
			!body.Shape.Extents().Translate(body.Transform.Position()).Intersects(rayBB) {
			continue
		}
		bodyHit := body.Shape.Raycast(rayOrigin, rayDir, body.Transform.Position(), maxDist)
		if bodyHit.Hit && bodyHit.Distance < closestBodyHit.Distance {
			closestBodyHit = bodyHit
			closestEnt = bodyId
		}
	}
	if !closestEnt.IsNil() {
		return closestBodyHit, closestEnt
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
