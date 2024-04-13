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
	weaponSickle                             WeaponSickle
	weaponChicken                            WeaponChicken
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
			Shape:  collision.NewSphere(0.7),
			Layer:  COL_LAYER_ACTORS,
			Filter: COL_FILTER_FOR_ACTORS,
			LockY:  true,
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

	// Initialize weapons
	p.weaponSickle = NewSickle(world, scene.Id[HasActor](p.id))
	p.weaponChicken = NewChickenCannon(world, scene.Id[HasActor](p.id))
	p.weapons = [...]Weapon{
		WEAPON_ORDER_SICKLE:      &p.weaponSickle,
		WEAPON_ORDER_CHICKEN:     &p.weaponChicken,
		WEAPON_ORDER_GRENADE:     nil,
		WEAPON_ORDER_PARUSU:      nil,
		WEAPON_ORDER_DBL_GRENADE: nil,
		WEAPON_ORDER_SIGN:        nil,
		WEAPON_ORDER_AIRHORN:     nil,
	}
	p.selectedWeapon = WEAPON_ORDER_NONE
	p.EquipWeapon(WEAPON_ORDER_SICKLE)
	p.EquipWeapon(WEAPON_ORDER_CHICKEN)
	p.SelectWeapon(WEAPON_ORDER_SICKLE)

	return
}

func (player *Player) Update(deltaTime float32) {
	if input.IsActionPressed(settings.ACTION_FORWARD) {
		player.actor.inputForward = 1.0
	} else if input.IsActionPressed(settings.ACTION_BACK) {
		player.actor.inputForward = -1.0
	} else {
		player.actor.inputForward = 0.0
	}

	if input.IsActionPressed(settings.ACTION_RIGHT) {
		player.actor.inputStrafe = 1.0
	} else if input.IsActionPressed(settings.ACTION_LEFT) {
		player.actor.inputStrafe = -1.0
	} else {
		player.actor.inputStrafe = 0.0
	}

	if input.IsActionJustPressed(settings.ACTION_NOCLIP) {
		message := "No-Clip "
		if player.Body().Layer != COL_LAYER_NONE {
			player.Body().Layer = COL_LAYER_NONE
			player.Body().Filter = COL_LAYER_NONE
			message += "Activated"
		} else {
			player.Body().Layer = COL_LAYER_ACTORS
			player.Body().Filter = COL_FILTER_FOR_ACTORS
			message += "Deactivated"
		}
		player.world.ShowMessage(message, 4.0, 100, color.Red)
	}

	if input.IsActionJustPressed(settings.ACTION_USE) {
		rayOrigin := player.Body().Transform.Position()
		rayDir := mgl32.TransformNormal(math2.Vec3Forward(), player.Body().Transform.Matrix())
		hit, closestBody := player.world.Raycast(rayOrigin, rayDir, COL_FILTER_FOR_ACTORS, USE_DIST, player)
		if hit.Hit && !closestBody.IsNil() {
			if usable, isUsable := scene.Get[Usable](closestBody); isUsable {
				usable.OnUse(player)
			}
		}
	}

	if input.IsActionJustPressed(settings.ACTION_SICKLE) {
		player.SelectWeapon(WEAPON_ORDER_SICKLE)
	} else if input.IsActionJustPressed(settings.ACTION_CHICKEN) {
		player.SelectWeapon(WEAPON_ORDER_CHICKEN)
	}

	if player.selectedWeapon >= 0 {
		var weapon Weapon = player.weapons[player.selectedWeapon]
		weapon.Update(deltaTime)
		if input.IsActionPressed(settings.ACTION_FIRE) && weapon.CanFire() {
			// Don't fire if there is a wall to close in front
			var cast collision.RaycastResult
			cast, _ = player.world.Raycast(player.Body().Transform.Position(), player.Body().Transform.Forward(), COL_LAYER_MAP, 1.5, player)
			if !cast.Hit {
				weapon.Fire()
			}
		}
	}

	if input.IsActionPressed(settings.ACTION_SLOW) {
		player.actor.MaxSpeed = player.WalkSpeed
	} else {
		player.actor.MaxSpeed = player.RunSpeed
	}

	if math2.Abs(player.actor.inputForward) > mgl32.Epsilon || math2.Abs(player.actor.inputStrafe) > mgl32.Epsilon {
		if player.actor.MaxSpeed == player.WalkSpeed {
			player.actor.Friction = player.WalkFriction
		} else {
			player.actor.Friction = player.RunFriction
		}
	} else {
		player.actor.Friction = player.StandFriction
	}

	player.actor.YawAngle -= input.ActionAxis(settings.ACTION_LOOK_HORZ)
	player.Body().Transform.SetRotation(0.0, player.actor.YawAngle, 0.0)

	player.actor.Update(deltaTime)
}

func (p *Player) ProcessSignal(s Signal, params any) {
	switch s {
	case SIGNAL_TELEPORTED:
		p.world.FlashScreen(color.Color{R: 1.0, G: 0.0, B: 1.0, A: 1.0}, 2.0)
	}
}

func (p *Player) SelectWeapon(order int) {
	if order == p.selectedWeapon || order >= len(p.weapons) || !p.weapons[order].Equipped() {
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
