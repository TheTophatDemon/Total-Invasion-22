package world

import (
	"github.com/go-gl/mathgl/mgl32"
	"tophatdemon.com/total-invasion-ii/game/hud"
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
	}
	player.noisyTimer = 0.5
}
