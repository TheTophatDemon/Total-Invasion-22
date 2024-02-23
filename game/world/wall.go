package world

import (
	"fmt"
	"strings"

	"github.com/go-gl/mathgl/mgl32"
	"tophatdemon.com/total-invasion-ii/engine/assets/cache"
	"tophatdemon.com/total-invasion-ii/engine/assets/shaders"
	"tophatdemon.com/total-invasion-ii/engine/assets/te3"
	"tophatdemon.com/total-invasion-ii/engine/audio"
	"tophatdemon.com/total-invasion-ii/engine/color"
	"tophatdemon.com/total-invasion-ii/engine/math2"
	"tophatdemon.com/total-invasion-ii/engine/math2/collision"
	"tophatdemon.com/total-invasion-ii/engine/render"
	"tophatdemon.com/total-invasion-ii/engine/scene"
	"tophatdemon.com/total-invasion-ii/engine/scene/comps"
)

type MovePhase uint8

const (
	MOVE_PHASE_CLOSED MovePhase = iota
	MOVE_PHASE_OPENING
	MOVE_PHASE_OPEN
	MOVE_PHASE_CLOSING
)

type Activator int8

const (
	ACTIVATOR_NONE Activator = iota - 1
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
	activator     Activator
	activateSound *audio.Sfx
}

var _ Usable = (*Wall)(nil)

func SpawnWallFromTE3(st *scene.Storage[Wall], world *World, ent te3.Ent) (id scene.Id[Wall], wall *Wall, err error) {
	id, wall, err = st.New()
	if err != nil {
		return
	}

	wall.world = world

	transform := ent.Transform(false, false)

	if ent.Display != te3.ENT_DISPLAY_MODEL {
		return scene.Id[Wall]{}, nil, fmt.Errorf("te3 ent display mode should be 'model'")
	}

	var bbox math2.Box
	if len(ent.Model) > 0 {
		wall.MeshRender.Mesh, err = cache.GetMesh(ent.Model)
		if err != nil {
			return scene.Id[Wall]{}, nil, err
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
		Pushiness: 10_000,
		NoClip:    false,
	}

	if typ, ok := ent.Properties["type"]; !ok {
		return scene.Id[Wall]{}, nil, fmt.Errorf("no type property")
	} else {
		switch typ {
		case "door":
			if err := wall.configureForDoor(ent); err != nil {
				return scene.Id[Wall]{}, nil, err
			}
		default:
			wall.Destination = wall.Origin
		}
	}

	return
}

func (w *Wall) configureForDoor(ent te3.Ent) error {
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
			moveOffset = mgl32.TransformNormal(mgl32.Vec3{dist, 0.0, 0.0}, w.body.Transform.Matrix())
		case "left", "lf", "l":
			moveOffset = mgl32.TransformNormal(mgl32.Vec3{-dist, 0.0, 0.0}, w.body.Transform.Matrix())
		case "forward", "fw", "f":
			moveOffset = mgl32.TransformNormal(mgl32.Vec3{0.0, 0.0, -dist}, w.body.Transform.Matrix())
		case "backward", "back", "b":
			moveOffset = mgl32.TransformNormal(mgl32.Vec3{0.0, 0.0, dist}, w.body.Transform.Matrix())
		}
		w.Destination = w.Origin.Add(moveOffset)

		// Get waiting time
		if waitStr, ok := ent.Properties["wait"]; ok {
			if l := strings.ToLower(waitStr); l == "inf" || l == "infinity" || l == "-1" {
				w.WaitTime = -1.0
			} else if wait, err := ent.FloatProperty("wait"); err != nil {
				w.WaitTime = wait
			} else {
				w.WaitTime = 0.0
			}
		} else {
			w.WaitTime = 3.0
		}

		// Get speed
		if speed, err := ent.FloatProperty("speed"); err == nil {
			w.Speed = speed
		} else {
			w.Speed = 2.0
		}
	} else {
		w.Destination = w.Origin
		w.activator = ACTIVATOR_NONE
	}

	if sfxStr, ok := ent.Properties["activateSound"]; ok {
		if len(sfxStr) > 0 {
			w.activateSound, _ = cache.GetSfx("assets/sounds/" + sfxStr)
		} else {
			w.activateSound = nil
		}
	} else {
		w.activateSound, _ = cache.GetSfx("assets/sounds/opendoor.wav")
	}

	return nil
}

func (w *Wall) Update(deltaTime float32) {
	switch w.movePhase {
	case MOVE_PHASE_OPENING:
		targetDir := w.Destination.Sub(w.body.Transform.Position())
		targetDist := targetDir.Len()
		if targetDist <= w.Speed*deltaTime {
			w.body.Transform.SetPosition(w.Destination)
			w.movePhase = MOVE_PHASE_OPEN
			w.body.Velocity = mgl32.Vec3{}
		} else {
			w.body.Velocity = targetDir.Mul(w.Speed / targetDist)
		}
	case MOVE_PHASE_CLOSING:
		targetDir := w.Origin.Sub(w.body.Transform.Position())
		targetDist := targetDir.Len()
		// Detect if something is standing in the way
		blockers := w.world.ActorsInSphere(w.Origin, w.body.Shape.Extents().LongestDimension(), nil)
		if len(blockers) > 0 {
			w.body.Velocity = mgl32.Vec3{}
			w.movePhase = MOVE_PHASE_OPENING
		} else if targetDist <= w.Speed*deltaTime {
			w.body.Transform.SetPosition(w.Origin)
			w.movePhase = MOVE_PHASE_CLOSED
			w.body.Velocity = mgl32.Vec3{}
		} else {
			w.body.Velocity = targetDir.Mul(w.Speed / targetDist)
		}
	case MOVE_PHASE_OPEN:
		w.waitTimer += deltaTime
		if w.waitTimer > w.WaitTime && w.WaitTime >= 0.0 {
			w.movePhase = MOVE_PHASE_CLOSING
			w.waitTimer = 0.0
		}
		fallthrough
	case MOVE_PHASE_CLOSED:
		w.body.Velocity = mgl32.Vec3{}
	}

	w.body.Update(deltaTime)
}

func (w *Wall) Render(context *render.Context) {
	if !render.IsBoxVisible(context, w.Body().Shape.Extents().Translate(w.body.Transform.Position())) {
		return
	}
	context.DrawnWallCount++

	w.MeshRender.Render(&w.body.Transform, &w.AnimPlayer, context)
}

func (w *Wall) Body() *comps.Body {
	return &w.body
}

func (w *Wall) OnUse(player *Player) {
	switch w.activator {
	case ACTIVATOR_NONE:
		w.world.ShowMessage("The door can't move...", 2.0, 10, color.Red)
	case ACTIVATOR_ALL:
		if !w.Origin.ApproxEqual(w.Destination) {
			switch w.movePhase {
			case MOVE_PHASE_CLOSED:
				w.movePhase = MOVE_PHASE_OPENING
				w.waitTimer = 0
				if w.activateSound != nil {
					w.activateSound.Play()
				}
			case MOVE_PHASE_OPEN:
				if w.WaitTime >= 0.0 {
					w.movePhase = MOVE_PHASE_CLOSING
				}
			}
		}
	case ACTIVATOR_KEY:
		w.world.ShowMessage("I need a key...", 2.0, 10, color.Red)
	case ACTIVATOR_TRIGGER:
		w.world.ShowMessage("This door opens elsewhere...", 2.0, 10, color.Red)
	}
}
