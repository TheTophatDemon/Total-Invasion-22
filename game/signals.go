package game

import (
	"tophatdemon.com/total-invasion-ii/engine/assets/te3"
	"tophatdemon.com/total-invasion-ii/engine/scene/comps/ui"
)

type (
	MapChangeSignal struct {
		NextMapPath string
		PlayerEnt   *te3.Ent
	}
	ResumeGameSignal   struct{}
	ChangeScreenSignal struct {
		Screen ui.Screen
	}
	TeleportationSignal struct{}
)
