package world

import (
	"fmt"
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
	"tophatdemon.com/total-invasion-ii/game/settings"
)

type MovePhase uint8

const (
	MOVE_PHASE_CLOSED MovePhase = iota
	MOVE_PHASE_OPENING
	MOVE_PHASE_OPEN
	MOVE_PHASE_CLOSING
)

type ActivationType int8

const (
	ACTIVATOR_NONE ActivationType = iota - 1
	ACTIVATOR_ALL
	ACTIVATOR_KEY
	ACTIVATOR_TRIGGER
)

// A moving wall. Could be a door, a switch, or any other dynamic level geometry.
type Wall struct {
	MeshRender    comps.MeshRender
	AnimPlayer    comps.AnimationPlayer
	Origin        mgl32.Vec3 // The position in global space that the wall starts in.
	Destination   mgl32.Vec3 // The position in global space that the wall will move to.
	WaitTime      float32    // Time the ent remains at its destination position before moving back. If it's less than 0, it waits forever.
	Speed         float32
	body          comps.Body
	waitTimer     float32
	movePhase     MovePhase
	world         *World
	activator     ActivationType
	activateSound string
}

var _ Usable = (*Wall)(nil)

func SpawnWallFromTE3(st *scene.Storage[Wall], world *World, ent te3.Ent) (id scene.Id[*Wall], wall *Wall, err error) {
	id, wall, err = st.New()
	if err != nil {
		return
	}

	wall.world = world

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
			if err := wall.configureForDoor(ent); err != nil {
				return scene.Id[*Wall]{}, nil, err
			}
		default:
			wall.Destination = wall.Origin
		}
	}

	return
}

func (wall *Wall) configureForDoor(ent te3.Ent) error {
	// Determine the door's destination position
	unopenable, _ := ent.BoolProperty("unopenable")
	if !unopenable {
		dist, err := ent.FloatProperty("distance")
		if err != nil {
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
	} else {
		wall.Destination = wall.Origin
		wall.activator = ACTIVATOR_NONE
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
		blockers := wall.world.ActorsInSphere(wall.Origin, wall.body.Shape.Extents().LongestDimension(), nil)
		if len(blockers) > 0 {
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

func (wall *Wall) Body() *comps.Body {
	return &wall.body
}

func (wall *Wall) OnUse(player *Player) {
	switch wall.activator {
	case ACTIVATOR_NONE:
		wall.world.Hud.ShowMessage(settings.Localize("doorStuck"), 2.0, 10, color.Red)
	case ACTIVATOR_ALL:
		if !wall.Origin.ApproxEqual(wall.Destination) {
			switch wall.movePhase {
			case MOVE_PHASE_CLOSED:
				wall.movePhase = MOVE_PHASE_OPENING
				wall.waitTimer = 0
				if len(wall.activateSound) > 0 {
					cache.GetSfx(wall.activateSound).PlayAttenuatedV(wall.body.Transform.Position())
				}
			case MOVE_PHASE_OPEN:
				if wall.WaitTime >= 0.0 {
					wall.movePhase = MOVE_PHASE_CLOSING
				}
			}
		}
	case ACTIVATOR_KEY:
		wall.world.Hud.ShowMessage(settings.Localize("doorNeedKey"), 2.0, 10, color.Red)
	case ACTIVATOR_TRIGGER:
		wall.world.Hud.ShowMessage(settings.Localize("doorSwitch"), 2.0, 10, color.Red)
	}
}
