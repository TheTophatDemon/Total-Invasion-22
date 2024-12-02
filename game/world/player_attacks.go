package world

import (
	"github.com/go-gl/mathgl/mgl32"
	"tophatdemon.com/total-invasion-ii/engine/assets/cache"
	"tophatdemon.com/total-invasion-ii/game/hud"
)

const (
	SFX_GRENADE      = "assets/sounds/weapon/grenadelaunch.wav"
	SFX_EGG_SHOOT    = "assets/sounds/weapon/chickengun.wav"
	SFX_PARUSU_SHOOT = "assets/sounds/weapon/parusu.wav"
)

func (player *Player) AttackWithWeapon() {
	weapon := player.world.Hud.SelectedWeapon()
	if weapon == nil {
		return
	}
	switch weapon.Order() {
	case hud.WEAPON_ORDER_SICKLE:
		firePos := mgl32.TransformCoordinate(mgl32.Vec3{0.0, 0.0, -0.5}, player.Body().Transform.Matrix())
		SpawnSickle(player.world, firePos, player.Body().Transform.Rotation(), player.id.Handle)
	case hud.WEAPON_ORDER_CHICKEN:
		firePos := mgl32.TransformCoordinate(mgl32.Vec3{0.0, -0.15, -0.5}, player.Body().Transform.Matrix())
		SpawnEgg(player.world, firePos, player.Body().Transform.Rotation(), player.id.Handle)
		cache.GetSfx(SFX_EGG_SHOOT).Play()
	case hud.WEAPON_ORDER_GRENADE:
		firePos := mgl32.TransformCoordinate(mgl32.Vec3{0.0, 0.15, -1.25}, player.Body().Transform.Matrix())
		SpawnGrenade(player.world, firePos, player.Body().Transform.Forward())
		cache.GetSfx(SFX_GRENADE).Play()
	case hud.WEAPON_ORDER_PARUSU:
		firePos := mgl32.TransformCoordinate(mgl32.Vec3{0.0, -0.25, -0.5}, player.Body().Transform.Matrix())
		SpawnPlasmaBall(player.world, firePos, player.Body().Transform.Rotation(), player.id.Handle)
		cache.GetSfx(SFX_PARUSU_SHOOT).Play()
	}
	player.noisyTimer = 0.5
}
