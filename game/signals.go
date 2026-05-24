package game

import "tophatdemon.com/total-invasion-ii/engine/scene/comps/ui"

type (
	MapChangeSignal struct {
		NextMapPath string
		Equipment   Equipment
	}
	ResumeGameSignal   struct{}
	ChangeScreenSignal struct {
		Screen ui.Screen
	}
	TeleportationSignal struct{}
)
