package world

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/go-gl/mathgl/mgl32"
	"tophatdemon.com/total-invasion-ii/engine/assets/cache"
	"tophatdemon.com/total-invasion-ii/engine/assets/shaders"
	"tophatdemon.com/total-invasion-ii/engine/assets/te3"
	"tophatdemon.com/total-invasion-ii/engine/color"
	"tophatdemon.com/total-invasion-ii/engine/math2"
	"tophatdemon.com/total-invasion-ii/engine/math2/collision"
	"tophatdemon.com/total-invasion-ii/engine/render"
	"tophatdemon.com/total-invasion-ii/engine/scene"
	"tophatdemon.com/total-invasion-ii/engine/scene/comps"
	"tophatdemon.com/total-invasion-ii/game"
	"tophatdemon.com/total-invasion-ii/game/settings"
)

type MovePhase uint8

const (
	MovePhaseClosed MovePhase = iota
	MovePhaseOpening
	MovePhaseOpen
	MovePhaseClosing
	MovePhaseCount
)

type SwitchState uint8

const (
	NotASwitch SwitchState = iota
	SwitchOff
	SwitchOn
)

type WallType string

const (
	WallTypeDoor     = "door"
	WallTypePushWall = "pushwall"
	WallTypeSwitch   = "switch"
)

// A moving wall. Could be a door, a switch, or any other dynamic level geometry.
type Wall struct {
	MeshRender         comps.MeshRender
	AnimPlayer         comps.AnimationPlayer
	Origin             mgl32.Vec3 // The position in global space that the wall starts in.
	Destination        mgl32.Vec3 // The position in global space that the wall will move to.
	WaitTime           float32    // Time the ent remains at its destination position before moving back. If it's less than 0, it waits forever.
	Speed              float32
	id                 scene.Id[*Wall]
	body               comps.Body
	movePhase          MovePhase
	switchState        SwitchState
	key                game.Keys
	activateSound      string
	activateMessageKey string
	disableUse         bool
	waitTimer          float32
	linkNumber         int
	proxiedWall        scene.Id[*Wall]
}

var _ Usable = (*Wall)(nil)

func SpawnWallFromTE3(ent te3.Ent) (id scene.Id[*Wall], wall *Wall, err error) {
	id, wall, err = gWorld.Walls.New()
	if err != nil {
		return
	}

	wall.id = id

	if ent.Display != te3.ENT_DISPLAY_MODEL {
		return scene.Id[*Wall]{}, nil, fmt.Errorf("te3 ent display mode should be 'model'")
	}

	var bbox math2.Box
	if len(ent.Model) > 0 {
		wall.MeshRender.Mesh, err = cache.GetMesh(ent.Model)
		if err != nil {
			return scene.Id[*Wall]{}, nil, err
		}
		transform := comps.TransformFromTE3Ent(ent, false, false)
		bbox = wall.MeshRender.Mesh.TransformedAABB(transform.Matrix().Mat3().Mat4())

		transform.SetPosition(0, 0, 0)
		wall.MeshRender.LocalTransform = transform
		wall.MeshRender.Shader = shaders.MapShader
	} else {
		bbox = math2.BoxFromRadius(1.0)
	}

	if len(ent.Texture) > 0 {
		wall.MeshRender.Texture = cache.GetTexture(ent.Texture)
	}

	wall.Origin = ent.Position
	wall.body = comps.Body{
		Position: ent.Position,
		Shape:    collision.NewBoxShape(bbox.Max[0], bbox.Max[1], bbox.Max[2]),
		Layers:   ColLayerMap,
	}

	if typ, ok := ent.Properties["type"]; !ok {
		return scene.Id[*Wall]{}, nil, fmt.Errorf("no type property")
	} else {
		switch strings.ToLower(typ) {
		case WallTypeDoor, WallTypePushWall:
			err = wall.configureForMover(ent)
		case WallTypeSwitch:
			err = wall.configureForSwitch(ent)
		default:
			wall.Destination = wall.Origin
		}
		if err != nil {
			return scene.Id[*Wall]{}, nil, err
		}
	}

	return
}

func (wall *Wall) configureForMover(ent te3.Ent) error {
	if wall == nil {
		return nil
	}

	wall.body.Layers |= ColLayerUsable

	isPushWall := strings.ToLower(ent.Properties["type"]) == WallTypePushWall

	// Determine the door's destination position
	unopenable, _ := ent.BoolProperty("unopenable")
	if !unopenable {
		dist, err := ent.FloatProperty("distance")
		if _, notFound := err.(te3.PropNotFoundError); notFound {
			if isPushWall {
				dist = 4.0
			} else {
				dist = 1.8
			}
		} else if err != nil {
			return err
		}

		dirStr, ok := ent.Properties["direction"]
		if !ok {
			if isPushWall {
				dirStr = "backward"
			} else {
				dirStr = "right"
			}
		}

		localTrans := wall.MeshRender.LocalTransform.Matrix()

		var moveOffset mgl32.Vec3
		switch dirStr {
		case "down", "dn", "d":
			moveOffset = mgl32.Vec3{0.0, -dist, 0.0}
		case "up", "u":
			moveOffset = mgl32.Vec3{0.0, dist, 0.0}
		case "right", "rg", "r":
			moveOffset = mgl32.TransformNormal(mgl32.Vec3{dist, 0.0, 0.0}, localTrans)
		case "left", "lf", "l":
			moveOffset = mgl32.TransformNormal(mgl32.Vec3{-dist, 0.0, 0.0}, localTrans)
		case "forward", "fw", "f":
			moveOffset = mgl32.TransformNormal(mgl32.Vec3{0.0, 0.0, -dist}, localTrans)
		case "backward", "back", "b":
			moveOffset = mgl32.TransformNormal(mgl32.Vec3{0.0, 0.0, dist}, localTrans)
		}
		wall.Destination = wall.Origin.Add(moveOffset)

		// Get waiting time
		if waitStr, ok := ent.Properties["wait"]; ok {
			if l := strings.ToLower(waitStr); l == "inf" || l == "infinity" || l == "-1" {
				wall.WaitTime = -1.0
			} else if wait, err := ent.FloatProperty("wait"); err != nil {
				wall.WaitTime = wait
			} else {
				wall.WaitTime = 0.0
			}
		} else if isPushWall {
			wall.WaitTime = -1.0
		} else {
			wall.WaitTime = 3.0
		}

		// Get speed
		if speed, err := ent.FloatProperty("speed"); err == nil {
			wall.Speed = speed
		} else {
			wall.Speed = 4.0
		}

		// Get key
		if keyName, ok := ent.Properties["key"]; ok {
			wall.key = game.KeyTypeFromName(keyName)
		}

		if linkStr, ok := ent.Properties["link"]; ok {
			if linkNum, err := strconv.ParseInt(linkStr, 10, 32); err == nil {
				wall.linkNumber = int(linkNum)
				if !isPushWall {
					wall.activateMessageKey = "doorSwitch"
				}
				wall.disableUse = true
			} else {
				return fmt.Errorf("could not parse link number; %v", err)
			}
		}

		if !isPushWall {
			_, _, err = SpawnProxyWall(wall)
			if err != nil {
				return fmt.Errorf("could not spawn proxy wall: %v", err)
			}
		} else {
			// Pushwalls set a map tile at their origin so that players don't get caught sliding along adjacent walls.
			wall.body.ExcludeLayers(ColLayerMap)
			gridX, gridY, gridZ := gWorld.GameMap.GridShape.WorldToGridPos(wall.Origin)
			gWorld.GameMap.GridShape.SetShapeAt(gridX, gridY, gridZ, wall.body.Shape, ColLayerMap)
		}
	} else {
		wall.Destination = wall.Origin
		wall.disableUse = true
		if !isPushWall {
			wall.activateMessageKey = "doorStuck"
		}
	}

	if sfxStr, ok := ent.Properties["activateSound"]; ok {
		if len(sfxStr) > 0 {
			wall.activateSound = "assets/sounds/" + sfxStr
		}
	} else if isPushWall {
		wall.activateSound = "assets/sounds/secretwall.wav"
	} else {
		wall.activateSound = "assets/sounds/opendoor.wav"
	}
	if len(wall.activateSound) > 0 {
		// Preload the sound
		cache.GetSfx(wall.activateSound)
	}

	return nil
}

func (wall *Wall) configureForSwitch(ent te3.Ent) error {
	var err error

	if wall == nil {
		return err
	}

	wall.body.Layers |= ColLayerUsable

	wall.switchState = SwitchOff
	wall.Destination = wall.Origin
	wall.linkNumber, err = ent.IntProperty("link")
	if err != nil {
		return err
	}

	wall.AnimPlayer = comps.NewAnimationPlayer(wall.MeshRender.Texture.GetDefaultAnimation(), false)

	return nil
}

// Proxy walls are for covering up the gap left by a door that is in the process of opening.
// This allows the player to hit the use key in the empty space to close the door
// and also prevents jittering caused by slipping through the narrow gap in the door as it opens.
func SpawnProxyWall(parentWall *Wall) (id scene.Id[*Wall], wall *Wall, err error) {
	if parentWall == nil {
		err = fmt.Errorf("parentWall shouldn't be nil")
		return
	}

	id, wall, err = gWorld.Walls.New()
	if err != nil {
		return
	}

	wall.id = id

	wall.Origin = parentWall.Origin
	wall.Destination = wall.Origin
	expandedBox := parentWall.body.Shape.Extents()
	expandedBox.Max = expandedBox.Max.Add(mgl32.Vec3{0.1, 0.1, 0.1})
	expandedBox.Min = expandedBox.Min.Sub(mgl32.Vec3{0.1, 0.1, 0.1})
	wall.body = comps.Body{
		Position: parentWall.body.Position,
		Shape:    collision.NewShapeFromBox(expandedBox),
		Layers:   ColLayerInvisible | ColLayerUsable,
	}
	wall.proxiedWall = parentWall.id

	return
}

func (wall *Wall) Update(deltaTime float32) {
	wall.AnimPlayer.Update(deltaTime)
	if wall.switchState != NotASwitch && wall.AnimPlayer.HitATriggerFrame() {
		if wall.switchState == SwitchOff {
			wall.switchState = SwitchOn
			gWorld.ActivateLinks(wall)
		} else {
			wall.switchState = SwitchOff
			gWorld.DeactivateLinks(wall)
		}
	}

	if proxiedWall, isProxy := wall.proxiedWall.Get(); isProxy {
		if proxiedWall.movePhase == MovePhaseClosed && wall.body.ExcludedLayers.On(ColLayerInvisible) {
			wall.body.RestoreLayers()
			proxiedWall.body.Layers |= ColLayerMap
		} else if proxiedWall.movePhase == MovePhaseOpen && !wall.body.ExcludedLayers.On(ColLayerInvisible) {
			proxiedWall.body.Layers &= (^ColLayerMap)
			wall.body.ExcludeLayers(ColLayerInvisible)
		}
	}

	switch wall.movePhase {
	case MovePhaseOpening:
		targetDir := wall.Destination.Sub(wall.body.Position)
		targetDist := targetDir.Len()
		if targetDist <= wall.Speed*deltaTime {
			wall.body.Position = wall.Destination
			wall.movePhase = MovePhaseOpen
			wall.body.Velocity = mgl32.Vec3{}
		} else {
			wall.body.Velocity = targetDir.Mul(wall.Speed / targetDist)
		}
	case MovePhaseClosing:
		targetDir := wall.Origin.Sub(wall.body.Position)
		targetDist := targetDir.Len()

		// Detect if something is standing in the way
		ents := gWorld.bspTree.PotentiallyTouchingEnts(wall.Origin, wall.body.Shape)
		obstructed := false
		for ent := range ents {
			if actorHaver, ok := scene.Get[HasActor](ent); ok {
				if actorHaver.Body().Shape.Touches(actorHaver.Body().Position, wall.body.Position, wall.body.Shape) {
					wall.body.Velocity = mgl32.Vec3{}
					wall.movePhase = MovePhaseOpening
					obstructed = true
					break
				}
			}
		}
		if obstructed {
			break
		}

		if targetDist <= wall.Speed*deltaTime {
			wall.body.Position = wall.Origin
			wall.movePhase = MovePhaseClosed
			wall.body.Velocity = mgl32.Vec3{}
		} else {
			wall.body.Velocity = targetDir.Mul(wall.Speed / targetDist)
		}
	case MovePhaseOpen:
		wall.waitTimer += deltaTime
		if wall.waitTimer > wall.WaitTime && wall.WaitTime >= 0.0 {
			wall.movePhase = MovePhaseClosing
			wall.waitTimer = 0.0
		}
		wall.body.Velocity = mgl32.Vec3{}
	case MovePhaseClosed:
		wall.body.Velocity = mgl32.Vec3{}
	}

	movement := wall.body.Velocity.Mul(deltaTime)
	wall.body.TranslateV(movement)
}

func (wall *Wall) Render(context *render.Context) {
	if !context.IsBoxVisible(wall.Body().Shape.Extents().Translate(wall.body.Position)) ||
		wall.MeshRender.Mesh == nil {
		return
	}
	context.DrawnWallCount++

	wall.MeshRender.Render(wall.body.Position, &wall.AnimPlayer, context)
}

func (wall *Wall) LinkNumber() int {
	return wall.linkNumber
}

func (wall *Wall) OnLinkActivate(source Linkable) {
	wall.Open()
}

func (wall *Wall) OnLinkDeactivate(source Linkable) {
	wall.Close()
}

func (wall *Wall) Handle() scene.Handle {
	return wall.id.Handle
}

func (wall *Wall) Body() *comps.Body {
	return &wall.body
}

func (wall *Wall) OnUse(player *Player) {
	if len(wall.activateMessageKey) > 0 {
		gWorld.Hud.ShowMessage(settings.Localize(wall.activateMessageKey), 10, color.Red)
	}
	if wall.disableUse {
		return
	}
	if proxiedWall, isProxy := wall.proxiedWall.Get(); isProxy {
		proxiedWall.OnUse(player)
		return
	}
	switch true {
	case wall.switchState == SwitchOff:
		anim, _ := wall.MeshRender.Texture.GetAnimation("on")
		wall.AnimPlayer.PlayNewAnim(anim)
		cache.GetSfx("assets/sounds/switch_on.wav").PlayAttenuatedV(wall.body.Position)
	case wall.switchState == SwitchOn:
		anim, _ := wall.MeshRender.Texture.GetAnimation("off")
		wall.AnimPlayer.PlayNewAnim(anim)
		cache.GetSfx("assets/sounds/switch_off.wav").PlayAttenuatedV(wall.body.Position)
	case wall.key != game.KeysNone && (player.keys&wall.key) != wall.key:
		// Locked if keycard not retrieved
		gWorld.Hud.ShowMessage(settings.Localize(wall.key.Name()+"KeyNeeded"), 10, color.Red)
		cache.GetSfx("assets/sounds/door_locked.wav").PlayAttenuatedV(wall.body.Position)
	case !wall.Origin.ApproxEqual(wall.Destination):
		wall.ToggleMovement()
	}
}

func (wall *Wall) ToggleMovement() {
	switch wall.movePhase {
	case MovePhaseClosed:
		wall.Open()
	case MovePhaseOpen:
		wall.Close()
	}
}

func (wall *Wall) Open() {
	wall.movePhase = MovePhaseOpening
	wall.waitTimer = 0
	if len(wall.activateSound) > 0 {
		cache.GetSfx(wall.activateSound).PlayAttenuatedV(wall.body.Position)
	}
	if wall.body.ExcludedLayers != 0 {
		wall.body.RestoreLayers()
		// Clear out any map cells that may be blocking the wall's area.
		gridX, gridY, gridZ := gWorld.GameMap.GridShape.WorldToGridPos(wall.Origin)
		gWorld.GameMap.GridShape.SetShapeAt(gridX, gridY, gridZ, collision.Shape{}, 0)
	}
}

func (wall *Wall) Close() {
	if wall.WaitTime >= 0.0 {
		wall.movePhase = MovePhaseClosing
		if len(wall.activateSound) > 0 {
			cache.GetSfx(wall.activateSound).PlayAttenuatedV(wall.body.Position)
		}
	}
}
