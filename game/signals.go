package game

import "tophatdemon.com/total-invasion-ii/engine/scene/comps/ui"

type (
	MapChangeSignal struct {
		NextMapPath                                       string
		GiveAmmo                                          Ammo
		GiveArmor                                         ArmorType
		ArmorAmount                                       float32
		EquippedChicken, EquippedGrenade, EquippedParusu  bool
		EquippedDblGrenade, EquippedSign, EquippedAirhorn bool
		EquippedDefenestrator, EquippedCluckster          bool
	}
	ResumeGameSignal   struct{}
	ChangeScreenSignal struct {
		Screen ui.Screen
	}
	TeleportationSignal struct{}
)
