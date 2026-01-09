package screens

import "tophatdemon.com/total-invasion-ii/game/settings"

type SettingsMenu struct {
	Menu
	changedSettings   settings.Data
	chooserScreenSize Chooser[[2]uint16]
}
