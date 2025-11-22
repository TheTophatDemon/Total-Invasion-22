package game

import "tophatdemon.com/total-invasion-ii/engine/scene/comps/ui/v2"

type (
	MapChangeSignal struct {
		NextMapPath     string
		GiveAmmo        Ammo
		GiveArmor       ArmorType
		ArmorAmount     float32
		EquippedWeapons [WeaponCount]bool
	}
	ResumeGameSignal   struct{}
	ChangeScreenSignal struct {
		Screen ui.Screen
	}
	TeleportationSignal struct{}
)
