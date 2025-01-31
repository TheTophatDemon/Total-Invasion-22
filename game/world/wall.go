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
	MOVE_PHASE_CLOSED MovePhase = iota
	MOVE_PHASE_OPENING
	MOVE_PHASE_OPEN
	MOVE_PHASE_CLOSING
)

type SwitchState uint8

const (
	NOT_A_SWITCH SwitchState = iota
	SWITCH_OFF
	SWITCH_ON
)

// A moving wall. Could be a door, a switch, or any other dynamic level geometry.
type Wall struct {
	MeshRender    comps.MeshRender
	AnimPlayer    comps.AnimationPlayer
	Origin        mgl32.Vec3 // The position in global space that the wall starts in.
	Destination   mgl32.Vec3 // The position in global space that the wall will move to.
	WaitTime      float32    // Time the ent remains at its destination position before moving back. If it's less than 0, it waits forever.
	Speed         float32
	id            scene.Id[*Wall]
	body          comps.Body
	waitTimer     float32
	movePhase     MovePhase
	world         *World
	unopenable    bool
	activateSound string
	key           game.KeyType
	linkNumber    int
	switchState   SwitchState
	blockUse      bool
}

var _ Usable = (*Wall)(nil)

func SpawnWallFromTE3(world *World, ent te3.Ent) (id scene.Id[*Wall], wall *Wall, err error) {
	id, wall, err = world.Walls.New()
	if err != nil {
		return
	}

	wall.world = world
	wall.id = id

	transform := comps.TransformFromTE3Ent(ent, false, false)

	if ent.Display != te3.ENT_DISPLAY_MODEL {
		return scene.Id[*Wall]{}, nil, fmt.Errorf("te3 ent display mode should be 'model'")
	}

	var bbox math2.Box
	if len(ent.Model) > 0 {
		wall.MeshRender.Mesh, err = cache.GetMesh(ent.Model)
		if err != nil {
			return scene.Id[*Wall]{}, nil, err
		}
		bbox = wall.MeshRender.Mesh.TransformedAABB(transform.Matrix().Mat3().Mat4())
		wall.MeshRender.Shader = shaders.MapShader
	} else {
		bbox = math2.BoxFromRadius(1.0)
	}

	if len(ent.Texture) > 0 {
		wall.MeshRender.Texture = cache.GetTexture(ent.Texture)
	}

	wall.Origin = ent.Position
	wall.body = comps.Body{
		Transform: transform,
		Shape:     collision.NewBox(bbox),
		Layer:     COL_LAYER_MAP,
		Filter:    COL_LAYER_NONE,
	}

	if typ, ok := ent.Properties["type"]; !ok {
		return scene.Id[*Wall]{}, nil, fmt.Errorf("no type property")
	} else {
		switch typ {
		case "door":
			err = wall.configureForDoor(ent)
		case "switch":
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

func (wall *Wall) configureForDoor(ent te3.Ent) error {
	// Determine the door's destination position
	unopenable, _ := ent.BoolProperty("unopenable")
	if !unopenable {
		dist, err := ent.FloatProperty("distance")
		if _, notFound := err.(te3.PropNotFoundError); notFound {
			dist = 1.8
		} else if err != nil {
			return err
		}

		dirStr, ok := ent.Properties["direction"]
		if !ok {
			return fmt.Errorf("need direction property")
		}

		var moveOffset mgl32.Vec3
		switch dirStr {
		case "down", "dn", "d":
			moveOffset = mgl32.Vec3{0.0, -dist, 0.0}
		case "up", "u":
			moveOffset = mgl32.Vec3{0.0, dist, 0.0}
		case "right", "rg", "r":
			moveOffset = mgl32.TransformNormal(mgl32.Vec3{dist, 0.0, 0.0}, wall.body.Transform.Matrix())
		case "left", "lf", "l":
			moveOffset = mgl32.TransformNormal(mgl32.Vec3{-dist, 0.0, 0.0}, wall.body.Transform.Matrix())
		case "forward", "fw", "f":
			moveOffset = mgl32.TransformNormal(mgl32.Vec3{0.0, 0.0, -dist}, wall.body.Transform.Matrix())
		case "backward", "back", "b":
			moveOffset = mgl32.TransformNormal(mgl32.Vec3{0.0, 0.0, dist}, wall.body.Transform.Matrix())
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

		if blockUse, err := ent.BoolProperty("blockUse"); err == nil {
			wall.blockUse = blockUse
		} else if _, notFound := err.(te3.PropNotFoundError); !notFound {
			return fmt.Errorf("could not parse blockuse property: %v", err)
		}

		if linkStr, ok := ent.Properties["link"]; ok {
			if linkNum, err := strconv.ParseInt(linkStr, 10, 32); err == nil {
				wall.linkNumber = int(linkNum)
			} else {
				return fmt.Errorf("could not parse link number; %v", err)
			}
		}
	} else {
		wall.Destination = wall.Origin
		wall.unopenable = true
	}

	if sfxStr, ok := ent.Properties["activateSound"]; ok {
		if len(sfxStr) > 0 {
			wall.activateSound = "assets/sounds/" + sfxStr
		}
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

	wall.switchState = SWITCH_OFF
	wall.Destination = wall.Origin
	wall.linkNumber, err = ent.IntProperty("link")
	if err != nil {
		return err
	}

	wall.AnimPlayer = comps.NewAnimationPlayer(wall.MeshRender.Texture.GetDefaultAnimation(), false)

	return nil
}

func SpawnInvisibleWall(
	world *World,
	position mgl32.Vec3,
	shape collision.Shape,
) (id scene.Id[*Wall], wall *Wall, err error) {
	id, wall, err = world.Walls.New()
	if err != nil {
		return
	}

	wall.world = world
	wall.Origin = position
	wall.body = comps.Body{
		Transform: comps.TransformFromTranslation(position),
		Shape:     shape,
		Layer:     COL_LAYER_INVISIBLE,
		Filter:    COL_LAYER_NONE,
	}

	return
}

func (wall *Wall) Update(deltaTime float32) {
	wall.AnimPlayer.Update(deltaTime)
	if wall.switchState != NOT_A_SWITCH && wall.AnimPlayer.HitATriggerFrame() {
		if wall.switchState == SWITCH_OFF {
			wall.switchState = SWITCH_ON
			wall.world.ActivateLinks(wall)
		} else {
			wall.switchState = SWITCH_OFF
			wall.world.DeactivateLinks(wall)
		}
	}

	switch wall.movePhase {
	case MOVE_PHASE_OPENING:
		targetDir := wall.Destination.Sub(wall.body.Transform.Position())
		targetDist := targetDir.Len()
		if targetDist <= wall.Speed*deltaTime {
			wall.body.Transform.SetPosition(wall.Destination)
			wall.movePhase = MOVE_PHASE_OPEN
			wall.body.Velocity = mgl32.Vec3{}
		} else {
			wall.body.Velocity = targetDir.Mul(wall.Speed / targetDist)
		}
	case MOVE_PHASE_CLOSING:
		targetDir := wall.Origin.Sub(wall.body.Transform.Position())
		targetDist := targetDir.Len()
		// Detect if something is standing in the way
		actorsIter := wall.world.IterActorsInSphere(wall.Origin, wall.body.Shape.Extents().LongestDimension(), nil)
		if actor, _ := actorsIter.Next(); actor != nil {
			wall.body.Velocity = mgl32.Vec3{}
			wall.movePhase = MOVE_PHASE_OPENING
		} else if targetDist <= wall.Speed*deltaTime {
			wall.body.Transform.SetPosition(wall.Origin)
			wall.movePhase = MOVE_PHASE_CLOSED
			wall.body.Velocity = mgl32.Vec3{}
		} else {
			wall.body.Velocity = targetDir.Mul(wall.Speed / targetDist)
		}
	case MOVE_PHASE_OPEN:
		wall.waitTimer += deltaTime
		if wall.waitTimer > wall.WaitTime && wall.WaitTime >= 0.0 {
			wall.movePhase = MOVE_PHASE_CLOSING
			wall.waitTimer = 0.0
		}
		fallthrough
	case MOVE_PHASE_CLOSED:
		wall.body.Velocity = mgl32.Vec3{}
	}
}

func (wall *Wall) Render(context *render.Context) {
	if !render.IsBoxVisible(context, wall.Body().Shape.Extents().Translate(wall.body.Transform.Position())) ||
		wall.MeshRender.Mesh == nil {
		return
	}
	context.DrawnWallCount++

	wall.MeshRender.Render(&wall.body.Transform, &wall.AnimPlayer, context)
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
	if wall.blockUse {
		return
	}
	switch true {
	case wall.switchState == SWITCH_OFF:
		anim, _ := wall.MeshRender.Texture.GetAnimation("on")
		wall.AnimPlayer.PlayNewAnim(anim)
		cache.GetSfx("assets/sounds/switch_on.wav").PlayAttenuatedV(wall.body.Transform.Position())
	case wall.switchState == SWITCH_ON:
		anim, _ := wall.MeshRender.Texture.GetAnimation("off")
		wall.AnimPlayer.PlayNewAnim(anim)
		cache.GetSfx("assets/sounds/switch_off.wav").PlayAttenuatedV(wall.body.Transform.Position())
	case wall.unopenable:
		wall.world.Hud.ShowMessage(settings.Localize("doorStuck"), 2.0, 10, color.Red)
	case wall.key != game.KEY_TYPE_INVALID && (player.keys&wall.key) != wall.key:
		// Locked if keycard not retrieved
		wall.world.Hud.ShowMessage(settings.Localize(game.KeycardNames[wall.key]+"KeyNeeded"), 2.0, 10, color.Red)
		cache.GetSfx("assets/sounds/door_locked.wav").PlayAttenuatedV(wall.body.Transform.Position())
	case wall.linkNumber != 0:
		// Door is opened by some other mechanism
		wall.world.Hud.ShowMessage(settings.Localize("doorSwitch"), 2.0, 10, color.Red)
	case !wall.Origin.ApproxEqual(wall.Destination):
		wall.ToggleMovement()
	}
}

func (wall *Wall) ToggleMovement() {
	switch wall.movePhase {
	case MOVE_PHASE_CLOSED:
		wall.Open()
	case MOVE_PHASE_OPEN:
		wall.Close()
	}
}

func (wall *Wall) Open() {
	wall.movePhase = MOVE_PHASE_OPENING
	wall.waitTimer = 0
	if len(wall.activateSound) > 0 {
		cache.GetSfx(wall.activateSound).PlayAttenuatedV(wall.body.Transform.Position())
	}
}

func (wall *Wall) Close() {
	if wall.WaitTime >= 0.0 {
		wall.movePhase = MOVE_PHASE_CLOSING
	}
}
