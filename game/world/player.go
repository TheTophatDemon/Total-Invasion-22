package world

import (
	"github.com/go-gl/mathgl/mgl32"
	"tophatdemon.com/total-invasion-ii/engine/color"
	"tophatdemon.com/total-invasion-ii/engine/input"
	"tophatdemon.com/total-invasion-ii/engine/math2"
	"tophatdemon.com/total-invasion-ii/engine/math2/collision"
	"tophatdemon.com/total-invasion-ii/engine/scene"
	"tophatdemon.com/total-invasion-ii/engine/scene/comps"

	"tophatdemon.com/total-invasion-ii/game/settings"
)

const (
	USE_DIST float32 = 3.0
)

type Player struct {
	Camera                                   comps.Camera
	RunSpeed, WalkSpeed                      float32
	StandFriction, WalkFriction, RunFriction float32
	id                                       scene.Id[*Player]
	actor                                    Actor
	world                                    *World
	weapons                                  [WEAPON_ORDER_MAX]Weapon
	selectedWeapon                           int
}

var _ HasActor = (*Player)(nil)
var _ comps.HasBody = (*Player)(nil)

func (p *Player) Actor() *Actor {
	return &p.actor
}

func (p *Player) Body() *comps.Body {
	return &p.actor.body
}

func SpawnPlayer(st *scene.Storage[Player], world *World, position, angles mgl32.Vec3) (id scene.Id[*Player], p *Player, err error) {
	id, p, err = st.New()
	if err != nil {
		return
	}
	p.id = id
	p.actor = Actor{
		body: comps.Body{
			Transform: comps.TransformFromTranslationAngles(
				position, angles,
			),
			Shape:     collision.NewSphere(0.7),
			Pushiness: 10,
			NoClip:    false,
		},
		YawAngle:  mgl32.DegToRad(angles[1]),
		AccelRate: 100.0,
		Friction:  20.0,
	}
	p.Camera = comps.NewCamera(
		70.0, settings.WINDOW_ASPECT_RATIO, 0.1, 1000.0,
	)
	p.RunSpeed = 12.0
	p.WalkSpeed = 7.0
	p.StandFriction = 80.0
	p.WalkFriction = 1.0
	p.RunFriction = 20.0
	p.world = world
	p.weapons = [...]Weapon{
		WEAPON_ORDER_SICKLE:      NewSickle(world, scene.Id[HasActor](p.id)),
		WEAPON_ORDER_CHICKEN:     {},
		WEAPON_ORDER_GRENADE:     {},
		WEAPON_ORDER_PARUSU:      {},
		WEAPON_ORDER_DBL_GRENADE: {},
		WEAPON_ORDER_SIGN:        {},
		WEAPON_ORDER_AIRHORN:     {},
	}
	p.selectedWeapon = WEAPON_ORDER_NONE
	p.EquipWeapon(WEAPON_ORDER_SICKLE)
	p.SelectWeapon(WEAPON_ORDER_SICKLE)

	return
}

func (p *Player) Update(deltaTime float32) {
	if input.IsActionPressed(settings.ACTION_FORWARD) {
		p.actor.inputForward = 1.0
	} else if input.IsActionPressed(settings.ACTION_BACK) {
		p.actor.inputForward = -1.0
	} else {
		p.actor.inputForward = 0.0
	}

	if input.IsActionPressed(settings.ACTION_RIGHT) {
		p.actor.inputStrafe = 1.0
	} else if input.IsActionPressed(settings.ACTION_LEFT) {
		p.actor.inputStrafe = -1.0
	} else {
		p.actor.inputStrafe = 0.0
	}

	if input.IsActionJustPressed(settings.ACTION_NOCLIP) {
		p.Body().NoClip = !p.Body().NoClip
		message := "No-Clip "
		if p.Body().NoClip {
			message += "Activated"
		} else {
			message += "Deactivated"
		}
		p.world.ShowMessage(message, 4.0, 100, color.Red)
	}

	if input.IsActionJustPressed(settings.ACTION_USE) {
		rayOrigin := p.Body().Transform.Position()
		rayDir := mgl32.TransformNormal(math2.Vec3Forward(), p.Body().Transform.Matrix())
		hit, closestBody := p.world.Raycast(rayOrigin, rayDir, true, USE_DIST, p)
		if hit.Hit && !closestBody.IsNil() {
			if usable, isUsable := scene.Get[Usable](closestBody); isUsable {
				usable.OnUse(p)
			}
		}
	}

	if input.IsActionJustPressed(settings.ACTION_FIRE) && p.selectedWeapon >= 0 {
		p.weapons[p.selectedWeapon].Fire()
	}

	if input.IsActionPressed(settings.ACTION_SLOW) {
		p.actor.MaxSpeed = p.WalkSpeed
	} else {
		p.actor.MaxSpeed = p.RunSpeed
	}

	if math2.Abs(p.actor.inputForward) > mgl32.Epsilon || math2.Abs(p.actor.inputStrafe) > mgl32.Epsilon {
		if p.actor.MaxSpeed == p.WalkSpeed {
			p.actor.Friction = p.WalkFriction
		} else {
			p.actor.Friction = p.RunFriction
		}
	} else {
		p.actor.Friction = p.StandFriction
	}

	p.actor.YawAngle -= input.ActionAxis(settings.ACTION_LOOK_HORZ)
	p.Body().Transform.SetRotation(0.0, p.actor.YawAngle, 0.0)

	p.actor.Update(deltaTime)
}

func (p *Player) ProcessSignal(s Signal, params any) {
	switch s {
	case SIGNAL_TELEPORTED:
		p.world.FlashScreen(color.Color{R: 1.0, G: 0.0, B: 1.0, A: 1.0}, 2.0)
	}
}

func (p *Player) SelectWeapon(order int) {
	if order == p.selectedWeapon || order >= len(p.weapons) || !p.weapons[order].equipped {
		return
	}
	if p.selectedWeapon >= 0 {
		p.weapons[p.selectedWeapon].Deselect()
	}
	p.selectedWeapon = order
	if p.selectedWeapon >= 0 {
		p.weapons[p.selectedWeapon].Select()
	}
}

func (p *Player) EquipWeapon(order int) {
	if order >= len(p.weapons) || order < 0 {
		return
	}
	p.weapons[order].Equip()
}
