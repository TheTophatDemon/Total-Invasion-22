package world

import (
	"math"

	"github.com/go-gl/mathgl/mgl32"
	"tophatdemon.com/total-invasion-ii/engine/color"
	"tophatdemon.com/total-invasion-ii/engine/input"
	"tophatdemon.com/total-invasion-ii/engine/math2"
	"tophatdemon.com/total-invasion-ii/engine/math2/collision"
	"tophatdemon.com/total-invasion-ii/engine/scene"
	"tophatdemon.com/total-invasion-ii/engine/scene/comps"

	"tophatdemon.com/total-invasion-ii/game/hud"
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
	selectedWeapon, nextWeapon               WeaponIndex
	initialCollisionLayers                   collision.Mask
}

var _ HasActor = (*Player)(nil)
var _ comps.HasBody = (*Player)(nil)

func (player *Player) Actor() *Actor {
	return &player.actor
}

func (player *Player) Body() *comps.Body {
	return &player.actor.body
}

func SpawnPlayer(st *scene.Storage[Player], world *World, position, angles mgl32.Vec3) (id scene.Id[*Player], p *Player, err error) {
	id, p, err = st.New()
	if err != nil {
		return
	}
	p.id = id
	p.initialCollisionLayers = COL_LAYER_ACTORS | COL_LAYER_PLAYERS
	p.actor = Actor{
		body: comps.Body{
			Transform: comps.TransformFromTranslationAngles(
				position, angles,
			),
			Shape:  collision.NewSphere(0.7),
			Layer:  p.initialCollisionLayers,
			Filter: COL_FILTER_FOR_ACTORS,
			LockY:  true,
		},
		YawAngle:  mgl32.DegToRad(angles[1]),
		AccelRate: 100.0,
		Friction:  20.0,
		Health:    100,
	}
	p.Camera = comps.NewCamera(
		70.0, settings.Current.WindowAspectRatio(), 0.1, 1000.0,
	)
	p.RunSpeed = 12.0
	p.WalkSpeed = 7.0
	p.StandFriction = 80.0
	p.WalkFriction = 1.0
	p.RunFriction = 20.0
	p.world = world

	// Initialize weapons
	p.weapons = [...]Weapon{
		WEAPON_ORDER_SICKLE:      NewSickle(world, scene.Id[HasActor](p.id)),
		WEAPON_ORDER_CHICKEN:     NewChickenCannon(world, scene.Id[HasActor](p.id)),
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
		var message string = settings.Localize("noclipActivate")
		if player.Body().Layer != COL_LAYER_NONE {
			player.Body().Layer = COL_LAYER_NONE
			player.Body().Filter = COL_LAYER_NONE
		} else {
			player.Body().Layer = player.initialCollisionLayers
			player.Body().Filter = COL_FILTER_FOR_ACTORS
			message = settings.Localize("noclipDeactivate")
		}
		player.world.Hud.ShowMessage(message, 4.0, 100, color.Red)
	}

	if input.IsActionJustPressed(settings.ACTION_USE) {
		rayOrigin := player.Body().Transform.Position()
		rayDir := player.Body().Transform.Forward()
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

	var weapon Weapon = nil
	if player.selectedWeapon >= 0 {
		weapon = player.weapons[player.selectedWeapon]
	}

	if weapon == nil || !weapon.IsSelected() {
		player.selectedWeapon = player.nextWeapon
		if player.selectedWeapon >= 0 {
			weapon = player.weapons[player.selectedWeapon]
			weapon.Select()
		}
	}

	if weapon != nil {
		weapon.Update(deltaTime, player.actor.body.Velocity.Len())

		if input.IsActionPressed(settings.ACTION_FIRE) && weapon.CanFire() {
			// Don't fire if there is a wall too close in front
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

	player.world.Hud.UpdatePlayerStats(deltaTime, hud.PlayerStats{
		Health: int(player.actor.Health),
		Noclip: player.Body().Layer == COL_LAYER_NONE,
	})
}

func (player *Player) ProcessSignal(s Signal, params any) {
	switch s {
	case SIGNAL_TELEPORTED:
		player.world.Hud.FlashScreen(color.Color{R: 1.0, G: 0.0, B: 1.0, A: 1.0}, 2.0)
	}
}

func (player *Player) SelectWeapon(order WeaponIndex) {
	if order == player.selectedWeapon || !player.weapons[order].IsEquipped() {
		return
	}
	if player.selectedWeapon >= 0 {
		player.weapons[player.selectedWeapon].Deselect()
	}
	player.nextWeapon = order
}

func (player *Player) EquipWeapon(order WeaponIndex) {
	if order < 0 {
		return
	}
	player.weapons[order].Equip()
}

func (player *Player) OnDamage(sourceEntity any, damage float32) {
	player.actor.Health = max(0, player.actor.Health-damage)
	player.world.Hud.FlashScreen(color.Red.WithAlpha(0.5), 1.0)

	if bodyHaver, ok := sourceEntity.(comps.HasBody); ok {
		dmgDir := bodyHaver.Body().Transform.Position().Sub(player.Body().Transform.Position())
		if dmgDir.LenSqr() > 0.0 {
			dmgDir = dmgDir.Normalize()
		}
		forward := player.actor.FacingVec()
		halfFov := mgl32.DegToRad(settings.Current.Fov / 2.0)
		if angleTo := math2.Acos(dmgDir.Dot(forward)); angleTo < halfFov || angleTo > math.Pi-halfFov {
			player.world.Hud.SuggestPlayerFace(hud.FaceStateHurtFront)
		} else if forward.Cross(dmgDir).Y() > 0.0 {
			player.world.Hud.SuggestPlayerFace(hud.FaceStateHurtLeft)
		} else {
			player.world.Hud.SuggestPlayerFace(hud.FaceStateHurtRight)
		}
	} else {
		player.world.Hud.SuggestPlayerFace(hud.FaceStateHurtFront)
	}
}
