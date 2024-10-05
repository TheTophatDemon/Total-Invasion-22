package hud

import "tophatdemon.com/total-invasion-ii/game"

type Weapon interface {
	Order() WeaponIndex
	Equip()
	IsEquipped() bool
	Select()
	Deselect()
	IsSelected() bool
	Fire(ammo *game.Ammo)
	CanFire(ammo *game.Ammo) bool
	Update(deltaTime float32, swayAmount float32, ammo *game.Ammo)
	AmmoType() game.AmmoType
	NoiseLevel() float32 // The distince within which an enemy can hear the weapon fire.
}
